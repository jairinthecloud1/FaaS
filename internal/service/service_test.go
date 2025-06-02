package service

import (
	"flag"
	"path/filepath"
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientset dynamic.Interface

func initK8() {

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// create the clientset
	clientset, err = dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
}

func TestGetKnativeService(t *testing.T) {
	// Create an unstructured Knative Service object to simulate a real resource.
	ksvc := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":      "test-service",
				"namespace": "default",
			},
			"spec": map[string]interface{}{
				"template": map[string]interface{}{
					"spec": map[string]interface{}{
						"containers": []interface{}{
							map[string]interface{}{
								"image": "gcr.io/test/image:latest",
							},
						},
					},
				},
			},
		},
	}

	// Create a simple runtime scheme. The dynamic fake client doesn't require you to add the object
	// to the scheme if you are using unstructured types.
	scheme := runtime.NewScheme()

	// Initialize a fake dynamic client with the test Knative Service.
	client := dynamicfake.NewSimpleDynamicClient(scheme, ksvc)

	// Attempt to retrieve the Knative Service using our  client.
	knativeService, err := GetKnativeService(client, "default", "test-service")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify that the retrieved service has the expected name and namespace.
	if knativeService.Metadata.Name != "test-service" {
		t.Errorf("expected service name 'test-service', got %s", knativeService.Metadata.Name)
	}

	if knativeService.Metadata.Namespace != "default" {
		t.Errorf("expected namespace 'default', got %s", knativeService.Metadata.Namespace)
	}

}

func TestListKnativeServiceWithRealClient(t *testing.T) {

	// Retrieve the Knative Service using the real client.
	ret, err := ListKnativeServices(clientset, "default")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// The result may be empty if there are no Knative Services in the namespace.
	if len(ret) == 0 {
		t.Logf("no Knative Services found in namespace 'default'")
	}

	// Verify that the retrieved services have the expected namespace.
	for _, svc := range ret {
		if svc.Metadata.Namespace != "default" {
			t.Errorf("expected namespace 'default', got %s", svc.Metadata.Namespace)
		}

	}

}

func TestCreateKnativeService(t *testing.T) {
	// Create a simple runtime scheme. The dynamic fake client doesn't require you to add the object
	// to the scheme if you are using unstructured types.
	scheme := runtime.NewScheme()

	// Initialize a fake dynamic client.
	client := dynamicfake.NewSimpleDynamicClient(scheme)

	// Attempt to create the Knative Service using our
	service := Service{
		Image:        "gcr.io/test/image:latest",
		Namespace:    "default",
		FunctionName: "test-service",
	}
	ret, err := service.Deploy(client)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify that the created service has the expected name and namespace.
	if ret.GetName() != "test-service" {
		t.Errorf("expected service name 'test-service', got %s", ret.GetName())
	}
	if ret.GetNamespace() != "default" {
		t.Errorf("expected namespace 'default', got %s", ret.GetNamespace())
	}

	// You can further inspect the content of the unstructured object as needed.
	// For example, verify that the spec contains the expected container image.
	spec, found, err := unstructured.NestedMap(ret.Object, "spec", "template", "spec")
	if err != nil || !found {
		t.Errorf("failed to retrieve spec from returned object: %v", err)
	} else {
		containers, found, err := unstructured.NestedSlice(spec, "containers")
		if err != nil || !found || len(containers) == 0 {
			t.Errorf("failed to retrieve containers from spec: %v", err)
		} else {
			container, ok := containers[0].(map[string]interface{})
			if !ok {
				t.Errorf("container is not a map[string]interface{}")
			} else if image, found, _ := unstructured.NestedString(container, "image"); !found || image != "gcr.io/test/image:latest" {
				t.Errorf("expected image 'gcr.io/test/image:latest', got %s", image)
			}
		}
	}
}

func TestCreateKnativeServiceWithRealClient(t *testing.T) {

	initK8()
	fn := Service{
		Image:        "jairjosafath/hellov4:latest",
		Namespace:    "default",
		FunctionName: "test",
	}
	ret, err := fn.Deploy(clientset)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify that the created service has the expected name and namespace.
	if ret.GetName() != "test" {
		t.Errorf("expected service name '', got %s", ret.GetName())
	}
	if ret.GetNamespace() != "default" {
		t.Errorf("expected namespace 'default', got %s", ret.GetNamespace())
	}

	// You can further inspect the content of the unstructured object as needed.
	// For example, verify that the spec contains the expected container image.
	spec, found, err := unstructured.NestedMap(ret.Object, "spec", "template", "spec")
	if err != nil || !found {
		t.Errorf("failed to retrieve spec from returned object: %v", err)
	} else {
		containers, found, err := unstructured.NestedSlice(spec, "containers")
		if err != nil || !found || len(containers) == 0 {
			t.Errorf("failed to retrieve containers from spec: %v", err)
		} else {
			container, ok := containers[0].(map[string]interface{})
			if !ok {
				t.Errorf("container is not a map[string]interface{}")
			} else if image, found, _ := unstructured.NestedString(container, "image"); !found || image != "jairjosafath/hellov4:latest" {
				t.Errorf("expected image 'jairjosafath/hellov4:latest', got %s", image)
			}
		}
	}

	// print the URL of the created service
	url, err := fn.GetUrl(clientset)
	if err != nil {
		t.Fatalf("failed to get service URL: %v", err)
	}
	t.Logf("Service URL: %s", url)

}

func TestGetFunctionURL(t *testing.T) {
	// Create an unstructured Knative Service object to simulate a real resource.
	ksvc := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "serving.knative.dev/v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name":      "test-service",
				"namespace": "default",
			},
			"status": map[string]interface{}{
				"url": "http://test-service.default.example.com",
			},
		},
	}

	// Create a simple runtime scheme. The dynamic fake client doesn't require you to add the object
	// to the scheme if you are using unstructured types.
	scheme := runtime.NewScheme()

	// Initialize a fake dynamic client with the test Knative Service.
	client := dynamicfake.NewSimpleDynamicClient(scheme, ksvc)

	// Attempt to retrieve the Knative Service using our
	service := Service{
		Image:        "gcr.io/test/image:latest",
		Namespace:    "default",
		FunctionName: "test-service",
	}
	ret, err := service.GetUrl(client)

	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	// Verify that the retrieved service has the expected URL.
	if ret != "http://test-service.default.example.com" {
		t.Errorf("expected URL 'http://test-service.default.example.com', got %s", ret)
	}
}

func TestListKnativeServicesFromCluster(t *testing.T) {

	//skip this test if the cluster is not running
	if clientset == nil {
		t.Skip("Skipping test as the cluster is not running")
	}

	ksvcs, err := ListKnativeServices(clientset, "default")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(ksvcs) == 0 {
		t.Logf("no Knative Services found in namespace 'default'")
	}
	// Verify that the retrieved services have the expected namespace.
	for _, svc := range ksvcs {
		if svc.Metadata.Namespace != "default" {
			t.Errorf("expected namespace 'default', got %s", svc.Metadata.Namespace)
		}

		t.Logf("Knative Service Name: %s", svc.Metadata.Name)
	}
}

func TestGetKnativeServiceFromCluster(t *testing.T) {

	initK8()

	//skip this test if the cluster is not running
	if clientset == nil {
		t.Skip("Skipping test as the cluster is not running")

		kservice, err := GetKnativeService(clientset, "default", "hello")
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		t.Log("Knative Service Name: ", kservice.Status.URL)
	}
}
