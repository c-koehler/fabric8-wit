package controller_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/fabric8-services/fabric8-wit/app"
	"github.com/fabric8-services/fabric8-wit/configuration"
	"github.com/fabric8-services/fabric8-wit/controller"
	"github.com/fabric8-services/fabric8-wit/kubernetesV1"
	"github.com/goadesign/goa"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/pkg/api/v1"
)

var spaceAttributeName = "spaceName"

type testKubeClient struct {
	closed bool
	// Don't implement methods we don't yet need
	kubernetesV1.KubeClientInterface
}

func (kc *testKubeClient) Close() {
	kc.closed = true
}

type testKubeClientGetter struct {
	client *testKubeClient
}

func (g *testKubeClientGetter) GetConfig() *configuration.Registry {
	return nil
}

func (g *testKubeClientGetter) GetKubeClient(ctx context.Context) (kubernetesV1.KubeClientInterface, error) {
	// Overwrites previous clients created by this getter
	g.client = &testKubeClient{}
	// Also return an error to avoid executing remainder of calling method
	return g.client, errors.New("Test")
}

func TestAPIMethodsCloseKube(t *testing.T) {
	testCases := []struct {
		name   string
		method func(*controller.AppsController) error
	}{
		{"SetDeployment", func(ctrl *controller.AppsController) error {
			count := 1
			ctx := &app.SetDeploymentAppsContext{
				PodCount: &count,
			}
			return ctrl.SetDeployment(ctx)
		}},
		{"ShowDeploymentStatSeries", func(ctrl *controller.AppsController) error {
			ctx := &app.ShowDeploymentStatSeriesAppsContext{}
			return ctrl.ShowDeploymentStatSeries(ctx)
		}},
		{"ShowDeploymentStats", func(ctrl *controller.AppsController) error {
			ctx := &app.ShowDeploymentStatsAppsContext{}
			return ctrl.ShowDeploymentStats(ctx)
		}},
		{"ShowEnvironment", func(ctrl *controller.AppsController) error {
			ctx := &app.ShowEnvironmentAppsContext{}
			return ctrl.ShowEnvironment(ctx)
		}},
		{"ShowSpace", func(ctrl *controller.AppsController) error {
			ctx := &app.ShowSpaceAppsContext{}
			return ctrl.ShowSpace(ctx)
		}},
		{"ShowSpaceApp", func(ctrl *controller.AppsController) error {
			ctx := &app.ShowSpaceAppAppsContext{}
			return ctrl.ShowSpaceApp(ctx)
		}},
		{"ShowSpaceAppDeployment", func(ctrl *controller.AppsController) error {
			ctx := &app.ShowSpaceAppDeploymentAppsContext{}
			return ctrl.ShowSpaceAppDeployment(ctx)
		}},
		{"ShowEnvAppPods", func(ctrl *controller.AppsController) error {
			ctx := &app.ShowEnvAppPodsAppsContext{}
			return ctrl.ShowEnvAppPods(ctx)
		}},
		{"ShowSpaceEnvironments", func(ctrl *controller.AppsController) error {
			ctx := &app.ShowSpaceEnvironmentsAppsContext{}
			return ctrl.ShowSpaceEnvironments(ctx)
		}},
	}
	// Check that each API method creating a KubeClientInterface also closes it
	getter := &testKubeClientGetter{}
	controller := &controller.AppsController{
		KubeClientGetter: getter,
	}
	for _, testCase := range testCases {
		err := testCase.method(controller)
		assert.Error(t, err, "Expected error \"Test\": "+testCase.name)
		// Check Close was called before returning
		assert.NotNil(t, getter.client, "No Kube client created: "+testCase.name)
		assert.True(t, getter.client.closed, "Kube client not closed: "+testCase.name)
	}
}

type ContextMock struct{
	valueResult interface{}
	context.Context
}

func (c ContextMock) Value(key interface{}) interface{} {
	return c.valueResult
}

type ResponseWriterMock struct {
	header     http.Header
	writeValue int
	writeError error
	http.ResponseWriter
}

func (r ResponseWriterMock) Header() http.Header {
	return r.header
}

func (r ResponseWriterMock) WriteHeader(int) {
}

func (r ResponseWriterMock) Write(b []byte) (int, error) {
	return r.writeValue, r.writeError
}

