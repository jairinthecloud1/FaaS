package namespace

import (
	"flag"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var clientset dynamic.Interface

func clientSetup() {

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

func TestNamespace(t *testing.T) {

	clientSetup()

	username := "testuser"
	provider := "testprovider"

	namespace, err := CreateOrGetNamespace(t.Context(), clientset, username, provider)
	require.NoError(t, err, "should not return an error when creating namespace")
	require.NotEmpty(t, namespace, "result should not be empty")
	require.Equal(t, provider+"-"+username, namespace, "namespace should match the expected format")

	err = DeleteNamespace(t.Context(), clientset, username, provider)
	require.NoError(t, err, "should not return an error when deleting namespace")
}
