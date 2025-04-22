package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

var Clientset dynamic.Interface

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

// Root structure for Knative Service
type KnativeService struct {
	APIVersion string   `json:"apiVersion"`
	Kind       string   `json:"kind"`
	Metadata   Metadata `json:"metadata"`
	Spec       Spec     `json:"spec"`
	Status     Status   `json:"status"`
}

// Metadata about the Knative Service
type Metadata struct {
	Annotations       map[string]string `json:"annotations"`
	CreationTimestamp time.Time         `json:"creationTimestamp"`
	Generation        int               `json:"generation"`
	ManagedFields     []ManagedField    `json:"managedFields"`
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	ResourceVersion   string            `json:"resourceVersion"`
	UID               string            `json:"uid"`
}

// Structure for managed fields
type ManagedField struct {
	APIVersion  string                 `json:"apiVersion"`
	FieldsType  string                 `json:"fieldsType"`
	FieldsV1    map[string]interface{} `json:"fieldsV1"`
	Manager     string                 `json:"manager"`
	Operation   string                 `json:"operation"`
	Subresource string                 `json:"subresource,omitempty"`
	Time        time.Time              `json:"time"`
}

// Spec of the Knative Service
type Spec struct {
	Template Template       `json:"template"`
	Traffic  []TrafficSplit `json:"traffic"`
}

// Template for the spec
type Template struct {
	Metadata MetadataSpec  `json:"metadata"`
	Spec     ContainerSpec `json:"spec"`
}

// MetadataSpec is used in the template section
type MetadataSpec struct {
	CreationTimestamp interface{} `json:"creationTimestamp"`
}

// ContainerSpec inside the template
type ContainerSpec struct {
	ContainerConcurrency int         `json:"containerConcurrency"`
	Containers           []Container `json:"containers"`
	EnableServiceLinks   bool        `json:"enableServiceLinks"`
	TimeoutSeconds       int         `json:"timeoutSeconds"`
}

// Container specification
type Container struct {
	Image          string                 `json:"image"`
	Name           string                 `json:"name"`
	ReadinessProbe ReadinessProbe         `json:"readinessProbe"`
	Resources      map[string]interface{} `json:"resources"`
}

// ReadinessProbe specification for containers
type ReadinessProbe struct {
	SuccessThreshold int       `json:"successThreshold"`
	TCPSocket        TCPSocket `json:"tcpSocket"`
}

// TCPSocket specification
type TCPSocket struct {
	Port int `json:"port"`
}

// TrafficSplit for traffic allocation
type TrafficSplit struct {
	LatestRevision bool `json:"latestRevision"`
	Percent        int  `json:"percent"`
}

// Status of the Knative Service
type Status struct {
	Address                   Address         `json:"address"`
	Conditions                []Condition     `json:"conditions"`
	LatestCreatedRevisionName string          `json:"latestCreatedRevisionName"`
	LatestReadyRevisionName   string          `json:"latestReadyRevisionName"`
	ObservedGeneration        int             `json:"observedGeneration"`
	Traffic                   []TrafficStatus `json:"traffic"`
	URL                       string          `json:"url"`
}

// Address information
type Address struct {
	URL string `json:"url"`
}

// Condition structure in status
type Condition struct {
	LastTransitionTime time.Time `json:"lastTransitionTime"`
	Status             string    `json:"status"`
	Type               string    `json:"type"`
}

// TrafficStatus for traffic details in status
type TrafficStatus struct {
	LatestRevision bool   `json:"latestRevision"`
	Percent        int    `json:"percent"`
	RevisionName   string `json:"revisionName"`
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

func JsontoService(data []byte) (*KnativeService, error) {
	ksvc := &KnativeService{}
	if err := json.Unmarshal(data, ksvc); err != nil {
		return nil, fmt.Errorf("failed to unmarshal knative service: %w", err)
	}

	return ksvc, nil
}

func UnstructuredToService(unstructuredKsvc *unstructured.Unstructured) (*KnativeService, error) {
	jsonData, err := json.Marshal(unstructuredKsvc.Object)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal unstructured object: %w", err)
	}
	ksvc, err := JsontoService(jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to convert unstructured object to KnativeService: %w", err)
	}
	return ksvc, nil
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
func GetKnativeService(client dynamic.Interface, namespace, name string) (*KnativeService, error) {
	ksvc, err := client.Resource(knativeServiceGVR).Namespace(namespace).Get(context.Background(), name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get knative service %s/%s: %w", namespace, name, err)
	}
	ksvcData, err := json.Marshal(ksvc.Object)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal knative service object: %w", err)
	}
	ksvcObj := &KnativeService{}
	if err := json.Unmarshal(ksvcData, ksvcObj); err != nil {
		return nil, fmt.Errorf("failed to unmarshal knative service object: %w", err)
	}
	return ksvcObj, nil
}

func ListKnativeServices(client dynamic.Interface, namespace string) ([]KnativeService, error) {
	ksvcs, err := client.Resource(knativeServiceGVR).Namespace(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list knative services in namespace %s: %w", namespace, err)
	}

	var services []KnativeService
	for _, unstructuredKsvc := range ksvcs.Items {
		svc, err := UnstructuredToService(&unstructuredKsvc)
		if err != nil {
			return nil, fmt.Errorf("failed to convert unstructured object to KnativeService: %w", err)
		}
		services = append(services, *svc)
	}
	return services, nil
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
