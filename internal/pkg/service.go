package function

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type Service struct {
	Image        string
	Namespace    string
	FunctionName string
}

const apiVersion = "serving.knative.dev/v1"

type KService struct {
	ApiVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   Metadata     `yaml:"metadata"`
	Spec       KServiceSpec `yaml:"spec"`
}

type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

type KServiceSpec struct {
	Template Template `yaml:"template"`
}

type Template struct {
	Spec Spec `yaml:"spec"`
}

type Spec struct {
	Containers []Container `yaml:"containers"`
}

type Container struct {
	Image string `yaml:"image"`
}

func (s *Service) Deploy() error {
	kubeHost := "https://kubernetes.default.svc"
	// path to token file
	tokenPath := "~/.minikube/profiles/minikube/client.key"
	// path to certificate file
	certPath := "~/.minikube/profiles/minikube/client.crt"

	endpoint := fmt.Sprintf("%s/apis/%s/namespaces/%s/services", kubeHost, apiVersion, s.Namespace)

	url, err := url.Parse(endpoint)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %w", err)
	}

	kservice := KService{
		ApiVersion: apiVersion,
		Kind:       "Service",
		Metadata: Metadata{
			Name:      s.FunctionName,
			Namespace: s.Namespace,
		},
		Spec: KServiceSpec{
			Template: Template{
				Spec: Spec{
					Containers: []Container{
						{
							Image: s.Image,
						},
					},
				},
			},
		},
	}

	httpClient := &http.Client{}

	// prepare the body of the request
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(kservice)
	if err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, url.String(), body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// set the headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", tokenPath))
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", certPath))

	// send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// check the response
	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}

