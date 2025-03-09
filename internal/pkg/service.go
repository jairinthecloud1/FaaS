package function

import (
	"context"
	"fmt"

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

func (s *Service) Deploy(client dynamic.Interface, namespace string) (*unstructured.Unstructured, error) {
	unstructuredKsvc := s.toUnstructured()
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

// GetFunctionURL retrieves the URL for a Knative Service (function) by reading its status.
func (s *Service) GetFunctionURL(client dynamic.Interface) (string, error) {
	namespace := s.Namespace
	name := s.FunctionName
	ksvc, err := client.Resource(knativeServiceGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get knative service %s/%s: %w", namespace, name, err)
	}

	// The URL is typically set in the status.url field once the service is ready.
	url, found, err := unstructured.NestedString(ksvc.Object, "status", "url")
	if err != nil {
		return "", fmt.Errorf("error extracting URL from knative service status: %w", err)
	}
	if !found || url == "" {
		return "", fmt.Errorf("URL not found in knative service status for %s/%s", namespace, name)
	}
	return url, nil
}

// GetKnativeService retrieves a Knative Service (ksvc) by namespace and name using the provided dynamic client.
func GetKnativeService(client dynamic.Interface, namespace, name string) (*unstructured.Unstructured, error) {
	ksvc, err := client.Resource(knativeServiceGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get knative service %s/%s: %w", namespace, name, err)
	}
	return ksvc, nil
}

func ListKnativeServices(client dynamic.Interface, namespace string) (*unstructured.UnstructuredList, error) {
	ksvcs, err := client.Resource(knativeServiceGVR).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list knative services in namespace %s: %w", namespace, err)
	}
	return ksvcs, nil
}
