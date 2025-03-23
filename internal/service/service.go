package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// knativeServiceGVR defines the GroupVersionResource for Knative Services.
var knativeServiceGVR = schema.GroupVersionResource{
	Group:    "serving.knative.dev",
	Version:  "v1",
	Resource: "services", // note: the plural form of "Service"
}

type ServiceOwner struct {
	Name     string
	Email    string
	UserName string
}

type Service struct {
	Image        string
	Namespace    string
	FunctionName string
	Owner        ServiceOwner
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

func (s *Service) Deploy(client dynamic.Interface) (*unstructured.Unstructured, error) {
	unstructuredKsvc := s.toUnstructured()
	namespace := s.Namespace
	created, err := client.Resource(knativeServiceGVR).Namespace(namespace).Create(context.Background(), unstructuredKsvc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create knative service in namespace %s: %w", namespace, err)
	}
	return created, nil
}

func (s *Service) toUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":      s.FunctionName,
				"namespace": s.Namespace,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"image": s.Image,
							},
						},
					},
				},
			},
		},
	}
}

// GetFunctionURL retrieves the URL for a Knative Service (function) by checking its status.
// It first attempts to extract "status.url". If not found or empty, it falls back to "status.address.url".
func GetFunctionURL(client dynamic.Interface, namespace, name string) (string, error) {
	ksvc, err := client.Resource(knativeServiceGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get knative service %s/%s: %w", namespace, name, err)
	}

	// Attempt to get the top-level status.url
	url, found, err := unstructured.NestedString(ksvc.Object, "status", "url")
	if err != nil {
		return "", fmt.Errorf("error extracting URL from knative service status: %w", err)
	}
	if !found || url == "" {
		// Fallback to status.address.url if status.url is not present.
		url, found, err = unstructured.NestedString(ksvc.Object, "status", "address", "url")
		if err != nil {
			return "", fmt.Errorf("error extracting URL from knative service status.address: %w", err)
		}
		if !found || url == "" {
			return "", fmt.Errorf("URL not found in knative service status for %s/%s", namespace, name)
		}
	}
	return url, nil
}

func (s *Service) GetUrl(client dynamic.Interface) (string, error) {
	return GetFunctionURL(client, s.Namespace, s.FunctionName)
}

// GetKnativeService retrieves a Knative Service (ksvc) by namespace and name using the provided dynamic client.
func GetKnativeService(client dynamic.Interface, namespace, name string) (*unstructured.Unstructured, error) {
	ksvc, err := client.Resource(knativeServiceGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get knative service %s/%s: %w", namespace, name, err)
	}
	return ksvc, nil
}

func ListKnativeServices(client dynamic.Interface, namespace string) (*[]Service, error) {
	ksvcs, err := client.Resource(knativeServiceGVR).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list knative services in namespace %s: %w", namespace, err)
	}

	var services []Service
	for _, unstructuredKsvc := range ksvcs.Items {
		services = append(services, *FromUnstructured(&unstructuredKsvc))
	}
	return &services, nil
}

func FromUnstructured(unstructuredKsvc *unstructured.Unstructured) *Service {
	spec := unstructuredKsvc.Object["spec"].(map[string]interface{})
	template := spec["template"].(map[string]interface{})
	containers := template["spec"].(map[string]interface{})["containers"].([]interface{})
	image := containers[0].(map[string]interface{})["image"].(string)
	return &Service{
		Image:        image,
		Namespace:    unstructuredKsvc.GetNamespace(),
		FunctionName: unstructuredKsvc.GetName(),
	}
}

type HarborProject struct {
	ProjectName string `json:"project_name"`
}

func (s *Service) CreateHarborProject(client dynamic.Interface, httpClient *http.Client) error {
	harborProject := &HarborProject{
		ProjectName: s.Owner.UserName,
	}
	ServerAddress, ok := os.LookupEnv("DOCKER_REGISTRY")
	if !ok {
		return fmt.Errorf("failed to get DOCKER_REGISTRY env")
	}
	Username, ok := os.LookupEnv("DOCKER_USERNAME")
	if !ok {
		return fmt.Errorf("failed to get DOCKER_USERNAME env")
	}
	Password, ok := os.LookupEnv("DOCKER_PASSWORD")
	if !ok {
		return fmt.Errorf("failed to get DOCKER_PASSWORD env")
	}

	jsonValue, err := json.Marshal(harborProject)
	if err != nil {
		return fmt.Errorf("failed to marshal Harbor project: %w", err)
	}
	logrus.Println(string(jsonValue))
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/v2.0/projects", ServerAddress), bytes.NewBuffer(jsonValue))
	if err != nil {
		return fmt.Errorf("failed to create Harbor project request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(Username, Password)
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to create Harbor project: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create Harbor project: %s", resp.Status)
	}
	return nil

}
