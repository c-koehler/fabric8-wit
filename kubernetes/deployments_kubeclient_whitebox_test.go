package kubernetes

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	corev1 "k8s.io/client-go/kubernetes/typed/core/v1"
	"k8s.io/client-go/pkg/api/v1"
)

func TestGetPodStatus(t *testing.T) {
	const undefinedPodPhase = "someUndefinedPhase"

	config := &KubeClientConfig{
		ClusterURL:    "http://api.myCluster",
		BearerToken:   "myToken",
		UserNamespace: "myNamespace",
	}
	kubeClient, err := createKubeClient(config)
	assert.NotNil(t, err)

	testCases := []struct {
		pods             []*v1.Pod
		expectedResult   [][]string
		expectedPodTotal int
	}{
		{
			pods: []*v1.Pod{
				{
					Status: v1.PodStatus{
						Phase: v1.PodFailed,
					},
				},
			},
			expectedResult:   [][]string{{podRunning, "0"}},
			expectedPodTotal: 0,
		},
		{
			pods: []*v1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						DeletionTimestamp: &metav1.Time{},
					},
				},
			},
			expectedResult:   [][]string{{podTerminating, "1"}},
			expectedPodTotal: 1,
		},
		{
			pods: []*v1.Pod{
				{
					Status: v1.PodStatus{
						Phase: v1.PodPending,
					},
				},
			},
			expectedResult:   [][]string{{podWarning, "1"}},
			expectedPodTotal: 1,
		},
		{
			pods: []*v1.Pod{
				{
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
						ContainerStatuses: []v1.ContainerStatus{
							{
								State: v1.ContainerState{
									Waiting: &v1.ContainerStateWaiting{
										Reason: containerCrashLoop,
									},
								},
							},
						},
					},
				},
			},
			expectedResult:   [][]string{{podError, "1"}},
			expectedPodTotal: 1,
		},
		{
			pods: []*v1.Pod{
				{
					Status: v1.PodStatus{
						Phase: v1.PodPending,
						ContainerStatuses: []v1.ContainerStatus{
							{
								State: v1.ContainerState{
									Waiting: &v1.ContainerStateWaiting{
										Reason: containerCreating,
									},
								},
							},
						},
					},
				},
			},
			expectedResult:   [][]string{{podPulling, "1"}},
			expectedPodTotal: 1,
		},
		{
			pods: []*v1.Pod{
				{
					Spec: v1.PodSpec{
						Containers: []v1.Container{{}},
					},
					Status: v1.PodStatus{
						Phase: v1.PodRunning,
					},
				},
			},
			expectedResult:   [][]string{{podNotReady, "1"}},
			expectedPodTotal: 1,
		},
		{
			pods: []*v1.Pod{
				{
					Status: v1.PodStatus{
						Phase: v1.PodPhase(undefinedPodPhase),
					},
				},
			},
			expectedResult:   [][]string{{undefinedPodPhase, "1"}},
			expectedPodTotal: 1,
		},
	}

	for _, testCase := range testCases {
		result, podTotal := kubeClient.getPodStatus(testCase.pods)
		assert.Equal(t, testCase.expectedResult, result)
		assert.Equal(t, testCase.expectedPodTotal, podTotal)
	}
}

type testKubeMockService struct {
	corev1.ServiceInterface
	serviceList *v1.ServiceList
}

func (k testKubeMockService) List(opts metav1.ListOptions) (*v1.ServiceList, error) {
	return k.serviceList, nil
}

type testKubeMockServiceError struct {
	corev1.ServiceInterface
	expectedNamespace string
	actualNamespace   string
}

func (k testKubeMockServiceError) List(opts metav1.ListOptions) (*v1.ServiceList, error) {
	return nil, errors.New("expected namespace '" + k.expectedNamespace + "', got '" + k.actualNamespace + "'")
}

type testKubeRESTAPIMock struct {
	KubeRESTAPI
	kubeMockService   testKubeMockService
	expectedNamespace string
}

func (k testKubeRESTAPIMock) Services(namespace string) corev1.ServiceInterface {
	if namespace == k.expectedNamespace {
		return k.kubeMockService
	} else {
		return testKubeMockServiceError{
			expectedNamespace: k.expectedNamespace,
			actualNamespace:   namespace,
		}
	}
}

type kubeClientResourceProviderMock struct {
	resourceMap map[interface{}]interface{}
}

func (k kubeClientResourceProviderMock) GetResource(clusterUrl string, bearerToken string, url string, allowMissing bool) (map[interface{}]interface{}, error) {
	return k.resourceMap, nil
}

func TestGetBestRoute(t *testing.T) {
	const targetHost = "someHost"
	const targetNamespace = "namespace"

	deployment := deployment{
		dcUID:      "dcUID",
		appVersion: "appVersion",
		current: &v1.ReplicationController{
			Spec: v1.ReplicationControllerSpec{
				Template: &v1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{
						Labels: map[string]string{"selectorKey": "selectorValue"},
					},
				},
			},
		},
	}
	kubeMockService := testKubeMockService{
		serviceList: &v1.ServiceList{
			Items: []v1.Service{
				{
					Spec: v1.ServiceSpec{
						Selector: map[string]string{"selectorKey": "selectorValue"},
					},
					ObjectMeta: metav1.ObjectMeta{
						Name: "serviceName",
					},
				},
			},
		},
	}
	kubeRESTAPIMock := testKubeRESTAPIMock{
		kubeMockService:   kubeMockService,
		expectedNamespace: targetNamespace,
	}
	mockRouteMap := []interface{}{
		map[interface{}]interface{}{
			"spec": map[interface{}]interface{}{
				"to": map[interface{}]interface{}{
					"name": "serviceName",
				},
			},
			"status": map[interface{}]interface{}{
				"ingress": []interface{}{
					map[interface{}]interface{}{
						"conditions": []interface{}{
							map[interface{}]interface{}{
								"type":               "Admitted",
								"status":             "True",
								"lastTransitionTime": "2015-12-02T21:01:23+00:00",
							},
							map[interface{}]interface{}{
								"type":               "Admitted",
								"status":             "True",
								"lastTransitionTime": "2014-01-03T05:05:53+00:00",
							},
						},
						"host": targetHost,
					},
				},
			},
		},
	}
	kubeClient := kubeClient{
		config: &KubeClientConfig{
			ClusterURL:    "http://api.myCluster",
			BearerToken:   "myToken",
			UserNamespace: "myNamespace",
		},
		KubeRESTAPI: kubeRESTAPIMock,
		KubeClientResourceProvider: kubeClientResourceProviderMock{
			resourceMap: map[interface{}]interface{}{
				"kind":  "RouteList",
				"items": mockRouteMap,
			},
		},
	}

	url, err := kubeClient.getBestRoute(targetNamespace, &deployment)
	assert.Nil(t, err)
	if err == nil {
		assert.Equal(t, "http", url.Scheme)
		assert.Equal(t, targetHost, url.Host)
	}
}