type KubeClientGetterMock struct {
	kubeClientInterface kubernetesV1.KubeClientInterface
	getKubeClientError  error
	configRegistry      *configuration.Registry
}

func (k KubeClientGetterMock) GetKubeClient(ctx context.Context) (kubernetesV1.KubeClientInterface, error) {
	return k.kubeClientInterface, k.getKubeClientError
}

func (k KubeClientGetterMock) GetConfig() *configuration.Registry {
	return k.configRegistry
}

type KubeClientMock struct {
	scaleDeploymentResult   *int
	scaleDeploymentError    error
	simpleStatSeries        *app.SimpleDeploymentStatSeriesV1
	simpleStatSeriesError   error
	deploymentStats         *app.SimpleDeploymentStatsV1
	deploymentStatsError    error
	simpleEnvironment       *app.SimpleEnvironmentV1
	simpleEnvironmentError  error
	simpleSpace             *app.SimpleSpaceV1
	simpleSpaceError        error
	simpleApp               *app.SimpleAppV1
	simpleAppError          error
	simpleDeployment        *app.SimpleDeploymentV1
	simpleDeploymentError   error
	envAppPods              []v1.Pod
	envAppPodsError         error
	environments            []*app.SimpleEnvironmentV1
	environmentsError       error
	kubernetesV1.KubeClientInterface
}

func (k KubeClientMock) ScaleDeployment(spaceName string, appName string, envName string, deployNumber int) (*int, error) {
	return k.scaleDeploymentResult, k.scaleDeploymentError
}

func (k KubeClientMock) GetDeploymentStatSeries(spaceName string, appName string, envName string, startTime time.Time,
	endTime time.Time, limit int) (*app.SimpleDeploymentStatSeriesV1, error) {
	return k.simpleStatSeries, k.simpleStatSeriesError
}

func (k KubeClientMock) GetDeploymentStats(spaceName string, appName string, envName string, startTime time.Time) (*app.SimpleDeploymentStatsV1, error) {
	return k.deploymentStats, k.deploymentStatsError
}

func (k KubeClientMock) GetEnvironment(envName string) (*app.SimpleEnvironmentV1, error) {
	return k.simpleEnvironment, k.simpleEnvironmentError
}

func (k KubeClientMock) GetSpace(spaceName string) (*app.SimpleSpaceV1, error) {
	return k.simpleSpace, k.simpleSpaceError
}

func (k KubeClientMock) GetApplication(spaceName string, appName string) (*app.SimpleAppV1, error) {
	return k.simpleApp, k.simpleAppError
}

func (k KubeClientMock) GetDeployment(spaceName string, appName string, envName string) (*app.SimpleDeploymentV1, error) {
	return k.simpleDeployment, k.simpleDeploymentError
}

func (k KubeClientMock) GetPodsInNamespace(nameSpace string, appName string) ([]v1.Pod, error) {
	return k.envAppPods, k.envAppPodsError
}

func (k KubeClientMock) GetEnvironments() ([]*app.SimpleEnvironmentV1, error) {
	return k.environments, k.environmentsError
}

func (k KubeClientMock) Close() {
}

type OSIOClientGetterMock struct {
	osioClient controller.OpenshiftIOClient
}

func (o OSIOClientGetterMock) GetAndCheckOSIOClient(ctx context.Context) controller.OpenshiftIOClient {
	return o.osioClient
}

type OSIOClientMock struct {
	namespaceAttributes *app.NamespaceAttributes
	namespaceTypeError  error
	userService         *app.UserService
	userServiceError    error
	space               *app.Space
	spaceError          error
}

func (o OSIOClientMock) GetNamespaceByType(ctx context.Context, userService *app.UserService, namespaceType string) (*app.NamespaceAttributes, error) {
	return o.namespaceAttributes, o.namespaceTypeError
}

func (o OSIOClientMock) GetUserServices(ctx context.Context) (*app.UserService, error) {
	return o.userService, o.userServiceError
}

func (o OSIOClientMock) GetSpaceByID(ctx context.Context, spaceID uuid.UUID) (*app.Space, error) {
	return o.space, o.spaceError
}

