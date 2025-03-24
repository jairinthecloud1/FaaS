package function

import (
	"bytes"
	"context"
	"faas-api/internal/service"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

type EnvVar struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type FunctionRequest struct {
	Runtime string   `json:"runtime"`
	Name    string   `json:"name"`
	EnvVars []EnvVar `json:"env_vars"`
	File    []byte   `json:"file"` // the binary contents of the uploaded zip file (base64 encoded in JSON)
}

func getEnvironmentVariables() (string, string, string) {
	username, ok := os.LookupEnv("DOCKER_USERNAME")
	if !ok {
		username = ""
		log.Println("DOCKER_USERNAME not found")
	}
	password, ok := os.LookupEnv("DOCKER_PASSWORD")
	if !ok {
		password = ""
		log.Println("DOCKER_PASSWORD not found")
	}
	serverAddress, ok := os.LookupEnv("DOCKER_REGISTRY")
	if !ok {
		serverAddress = ""
		log.Println("DOCKER_REGISTRY not found")
	}

	// sanitize the values to avoid errors remove \n
	username = strings.Replace(username, "\n", "", -1)
	password = strings.Replace(password, "\n", "", -1)
	serverAddress = strings.Replace(serverAddress, "\n", "", -1)

	return username, password, serverAddress
}

func (f *FunctionRequest) Validate() error {
	if f.Runtime == "" {
		return fmt.Errorf("runtime is required")
	}
	if f.Name == "" {
		return fmt.Errorf("name is required")
	}
	return nil
}

func (f *FunctionRequest) GetTar() ([]byte, error) {
	tar, err := UnknownToTar(f.File)
	if err != nil {
		return nil, fmt.Errorf("error converting file to tar: %w", err)
	}

	tarWithDocker, err := InjectDockerfile(tar)
	if err != nil {
		return nil, fmt.Errorf("error injecting Dockerfile: %w", err)
	}

	return tarWithDocker, nil
}

func (f *FunctionRequest) BuildDockerImage() (string, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return "", fmt.Errorf("failed to create Docker client: %w", err)
	}

	// login to the registry
	username, password, serverAddress := getEnvironmentVariables()

	log.Printf("Logging in to Docker registry %v, %v, %v", serverAddress, username, password)

	authConfig := registry.AuthConfig{
		Username:      username,
		Password:      password,
		ServerAddress: serverAddress,
	}

	token, err := registry.EncodeAuthConfig(authConfig)
	if err != nil {
		return "", fmt.Errorf("failed to encode auth config: %w", err)
	}

	tar, err := f.GetTar()
	if err != nil {
		return "", err
	}

	log.Printf("Building Docker image %s", f.GetImageName())

	buildCtx := bytes.NewReader(tar)
	buildOptions := types.ImageBuildOptions{
		Tags:        []string{f.GetImageName()},
		Remove:      true,
		ForceRemove: true,
	}
	buildResponse, err := cli.ImageBuild(context.Background(), buildCtx, buildOptions)
	if err != nil {
		return "", fmt.Errorf("failed to build Docker image: %w", err)
	}
	defer buildResponse.Body.Close()

	// Read the build response to completion.
	_, err = io.Copy(os.Stdout, buildResponse.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read build response: %w", err)
	}

	// push the image to the registry
	pushOptions := image.PushOptions{
		RegistryAuth: token,
	}
	pushResponse, err := cli.ImagePush(context.Background(),
		f.GetImageName(),
		pushOptions)
	if err != nil {
		return "", fmt.Errorf("failed to push Docker image: %w", err)
	}
	defer pushResponse.Close()

	// Read the push response to completion.
	_, err = io.Copy(os.Stdout, pushResponse)
	if err != nil {
		return "", fmt.Errorf("failed to read push response: %w", err)
	}

	// Read the push response to a string.
	var response bytes.Buffer
	_, err = io.Copy(&response, pushResponse)

	if err != nil {
		return "", fmt.Errorf("failed to read push response: %w", err)
	}

	// Check if the push was successful.
	if strings.Contains(response.String(), "error") {
		return "", fmt.Errorf("failed to push Docker image: %w", err)
	}

	return strings.Join(buildOptions.Tags, ":"), nil
}

func (f *FunctionRequest) Serve() (string, error) {

	// creates the in-cluster config
	config, err := rest.InClusterConfig()
	if err != nil {
		panic(err.Error())
	}
	// creates the clientset
	clientset, err := dynamic.NewForConfig(config)
	if err != nil {
		panic(err.Error())
	}
	image, err := f.BuildDockerImage()
	if err != nil {
		return "", fmt.Errorf("failed to build Docker image: %w", err)
	}

	log.Printf("Deploying service %s", f.Name)

	// deploy the service
	svc := service.Service{
		FunctionName: f.Name,
		Namespace:    "default",
		Image:        image,
	}

	deployed, err := svc.Deploy(clientset)
	if err != nil {
		return "", fmt.Errorf("failed to deploy service: %w", err)
	}

	if deployed == nil {
		return "", fmt.Errorf("failed to deploy service")
	}

	return fmt.Sprintf("Service %v successfully deployed", f.Name), nil

}

func (f *FunctionRequest) GetImageName() string {
	username, _, registry := getEnvironmentVariables()
	return fmt.Sprintf("%s/%s/%s", registry, username, f.Name)
}
