package controller_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"testing"

	"github.com/fabric8-services/fabric8-wit/app"
	"github.com/fabric8-services/fabric8-wit/controller"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/require"
)

const (
	deploymentsOsioTestFilePath = "test-files/deployments_osio/"
)

func readJsonFromTestFile(t *testing.T, testFilePath string) *bytes.Buffer {
	data, err := ioutil.ReadFile(testFilePath)
	require.NoError(t, err, "Unable to read test file data from: " + testFilePath)
	return bytes.NewBuffer(data)
}

// Structs and interfaces for mocking/testing
type MockContext struct {
	context.Context
}

type JsonResponseReader struct {
	jsonBytes *bytes.Buffer
}

func (r *JsonResponseReader) ReadResponse(resp *http.Response) ([]byte, error) {
	return r.jsonBytes.Bytes(), nil
}

type MockResponseBodyReader struct {
	io.ReadCloser
}

func (m MockResponseBodyReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (m MockResponseBodyReader) Close() error {
	return nil
}

type MockWitClient struct {
	SpaceHttpResponse            *http.Response
	SpaceHttpResponseError       error
	UserServiceHttpResponse      *http.Response
	UserServiceHttpResponseError error
}

func (m *MockWitClient) ShowSpace(ctx context.Context, path string, ifModifiedSince *string, ifNoneMatch *string) (*http.Response, error) {
	return m.SpaceHttpResponse, m.SpaceHttpResponseError
}

func (m *MockWitClient) ShowUserService(ctx context.Context, path string) (*http.Response, error) {
	return m.UserServiceHttpResponse, m.UserServiceHttpResponseError
}

// Unit tests
func TestGetUserServicesWithShowUserServiceError(t *testing.T) {
	mockWitClient := &MockWitClient{
		UserServiceHttpResponseError: errors.New("error"),
	}
	mockOSIOClient := controller.CreateOSIOClient(mockWitClient, &controller.IOResponseReader{})

	_, err := mockOSIOClient.GetUserServices(&MockContext{})
	require.Error(t, err)
}

func TestGetUserServicesBadStatusCodes(t *testing.T) {
	testCases := []struct {
		statusCode  int
		shouldBeNil bool
	}{
		{http.StatusMovedPermanently, false},
		{http.StatusNotFound, true},
		{http.StatusInternalServerError, false},
	}

	for _, testCase := range testCases {
		mockResponse := &http.Response{
			Body:       &MockResponseBodyReader{},
			StatusCode: testCase.statusCode,
		}
		mockWitClient := &MockWitClient{
			UserServiceHttpResponse: mockResponse,
		}
		mockOSIOClient := controller.CreateOSIOClient(mockWitClient, &controller.IOResponseReader{})

		userService, err := mockOSIOClient.GetUserServices(&MockContext{})
		if testCase.shouldBeNil {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
		require.Nil(t, userService)
	}
}

func TestGetUserServiceWithMalformedJSON(t *testing.T) {
	jsonReader := &JsonResponseReader{
		jsonBytes: bytes.NewBuffer([]byte(`{`)),
	}
	mockResponse := &http.Response{
		Body:       &MockResponseBodyReader{},
		StatusCode: http.StatusOK,
	}
	mockWitClient := &MockWitClient{
		UserServiceHttpResponse: mockResponse,
	}
	mockOSIOClient := controller.CreateOSIOClient(mockWitClient, jsonReader)

	_, err := mockOSIOClient.GetUserServices(&MockContext{})
	require.Error(t, err)
}

func TestUserServiceWithProperJSON(t *testing.T) {
	jsonReader := &JsonResponseReader{
		jsonBytes: readJsonFromTestFile(t, deploymentsOsioTestFilePath + "user-service.json"),
	}
	mockResponse := &http.Response{
		Body:       &MockResponseBodyReader{},
		StatusCode: http.StatusOK,
	}
	mockWitClient := &MockWitClient{
		UserServiceHttpResponse: mockResponse,
	}
	mockOSIOClient := controller.CreateOSIOClient(mockWitClient, jsonReader)

	userService, err := mockOSIOClient.GetUserServices(&MockContext{})
	require.NoError(t, err)
	require.NotNil(t, userService)
	require.Equal(t, "https://auth.openshift.io/api/users/77777777-7777-7777-7777-777777777777", *userService.Links.Related)
	require.Equal(t, "https://auth.openshift.io/api/users/88888888-8888-8888-8888-888888888888", *userService.Links.Self)
	require.Equal(t, "66666666-6666-6666-6666-666666666666", userService.ID.String())
	require.Equal(t, "identities", userService.Type)
}

func TestGetSpaceByIDWithShowSpaceError(t *testing.T) {
	mockContext := &MockContext{}
	mockWitClient := &MockWitClient{
		SpaceHttpResponseError: errors.New("error"),
	}
	mockOSIOClient := controller.CreateOSIOClient(mockWitClient, &controller.IOResponseReader{})

	_, err := mockOSIOClient.GetSpaceByID(mockContext, uuid.Nil)
	require.Error(t, err)
}

func TestGetSpaceByIDBadStatusCode(t *testing.T) {
	testCases := []struct {
		statusCode  int
		shouldBeNil bool
	}{
		{http.StatusMovedPermanently, false},
		{http.StatusNotFound, true},
		{http.StatusInternalServerError, false},
	}

	for _, testCase := range testCases {
		mockResponse := &http.Response{
			Body:       &MockResponseBodyReader{},
			StatusCode: testCase.statusCode,
		}
		mockWitClient := &MockWitClient{
			SpaceHttpResponse: mockResponse,
		}
		mockOSIOClient := controller.CreateOSIOClient(mockWitClient, &controller.IOResponseReader{})

		userService, err := mockOSIOClient.GetSpaceByID(&MockContext{}, uuid.Nil)
		if testCase.shouldBeNil {
			require.NoError(t, err)
		} else {
			require.Error(t, err)
		}
		require.Nil(t, userService)
	}
}

func TestGetSpaceByIDWithMalformedJSON(t *testing.T) {
	jsonReader := &JsonResponseReader{
		jsonBytes: bytes.NewBuffer([]byte(`{`)),
	}
	mockResponse := &http.Response{
		Body:       &MockResponseBodyReader{},
		StatusCode: http.StatusOK,
	}
	mockWitClient := &MockWitClient{
		SpaceHttpResponse: mockResponse,
	}
	mockOSIOClient := controller.CreateOSIOClient(mockWitClient, jsonReader)

	_, err := mockOSIOClient.GetSpaceByID(&MockContext{}, uuid.Nil)
	require.Error(t, err)
}

func TestGetSpaceByIDWithProperJSON(t *testing.T) {
	jsonReader := &JsonResponseReader{
		jsonBytes: readJsonFromTestFile(t, deploymentsOsioTestFilePath + "space-id.json"),
	}
	mockContext := &MockContext{}
	mockResponse := &http.Response{
		Body:       &MockResponseBodyReader{},
		StatusCode: http.StatusOK,
	}
	mockWitClient := &MockWitClient{
		SpaceHttpResponse: mockResponse,
	}
	mockOSIOClient := controller.CreateOSIOClient(mockWitClient, jsonReader)

	space, err := mockOSIOClient.GetSpaceByID(mockContext, uuid.Nil)
	require.NoError(t, err)
	require.NotNil(t, space)
	require.Equal(t, "https://api.openshift.io/api/spaces/00000000-0000-0000-0000-000000000000", *space.Links.Self)
	require.Equal(t, "https://api.openshift.io/api/spaces/00000000-0000-0000-0000-000000000000", *space.Links.Related)
	require.Equal(t, "https://api.openshift.io/api/spaces/00000000-0000-0000-0000-000000000000/backlog", *space.Links.Backlog.Self)
	require.Equal(t, 0, space.Links.Backlog.Meta.TotalCount)
	require.Equal(t, "https://api.openshift.io/api/filters", *space.Links.Filters)
	require.Equal(t, "https://api.openshift.io/api/spaces/00000000-0000-0000-0000-000000000000/workitemlinktypes", *space.Links.Workitemlinktypes)
	require.Equal(t, "https://api.openshift.io/api/spaces/00000000-0000-0000-0000-000000000000/workitemtypes", *space.Links.Workitemtypes)
	require.Equal(t, "00000000-0000-0000-0000-000000000000", space.ID.String())
	require.Equal(t, "spaces", space.Type)
	require.Equal(t, "", *space.Attributes.Description)
	require.Equal(t, "yet_another", *space.Attributes.Name)
}

func TestGetNamespaceByTypeErrorFromUserServices(t *testing.T) {
	mockWitClient := &MockWitClient{
		UserServiceHttpResponseError: errors.New("error"),
	}
	mockOSIOClient := controller.CreateOSIOClient(mockWitClient, &controller.IOResponseReader{})

	namespaceAttributes, err := mockOSIOClient.GetNamespaceByType(&MockContext{}, nil, "namespace")
	require.Error(t, err)
	require.Nil(t, namespaceAttributes)
}

func TestGetNamespaceByTypeNoMatch(t *testing.T) {
	mockOSIOClient := controller.CreateOSIOClient(&MockWitClient{}, &controller.IOResponseReader{})
	mockUserService := &app.UserService{
		Attributes: &app.UserServiceAttributes{
			Namespaces: make([]*app.NamespaceAttributes, 0),
		},
	}

	namespaceAttributes, err := mockOSIOClient.GetNamespaceByType(&MockContext{}, mockUserService, "namespace")
	require.NoError(t, err)
	require.Nil(t, namespaceAttributes)
}

func TestGetNamespaceByTypeMatchNamespace(t *testing.T) {
	namespaceType := "desiredType"
	jsonProvider := &JsonResponseReader{
		jsonBytes: readJsonFromTestFile(t, deploymentsOsioTestFilePath + "namespace-by-type.json"),
	}
	mockNamespace := &app.NamespaceAttributes{
		Type: &namespaceType,
	}
	mockResponse := &http.Response{
		Body:       &MockResponseBodyReader{},
		StatusCode: http.StatusOK,
	}
	mockWitClient := &MockWitClient{
		UserServiceHttpResponse: mockResponse,
	}
	mockOSIOClient := controller.CreateOSIOClient(mockWitClient, jsonProvider)
	mockUserService := &app.UserService{
		Attributes: &app.UserServiceAttributes{
			Namespaces: []*app.NamespaceAttributes{mockNamespace},
		},
	}
	namespaceAttributes, err := mockOSIOClient.GetNamespaceByType(&MockContext{}, mockUserService, namespaceType)
	require.NoError(t, err)
	require.Equal(t, mockNamespace, namespaceAttributes)
}

func TestGetNamespaceByTypeMatchNamespaceWithDiscoveredUserService(t *testing.T) {
	namespaceType := "desiredType"
	jsonProvider := &JsonResponseReader{
		jsonBytes: readJsonFromTestFile(t, deploymentsOsioTestFilePath + "namespace-discovered-user-service.json"),
	}
	mockResponse := &http.Response{
		Body:       &MockResponseBodyReader{},
		StatusCode: http.StatusOK,
	}
	mockWitClient := &MockWitClient{
		UserServiceHttpResponse: mockResponse,
	}
	mockOSIOClient := controller.CreateOSIOClient(mockWitClient, jsonProvider)
	namespaceAttributes, err := mockOSIOClient.GetNamespaceByType(&MockContext{}, nil, namespaceType)
	require.NoError(t, err)
	require.Equal(t, namespaceType, *namespaceAttributes.Type)
	require.Equal(t, "some-name", *namespaceAttributes.Name)
	require.Equal(t, "some-state", *namespaceAttributes.State)
	require.Equal(t, "123", *namespaceAttributes.Version)
}