type ContextResponderMock struct {
	deploymentStatSeriesAppsError    error
	deploymentStatSeriesAppsVerifier func(res *app.SimpleDeploymentStatSeriesV1Single) error
	deploymentStatsError             error
	deploymentStatsVerifier          func(res *app.SimpleDeploymentStatsV1Single) error
	deploymentEnvironmentError       error
	deploymentEnvironmentVerifier    func(res *app.SimpleEnvironmentV1Single) error
	deploymentSpaceError             error
	deploymentSpaceVerifier          func(res *app.SimpleSpaceV1Single) error
	deploymentAppsError              error
	deploymentAppsVerifier           func(res *app.SimpleApplicationV1Single) error
	deploymentSimpleAppError         error
	deploymentSimpleAppVerifier      func(res *app.SimpleDeploymentV1Single) error
	deploymentsEnvPodsError          error
	deploymentsEnvPodsVerifier       func(res []byte) error
	deploymentsEnvironmentsError     error
	deploymentsEnvironmentsVerifier  func(res *app.SimpleEnvironmentV1List) error
}

func (c ContextResponderMock) SendShowDeploymentStatSeriesAppsOK(res *app.SimpleDeploymentStatSeriesV1Single, ctx *app.ShowDeploymentStatSeriesAppsContext) error {
	if c.deploymentStatSeriesAppsVerifier != nil {
		return c.deploymentStatSeriesAppsVerifier(res)
	}
	return c.deploymentStatSeriesAppsError
}

func (c ContextResponderMock) SendShowDeploymentStatsOK(res *app.SimpleDeploymentStatsV1Single, ctx *app.ShowDeploymentStatsAppsContext) error {
	if c.deploymentStatsVerifier != nil {
		return c.deploymentStatsVerifier(res)
	}
	return c.deploymentStatsError
}

func (c ContextResponderMock) SendEnvironmentSingleOK(res *app.SimpleEnvironmentV1Single, ctx *app.ShowEnvironmentAppsContext) error {
	if c.deploymentEnvironmentVerifier != nil {
		return c.deploymentEnvironmentVerifier(res)
	}
	return c.deploymentEnvironmentError
}

func (c ContextResponderMock) SendSpaceOK(res *app.SimpleSpaceV1Single, ctx *app.ShowSpaceAppsContext) error {
	if c.deploymentSpaceVerifier != nil {
		return c.deploymentSpaceVerifier(res)
	}
	return c.deploymentSpaceError
}

func (c ContextResponderMock) SendSimpleAppOK(res *app.SimpleApplicationV1Single, ctx *app.ShowSpaceAppAppsContext) error {
	if c.deploymentAppsVerifier != nil {
		return c.deploymentAppsVerifier(res)
	}
	return c.deploymentAppsError
}

func (c ContextResponderMock) SendSimpleDeploymentSingleOK(res *app.SimpleDeploymentV1Single, ctx *app.ShowSpaceAppDeploymentAppsContext) error {
	if c.deploymentSimpleAppVerifier != nil {
		return c.deploymentSimpleAppVerifier(res)
	}
	return c.deploymentSimpleAppError
}

func (c ContextResponderMock) SendEnvAppPodsOK(res []byte, ctx *app.ShowEnvAppPodsAppsContext) error {
	if c.deploymentsEnvPodsVerifier != nil {
		return c.deploymentsEnvPodsVerifier(res)
	}
	return c.deploymentsEnvPodsError
}

func (c ContextResponderMock) SendEnvironmentsOK(res *app.SimpleEnvironmentV1List, ctx *app.ShowSpaceEnvironmentsAppsContext) error {
	if c.deploymentsEnvironmentsVerifier != nil {
		return c.deploymentsEnvironmentsVerifier(res)
	}
	return c.deploymentsEnvironmentsError
}

