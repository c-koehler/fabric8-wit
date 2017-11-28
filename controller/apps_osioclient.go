package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fabric8-services/fabric8-wit/app"
	goajwt "github.com/goadesign/goa/middleware/security/jwt"
	rest "k8s.io/client-go/rest"
)

// OsioClient contains configuration and methods for interacting with OSIO API
type OsioClient struct {
	config *rest.Config
}

// NewOsioClient creates an openshift IO client given an http request context
func NewOsioClient(ctx context.Context, witURL string) (*OsioClient, error) {

	config := rest.Config{
		Host:        witURL,
		BearerToken: goajwt.ContextJWT(ctx).Raw,
	}

	client := new(OsioClient)
	client.config = &config

	return client, nil
}

// GetResource - generic JSON resource fetch
func (osioclient *OsioClient) GetResource(url string, allowMissing bool) (map[string]interface{}, error) {
	var body []byte
	fullURL := strings.TrimSuffix(osioclient.config.Host, "/") + url
	fmt.Println("full URL=", fullURL)
	req, err := http.NewRequest("GET", fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+osioclient.config.BearerToken)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	status := resp.StatusCode
	if status == 404 && allowMissing {
		return nil, nil
	} else if status < 200 || status > 300 {
		return nil, fmt.Errorf("Failed to GET url %s due to status code %d", fullURL, status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	var respType map[string]interface{}
	err = json.Unmarshal(respBody, &respType)
	if err != nil {
		return nil, err
	}
	return respType, nil
}

// GetUserServices - fetch array of user services
func (osioclient *OsioClient) GetUserServices() (*app.UserService, error) {
	var body []byte
	fullURL := strings.TrimSuffix(osioclient.config.Host, "/") + "/user/services"
	req, err := http.NewRequest("GET", fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+osioclient.config.BearerToken)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	status := resp.StatusCode
	if status == 404 {
		return nil, nil
	} else if status < 200 || status > 300 {
		return nil, fmt.Errorf("Failed to GET url %s due to status code %d", fullURL, status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	var respType app.UserServiceSingle
	err = json.Unmarshal(respBody, &respType)
	if err != nil {
		return nil, err
	}
	return respType.Data, nil
}

// GetSpaceByName - fetch space given useedname and spacename
func (osioclient *OsioClient) GetSpaceByName(username string, spaceName string, allowMissing bool) (*app.Space, error) {
	fullURL := strings.TrimSuffix(osioclient.config.Host, "/") + "/namedspaces/" + username + "/" + spaceName
	return osioclient.getSpace(fullURL, allowMissing)
}

//GetSpaceByID - fetch space given UUID
func (osioclient *OsioClient) GetSpaceByID(spaceID string, allowMissing bool) (*app.Space, error) {
	fullURL := strings.TrimSuffix(osioclient.config.Host, "/") + "/spaces/" + spaceID
	return osioclient.getSpace(fullURL, allowMissing)
}

func (osioclient *OsioClient) getSpace(fullURL string, allowMissing bool) (*app.Space, error) {
	var body []byte
	req, err := http.NewRequest("GET", fullURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Authorization", "Bearer "+osioclient.config.BearerToken)

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	status := resp.StatusCode
	if status == 404 && allowMissing {
		return nil, nil
	} else if status < 200 || status > 300 {
		return nil, fmt.Errorf("Failed to GET url %s due to status code %d", fullURL, status)
	}

	respBody, err := ioutil.ReadAll(resp.Body)

	var respType app.SpaceSingle
	err = json.Unmarshal(respBody, &respType)
	if err != nil {
		return nil, err
	}
	return respType.Data, nil
}
