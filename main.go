package knative

import (
	"context"
	"fmt"

	function "github.com/jairinthecloud1/FaaS/internal/pkg"
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

func CreateKnativeService(client dynamic.Interface, namespace string, ksvc *function.KService) (*unstructured.Unstructured, error) {
	unstructuredKsvc := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": ksvc.ApiVersion,
			"kind":       ksvc.Kind,
			"metadata": map[string]interface{}{
				"name":      ksvc.Metadata.Name,
				"namespace": ksvc.Metadata.Namespace,
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"image": ksvc.Spec.Template.Spec.Containers[0].Image,
							},
						},
					},
				},
			},
		},
	}
	created, err := client.Resource(knativeServiceGVR).Namespace(namespace).Create(context.Background(), unstructuredKsvc, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create knative service in namespace %s: %w", namespace, err)
	}
	return created, nil
}