func TestSetDeployment(t *testing.T) {
	testCases := []struct{
		testName         string
		deployCtx   	 *app.SetDeploymentAppsContext
		goaController    *goa.Controller
		registryConfig   *configuration.Registry
		kubeClientGetter controller.KubeClientGetter
		osioClientGetter controller.OSIOClientGetter
		shouldError      bool
	}{
		{
			testName: "Nil pod count should fail early with an error",
			deployCtx: &app.SetDeploymentAppsContext{},
			shouldError: true,
		},
		{
			testName: "Failure to get the kube client fails with an error",
			deployCtx: &app.SetDeploymentAppsContext{
				PodCount: new(int),
			},
			kubeClientGetter: KubeClientGetterMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Unable to get OSIO space",
			deployCtx: &app.SetDeploymentAppsContext{
				PodCount: new(int),
			},
			kubeClientGetter: KubeClientGetterMock{},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					spaceError: errors.New("no-space"),
				},
			},
			shouldError: true,
		},
		{
			testName: "Cannot write response",
			deployCtx: &app.SetDeploymentAppsContext{
				Context: ContextMock{
					valueResult: "someValue",
				},
				PodCount: new(int),
				DeployName: "deployName",
				ResponseData: &goa.ResponseData{
					ResponseWriter: ResponseWriterMock{
						header: map[string][]string{},
						writeError: errors.New("some-error"),
					},
				},
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					scaleDeploymentResult: new(int),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			shouldError: true,
		},
		{
			testName: "Can set deployments correctly",
			deployCtx: &app.SetDeploymentAppsContext{
				Context: ContextMock{
					valueResult: "someValue",
				},
				PodCount: new(int),
				DeployName: "deployName",
				ResponseData: &goa.ResponseData{
					ResponseWriter: ResponseWriterMock{
						header: map[string][]string{},
						writeValue: 0,
					},
				},
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					scaleDeploymentResult: new(int),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			shouldError: false,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Controller: testCase.goaController,
			Config: testCase.registryConfig,
			KubeClientGetter: testCase.kubeClientGetter,
			OSIOClientGetter: testCase.osioClientGetter,
		}

		err := appsController.SetDeployment(testCase.deployCtx)
		if testCase.shouldError {
			assert.NotNil(t, err, testCase.testName)
		} else {
			assert.Nil(t, err, testCase.testName)
		}
	}
}

