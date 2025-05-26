package namespace

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

func GetNamespace(ctx context.Context, client dynamic.Interface, username string, provider string) (string, error) {
	resource := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}

	namespace, err := client.Resource(resource).Get(ctx, BuildNameSpaceName(username, provider), metav1.GetOptions{})
	if err != nil {
		return "", err
	}
	return namespace.GetName(), nil
}

func CreateNamespace(ctx context.Context, client dynamic.Interface, username string, provider string) (string, error) {
	resource := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}

	namespace := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Namespace",
			"metadata": map[string]interface{}{
				"name": BuildNameSpaceName(username, provider),
			},
		},
	}

	createdNamespace, err := client.Resource(resource).Create(ctx, namespace, metav1.CreateOptions{})
	if err != nil {
		return "", err
	}
	return createdNamespace.GetName(), nil
}

func CreateOrGetNamespace(ctx context.Context, client dynamic.Interface, username string, provider string) (string, error) {
	namespace, err := GetNamespace(ctx, client, username, provider)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "not found") {
			return CreateNamespace(ctx, client, username, provider)
		}
		return "", fmt.Errorf("error getting namespace: %w", err)
	}
	return namespace, nil
}

func BuildNameSpaceName(username string, provider string) string {
	return fmt.Sprintf("%s-%s", provider, username)
}

func DeleteNamespace(ctx context.Context, client dynamic.Interface, username string, provider string) error {
	resource := schema.GroupVersionResource{
		Version:  "v1",
		Resource: "namespaces",
	}

	name := BuildNameSpaceName(username, provider)
	return client.Resource(resource).Delete(ctx, name, metav1.DeleteOptions{})
}
