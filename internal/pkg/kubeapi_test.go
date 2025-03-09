package function

import (
	"fmt"
	"net/http"
	"net/url"
	"testing"
)

func TestKubeAPISendRequest(t *testing.T) {
	kubeHost := "https://kubernetes.default.svc"
	// path to token file
	tokenPath := "~/.minikube/profiles/minikube/client.key"
	// path to certificate file
	// certPath := "~/.minikube/profiles/minikube/client.crt"

	endpoint := fmt.Sprintf("%s/apis/%s/namespaces/%s/services", kubeHost, apiVersion, "default")

	url, err := url.Parse(endpoint)
	if err != nil {
		t.Errorf("failed to parse URL: %v", err)
	}

	httpClient := &http.Client{}
	req, err := http.NewRequest(http.MethodGet, url.String(), nil)
	if err != nil {
		t.Errorf("failed to create request: %v", err)
	}

	// set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenPath))
	resp, err := httpClient.Do(req)
	if err != nil {
		t.Errorf("failed to send request: %v", err)
	}

	// check the response
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("unexpected status code: %d", resp.StatusCode)
	}
}

func TestHellowWorld(t *testing.T) {
	fmt.Println("Hello, World!")
}