func TestShowDeploymentStatSeries(t *testing.T) {
	startTime := new(float64)
	*startTime = 1.0
	endTime := new(float64)
	*endTime = 0.0

	testCases := []struct{
		testName          string
		deployCtx   	  *app.ShowDeploymentStatSeriesAppsContext
		registryConfig    *configuration.Registry
		kubeClientGetter  controller.KubeClientGetter
		osioClientGetter  controller.OSIOClientGetter
		contextResponder  controller.ContextResponder
		shouldError       bool
	}{
		{
			testName: "Does not allow bad timestamps",
			deployCtx: &app.ShowDeploymentStatSeriesAppsContext{
				Start: startTime,
				End: endTime,
			},
			shouldError: true,
		},
		{
			testName: "Errors if getting the KubeClient fails",
			deployCtx: &app.ShowDeploymentStatSeriesAppsContext{},
			kubeClientGetter: KubeClientGetterMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Errors if the OSIO client cannot get the space",
			deployCtx: &app.ShowDeploymentStatSeriesAppsContext{},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					scaleDeploymentResult: new(int),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					spaceError: errors.New("space-error"),
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting the stat series fails with an error",
			deployCtx: &app.ShowDeploymentStatSeriesAppsContext{},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleStatSeriesError: errors.New("some-error"),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting the stat series fails if the returned struct is nil",
			deployCtx: &app.ShowDeploymentStatSeriesAppsContext{},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleStatSeries: nil,
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			shouldError: true,
		},
		{
			testName: "Cannot write response",
			deployCtx: &app.ShowDeploymentStatSeriesAppsContext{
				Context: ContextMock{
				},
				ResponseData: &goa.ResponseData{
					Service: &goa.Service{
					},
					ResponseWriter: ResponseWriterMock{
						header: map[string][]string{},
					},
				},
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleStatSeries: &app.SimpleDeploymentStatSeriesV1{},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentStatSeriesAppsError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Successful response writing for deployment stat series",
			deployCtx: &app.ShowDeploymentStatSeriesAppsContext{
				ResponseData: &goa.ResponseData{
					Service: &goa.Service{
						Encoder: &goa.HTTPEncoder{},
					},
					ResponseWriter: ResponseWriterMock{
						header: map[string][]string{},
					},
				},
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleStatSeries: &app.SimpleDeploymentStatSeriesV1{
						Start: startTime,
						End: endTime,
					},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentStatSeriesAppsVerifier: func(res *app.SimpleDeploymentStatSeriesV1Single) error {
					if res.Data.Start == startTime && res.Data.End == endTime {
						return nil
					} else {
						return errors.New("expected mock object/data was not returned")
					}
				},
			},
			shouldError: false,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Config: testCase.registryConfig,
			KubeClientGetter: testCase.kubeClientGetter,
			OSIOClientGetter: testCase.osioClientGetter,
			ContextResponder: testCase.contextResponder,
		}

		err := appsController.ShowDeploymentStatSeries(testCase.deployCtx)
		if testCase.shouldError {
			assert.NotNil(t, err, testCase.testName)
		} else {
			assert.Nil(t, err, testCase.testName)
		}
	}
}

func TestShowDeploymentStats(t *testing.T) {
	mockTime := new(float64)
	*mockTime = 1.5
	mockValue := new(float64)
	*mockValue = 4.2

	testCases := []struct{
		testName          string
		deployCtx   	  *app.ShowDeploymentStatsAppsContext
		registryConfig    *configuration.Registry
		kubeClientGetter  controller.KubeClientGetter
		osioClientGetter  controller.OSIOClientGetter
		contextResponder  controller.ContextResponder
		shouldError       bool
	}{
		{
			testName:  "Errors if getting the KubeClient fails",
			deployCtx: &app.ShowDeploymentStatsAppsContext{},
			kubeClientGetter: KubeClientGetterMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Getting the stat series fails with an error",
			deployCtx: &app.ShowDeploymentStatsAppsContext{
				AppName: "appName",
				DeployName: "deployName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					deploymentStatsError: errors.New("some-error"),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting the stat series fails with an error when sending http OK",
			deployCtx: &app.ShowDeploymentStatsAppsContext{
				AppName: "appName",
				DeployName: "deployName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					deploymentStats: &app.SimpleDeploymentStatsV1{},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentStatsError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Successful response writing for deployment stats",
			deployCtx: &app.ShowDeploymentStatsAppsContext{
				AppName: "appName",
				DeployName: "deployName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					deploymentStats: &app.SimpleDeploymentStatsV1{
						Cores: &app.TimedNumberTupleV1{
							Time: mockTime,
							Value: mockValue,
						},
					},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentStatsVerifier: func(res *app.SimpleDeploymentStatsV1Single) error {
					if res.Data.Cores.Value == mockValue && res.Data.Cores.Time == mockTime {
						return nil
					} else {
						return errors.New("expected mock object/data was not returned")
					}
				},
			},
			shouldError: false,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Config: testCase.registryConfig,
			KubeClientGetter: testCase.kubeClientGetter,
			OSIOClientGetter: testCase.osioClientGetter,
			ContextResponder: testCase.contextResponder,
		}

		err := appsController.ShowDeploymentStats(testCase.deployCtx)
		if testCase.shouldError {
			assert.NotNil(t, err, testCase.testName)
		} else {
			assert.Nil(t, err, testCase.testName)
		}
	}
}

func TestShowEnvironment(t *testing.T) {
	mockEnvironmentName := new(string)
	*mockEnvironmentName = "mockEnvironment"

	testCases := []struct{
		testName          string
		deployCtx   	  *app.ShowEnvironmentAppsContext
		registryConfig    *configuration.Registry
		kubeClientGetter  controller.KubeClientGetter
		osioClientGetter  controller.OSIOClientGetter
		contextResponder  controller.ContextResponder
		shouldError       bool
	}{
		{
			testName: "Errors if getting the KubeClient fails",
			deployCtx: &app.ShowEnvironmentAppsContext{
				EnvName: "envName",
			},
			kubeClientGetter: KubeClientGetterMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Error occurs if getting environment fails",
			deployCtx: &app.ShowEnvironmentAppsContext{
				EnvName: "envName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleEnvironmentError: errors.New("some-error"),
				},
			},
			shouldError: true,
		},
		{
			testName: "Sending context fails",
			deployCtx: &app.ShowEnvironmentAppsContext{
				EnvName: "envName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleEnvironment: &app.SimpleEnvironmentV1{},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentEnvironmentError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Successful response writing for deployment environment",
			deployCtx: &app.ShowEnvironmentAppsContext{
				EnvName: "envName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleEnvironment: &app.SimpleEnvironmentV1{
						Name: mockEnvironmentName,
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentEnvironmentVerifier: func(res *app.SimpleEnvironmentV1Single) error {
					if res.Data.Name == mockEnvironmentName {
						return nil
					} else {
						return errors.New("expected mock object/data was not returned")
					}
				},
			},
			shouldError: false,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Config: testCase.registryConfig,
			KubeClientGetter: testCase.kubeClientGetter,
			OSIOClientGetter: testCase.osioClientGetter,
			ContextResponder: testCase.contextResponder,
		}

		err := appsController.ShowEnvironment(testCase.deployCtx)
		if testCase.shouldError {
			assert.NotNil(t, err, testCase.testName)
		} else {
			assert.Nil(t, err, testCase.testName)
		}
	}
}

func TestShowSpace(t *testing.T) {
	mockSpaceName := new(string)
	*mockSpaceName = "mockSpaceName"

	testCases := []struct{
		testName          string
		deployCtx   	  *app.ShowSpaceAppsContext
		registryConfig    *configuration.Registry
		kubeClientGetter  controller.KubeClientGetter
		osioClientGetter  controller.OSIOClientGetter
		contextResponder  controller.ContextResponder
		shouldError       bool
	}{
		{
			testName: "Errors if getting the KubeClient fails",
			deployCtx: &app.ShowSpaceAppsContext{
				//SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Getting the name of the space from kube fails with an error",
			deployCtx: &app.ShowSpaceAppsContext{
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					deploymentStatsError: errors.New("some-error"),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					spaceError: errors.New("some-error"),
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting the space fails with an error",
			deployCtx: &app.ShowSpaceAppsContext{
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleSpaceError: errors.New("some-error"),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			shouldError: true,
		},
		{
			testName: "Sending space by context fails",
			deployCtx: &app.ShowSpaceAppsContext{
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleSpace: &app.SimpleSpaceV1{},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentSpaceError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Sending space by context succeeds",
			deployCtx: &app.ShowSpaceAppsContext{
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleSpace: &app.SimpleSpaceV1{
						Name: mockSpaceName,
					},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentSpaceVerifier: func(res *app.SimpleSpaceV1Single) error {
					if res.Data.Name == mockSpaceName {
						return nil
					} else {
						return errors.New("expected mock object/data was not returned")
					}
				},
			},
			shouldError: false,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Config: testCase.registryConfig,
			KubeClientGetter: testCase.kubeClientGetter,
			OSIOClientGetter: testCase.osioClientGetter,
			ContextResponder: testCase.contextResponder,
		}

		err := appsController.ShowSpace(testCase.deployCtx)
		if testCase.shouldError {
			assert.NotNil(t, err, testCase.testName)
		} else {
			assert.Nil(t, err, testCase.testName)
		}
	}
}

func TestShowSpaceApp(t *testing.T) {
	mockAppName := new(string)
	*mockAppName = "mockAppName"

	testCases := []struct{
		testName          string
		deployCtx   	  *app.ShowSpaceAppAppsContext
		registryConfig    *configuration.Registry
		kubeClientGetter  controller.KubeClientGetter
		osioClientGetter  controller.OSIOClientGetter
		contextResponder  controller.ContextResponder
		shouldError       bool
	}{
		{
			testName: "Errors if getting the KubeClient fails",
			deployCtx: &app.ShowSpaceAppAppsContext{},
			kubeClientGetter: KubeClientGetterMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Getting the name of the space from kube fails with an error",
			deployCtx: &app.ShowSpaceAppAppsContext{
				AppName: "appName",
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					deploymentStatsError: errors.New("some-error"),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					spaceError: errors.New("some-error"),
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting the app fails with an error",
			deployCtx: &app.ShowSpaceAppAppsContext{
				AppName: "appName",
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleAppError: errors.New("some-error"),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting the app returns nil",
			deployCtx: &app.ShowSpaceAppAppsContext{
				AppName: "appName",
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			shouldError: true,
		},
		{
			testName: "Writing the app to the context fails",
			deployCtx: &app.ShowSpaceAppAppsContext{
				AppName: "appName",
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleApp: &app.SimpleAppV1{},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentAppsError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Writing the app to the context fails",
			deployCtx: &app.ShowSpaceAppAppsContext{
				AppName: "appName",
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleApp: &app.SimpleAppV1{
						Name: mockAppName,
					},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentAppsVerifier: func(res *app.SimpleApplicationV1Single) error {
					if res.Data.Name == mockAppName {
						return nil
					} else {
						return errors.New("expected mock object/data was not returned")
					}
				},
			},
			shouldError: false,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Config: testCase.registryConfig,
			KubeClientGetter: testCase.kubeClientGetter,
			OSIOClientGetter: testCase.osioClientGetter,
			ContextResponder: testCase.contextResponder,
		}

		err := appsController.ShowSpaceApp(testCase.deployCtx)
		if testCase.shouldError {
			assert.NotNil(t, err, testCase.testName)
		} else {
			assert.Nil(t, err, testCase.testName)
		}
	}
}

func TestShowSpaceAppDeployment(t *testing.T) {
	mockDeploymentName := new(string)
	*mockDeploymentName = "mockDeploymentName"

	testCases := []struct{
		testName          string
		deployCtx   	  *app.ShowSpaceAppDeploymentAppsContext
		registryConfig    *configuration.Registry
		kubeClientGetter  controller.KubeClientGetter
		osioClientGetter  controller.OSIOClientGetter
		contextResponder  controller.ContextResponder
		shouldError       bool
	}{
		{
			testName: "Errors if getting the KubeClient fails",
			deployCtx: &app.ShowSpaceAppDeploymentAppsContext{},
			kubeClientGetter: KubeClientGetterMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Getting the name of the space from kube fails with an error",
			deployCtx: &app.ShowSpaceAppDeploymentAppsContext{
				AppName: "appName",
				DeployName: "deployName",
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleDeploymentError: errors.New("some-error"),
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					spaceError: errors.New("some-error"),
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting a nil deployment fails with an error",
			deployCtx: &app.ShowSpaceAppDeploymentAppsContext{
				AppName: "appName",
				DeployName: "deployName",
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleDeployment: nil,
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			shouldError: true,
		},
		{
			testName: "Fails to write with the context",
			deployCtx: &app.ShowSpaceAppDeploymentAppsContext{
				AppName: "appName",
				DeployName: "deployName",
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleDeployment: &app.SimpleDeploymentV1{},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentSimpleAppError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Successfully write with the context",
			deployCtx: &app.ShowSpaceAppDeploymentAppsContext{
				AppName: "appName",
				DeployName: "deployName",
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					simpleDeployment: &app.SimpleDeploymentV1{
						Name: mockDeploymentName,
					},
				},
			},
			osioClientGetter: OSIOClientGetterMock{
				osioClient: OSIOClientMock{
					space: &app.Space{
						Attributes: &app.SpaceAttributes{
							Name: &spaceAttributeName,
						},
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentSimpleAppVerifier: func(res *app.SimpleDeploymentV1Single) error {
					if res.Data.Name == mockDeploymentName {
						return nil
					} else {
						return errors.New("expected mock object/data was not returned")
					}
				},
			},
			shouldError: false,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Config: testCase.registryConfig,
			KubeClientGetter: testCase.kubeClientGetter,
			OSIOClientGetter: testCase.osioClientGetter,
			ContextResponder: testCase.contextResponder,
		}

		err := appsController.ShowSpaceAppDeployment(testCase.deployCtx)
		if testCase.shouldError {
			assert.NotNil(t, err, testCase.testName)
		} else {
			assert.Nil(t, err, testCase.testName)
		}
	}
}

func TestShowEnvAppPods(t *testing.T) {
	expectedPods := []v1.Pod{ {} }
	expectedBytes, err := json.MarshalIndent(expectedPods, "", "  ")
	assert.NoError(t, err)
	expectedJson := "{\"pods\":" + string(expectedBytes) + "}\n"
	expectedByteResponse := []byte(expectedJson)

	testCases := []struct{
		testName          string
		deployCtx   	  *app.ShowEnvAppPodsAppsContext
		registryConfig    *configuration.Registry
		kubeClientGetter  controller.KubeClientGetter
		osioClientGetter  controller.OSIOClientGetter
		contextResponder  controller.ContextResponder
		shouldError       bool
	}{
		{
			testName: "Errors if getting the KubeClient fails",
			deployCtx: &app.ShowEnvAppPodsAppsContext{},
			kubeClientGetter: KubeClientGetterMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Getting pods from the KubeClient fails",
			deployCtx: &app.ShowEnvAppPodsAppsContext{
				AppName: "appName",
				EnvName: "envName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					envAppPodsError: errors.New("some-error"),
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting nil pods from the KubeClient yields a failure",
			deployCtx: &app.ShowEnvAppPodsAppsContext{
				AppName: "appName",
				EnvName: "envName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					envAppPods: nil,
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting zero pods back yields a failure",
			deployCtx: &app.ShowEnvAppPodsAppsContext{
				AppName: "appName",
				EnvName: "envName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					envAppPods: []v1.Pod{},
				},
			},
			shouldError: true,
		},
		{
			testName: "Sending by context with failure is an error",
			deployCtx: &app.ShowEnvAppPodsAppsContext{
				AppName: "appName",
				EnvName: "envName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					envAppPods: []v1.Pod{ {} },
				},
			},
			contextResponder: ContextResponderMock{
				deploymentsEnvPodsError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Successfully send from context",
			deployCtx: &app.ShowEnvAppPodsAppsContext{
				AppName: "appName",
				EnvName: "envName",
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					envAppPods: expectedPods,
				},
			},
			contextResponder: ContextResponderMock{
				deploymentsEnvPodsVerifier: func(res []byte) error {
					if bytes.Compare(res, expectedByteResponse) == 0 {
						return nil
					} else {
						return errors.New("expected mock object/data was not returned")
					}
				},
			},
			shouldError: false,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Config: testCase.registryConfig,
			KubeClientGetter: testCase.kubeClientGetter,
			OSIOClientGetter: testCase.osioClientGetter,
			ContextResponder: testCase.contextResponder,
		}

		err := appsController.ShowEnvAppPods(testCase.deployCtx)
		if testCase.shouldError {
			assert.NotNil(t, err, testCase.testName)
		} else {
			assert.Nil(t, err, testCase.testName)
		}
	}
}

func TestShowSpaceEnvironments(t *testing.T) {
	mockEnv := new(app.SimpleEnvironmentV1)

	testCases := []struct{
		testName          string
		deployCtx   	  *app.ShowSpaceEnvironmentsAppsContext
		registryConfig    *configuration.Registry
		kubeClientGetter  controller.KubeClientGetter
		osioClientGetter  controller.OSIOClientGetter
		contextResponder  controller.ContextResponder
		shouldError       bool
	}{
		{
			testName:  "Errors if getting the KubeClient fails",
			deployCtx: &app.ShowSpaceEnvironmentsAppsContext{},
			kubeClientGetter: KubeClientGetterMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Getting environments fails",
			deployCtx: &app.ShowSpaceEnvironmentsAppsContext{
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					environmentsError: errors.New("some-error"),
				},
			},
			shouldError: true,
		},
		{
			testName: "Getting environments as nil returns an error",
			deployCtx: &app.ShowSpaceEnvironmentsAppsContext{
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					environments: nil,
				},
			},
			shouldError: true,
		},
		{
			testName: "Sending by context with failure is an error",
			deployCtx: &app.ShowSpaceEnvironmentsAppsContext{
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					environments: []*app.SimpleEnvironmentV1{},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentsEnvironmentsError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Sending by context sucessfully causes no errors",
			deployCtx: &app.ShowSpaceEnvironmentsAppsContext{
				SpaceID: uuid.Nil,
			},
			kubeClientGetter: KubeClientGetterMock{
				kubeClientInterface: KubeClientMock{
					environments: []*app.SimpleEnvironmentV1{
						mockEnv,
					},
				},
			},
			contextResponder: ContextResponderMock{
				deploymentsEnvironmentsVerifier: func(res *app.SimpleEnvironmentV1List) error {
					if len(res.Data) == 1 && res.Data[0] == mockEnv {
						return nil
					} else {
						return errors.New("expected mock object/data was not returned")
					}
				},
			},
			shouldError: false,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Config: testCase.registryConfig,
			KubeClientGetter: testCase.kubeClientGetter,
			OSIOClientGetter: testCase.osioClientGetter,
			ContextResponder: testCase.contextResponder,
		}

		err := appsController.ShowSpaceEnvironments(testCase.deployCtx)
		if testCase.shouldError {
			assert.NotNil(t, err, testCase.testName)
		} else {
			assert.Nil(t, err, testCase.testName)
		}
	}
}
