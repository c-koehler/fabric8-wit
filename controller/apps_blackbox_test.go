package controller_test

import (
	"context"
	"testing"

	"github.com/fabric8-services/fabric8-wit/app"
	"github.com/fabric8-services/fabric8-wit/configuration"
	"github.com/fabric8-services/fabric8-wit/controller"
	"github.com/fabric8-services/fabric8-wit/kubernetesV1"
	"github.com/goadesign/goa"
	"github.com/stretchr/testify/assert"
)

type KubeClientProviderMock struct {
	kubeClientInterface kubernetesV1.KubeClientInterface
	getKubeClientError  error
	configRegistry      *configuration.Registry
}

func (k KubeClientProviderMock) GetKubeClient(ctx context.Context) (kubernetesV1.KubeClientInterface, error) {
	return k.kubeClientInterface, k.getKubeClientError
}

func (k KubeClientProviderMock) GetConfig() *configuration.Registry {
	return k.configRegistry
}

type OSIOClientV1ProviderMock struct {
	osioClientV1 *controller.OSIOClientV1
}

func (o OSIOClientV1ProviderMock) GetAndCheckOSIOClientV1(ctx context.Context) *controller.OSIOClientV1 {
	return o.osioClientV1
}

func TestSetDeployment(t *testing.T) {
	testCases := []struct{
		testName             string
		deployCtx   	     *app.SetDeploymentAppsContext
		goaController        *goa.Controller
		registryConfig       *configuration.Registry
		kubeClientProvider   controller.KubeClientProvider
		osioClientV1Provider controller.OSIOClientV1Provider
		shouldError          bool
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
			kubeClientProvider: KubeClientProviderMock{
				getKubeClientError: errors.New("some-error"),
			},
			shouldError: true,
		},
		{
			testName: "Unable to get OSIO space",
			deployCtx: &app.SetDeploymentAppsContext{
				PodCount: new(int),
			},
			kubeClientProvider: KubeClientProviderMock{},
			osioClientV1Provider: OSIOClientV1ProviderMock{},
			shouldError: true,
		},
	}

	for _, testCase := range testCases {
		appsController := &controller.AppsController{
			Controller: testCase.goaController,
			Config: testCase.registryConfig,
			KubeClientProvider: testCase.kubeClientProvider,
			OSIOClientV1Provider: testCase.osioClientV1Provider,
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
	// TODO: stuff
}

func TestShowDeploymentStats(t *testing.T) {
	// TODO: stuff
}

func TestShowEnvironment(t *testing.T) {
	// TODO: stuff
}

func TestShowSpace(t *testing.T) {
	// TODO: stuff
}

func TestShowSpaceApp(t *testing.T) {
	// TODO: stuff
}

func TestShowSpaceAppDeployment(t *testing.T) {
	// TODO: stuff
}

func TestShowEnvAppPods(t *testing.T) {
	// TODO: stuff
}

func TestShowSpaceEnvironments(t *testing.T) {
	// TODO: stuff
}
