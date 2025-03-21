package main

import (
	"context"
	"fmt"
	"os"
	"time"

	handler "faas-api/internal"
	"faas-api/internal/container"

	"github.com/docker/docker/client"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

// waitForDocker pings the Docker daemon until it becomes available or times out.
func waitForDocker(ctx context.Context, cli *client.Client, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for {
		if _, err := cli.Ping(ctx); err == nil {
			return nil
		}
		if time.Now().After(deadline) {
			return fmt.Errorf("timed out waiting for Docker daemon")
		}
		log.Warn("Docker daemon not ready, retrying in 2 seconds...")
		time.Sleep(2 * time.Second)
	}
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	api := e.Group("/api")

	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.WithError(err).Error("failed to create docker client")
		os.Exit(1)
	}
	defer cli.Close()

	ctx := context.Background()
	// Wait up to 30 seconds for the Docker daemon to become available.
	if err := waitForDocker(ctx, cli, 30*time.Second); err != nil {
		log.WithError(err).Error("docker daemon not available")
		os.Exit(1)
	}

	if err := container.Auth(ctx, cli); err != nil {
		log.WithError(err).WithField("client", cli).Error("failed to login to registry")
	}

	api.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	api.POST("/functions", handler.PostFunctionHandler)

	// Start the server on port 8090.
	e.Logger.Fatal(e.Start(":8090"))
}
