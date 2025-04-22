package container

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	log "github.com/sirupsen/logrus"
)

func Auth(ctx context.Context, cli *client.Client) error {
	ServerAddress, ok := os.LookupEnv("DOCKER_REGISTRY")
	if !ok {
		return fmt.Errorf("failed to get DOCKER_REGISTRY env")
	}
	defer cli.Close()
	Username, ok := os.LookupEnv("DOCKER_USERNAME")
	if !ok {
		return fmt.Errorf("failed to get DOCKER_USERNAME env")
	}
	Password, ok := os.LookupEnv("DOCKER_PASSWORD")
	if !ok {
		return fmt.Errorf("failed to get DOCKER_PASSWORD env")
	}

	authConfig := registry.AuthConfig{
		Username:      strings.ReplaceAll(Username,"\n", ""),
		Password:      strings.ReplaceAll(Password,"\n", ""),
		ServerAddress: ServerAddress,
	}
	result, err := cli.RegistryLogin(ctx, authConfig)
	if err != nil {
		log.WithError(err).Error("failed to login to registry")
		return err
	}

	log.WithField("status", result.Status).Info("login to registry")

	if result.Status != "Login Succeeded" {
		return err
	}

	return nil
}
