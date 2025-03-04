package function

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
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

func getEnvironmentVariables() (string, string) {
	username, err := os.LookupEnv("REGISTRY_USERNAME")
	if !err {
		username = ""
	}
	password, err := os.LookupEnv("REGISTRY_IDENTITY_TOKEN")
	if !err {
		password = ""
	}
	return username, password
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
	username, password := getEnvironmentVariables()

	authConfig := registry.AuthConfig{
		Username: username,
		Password: password,
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

	image, err := f.BuildDockerImage()
	if err != nil {
		return "", fmt.Errorf("failed to build Docker image: %w", err)
	}

	// Build the YAML string dynamically using fmt.Sprintf.
	serviceYaml := fmt.Sprintf("apiVersion: serving.knative.dev/v1\n"+
		"kind: Service\n"+
		"metadata:\n"+
		"  name: %s\n"+
		"  namespace: %s\n"+
		"spec:\n"+
		"  template:\n"+
		"    spec:\n"+
		"      containers:\n"+
		"      - image: %s\n", f.Name, "default", image)

	// Print the resulting YAML.
	fmt.Println(serviceYaml)

	// store the service yaml in a file
	file, err := os.Create("service.yaml")
	if err != nil {
		return "", fmt.Errorf("failed to create service.yaml: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(serviceYaml)
	if err != nil {
		return "", fmt.Errorf("failed to write to service.yaml: %w", err)
	}

	// kubectl apply -f service.yaml
	app := "kubectl"
	action := "apply"
	arg0 := "-f"
	arg1 := "service.yaml"
	cmd := exec.Command(app, action, arg0, arg1)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("failed to apply service.yaml: %w", err)
	}

	return fmt.Sprintf("Service %v successfully deployed", f.Name), nil

}

func (f *FunctionRequest) GetImageName() string {
	username, _ := getEnvironmentVariables()
	return username + "/" + f.Name
}
