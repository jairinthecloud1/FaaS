package main

import (
	"os"

	handler "github.com/jairinthecloud1/FaaS/internal"
	"github.com/jairinthecloud1/FaaS/internal/function"
	myMiddleware "github.com/jairinthecloud1/FaaS/internal/middleware"
	"github.com/jairinthecloud1/FaaS/internal/service"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/github"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

func init() {
	goth.UseProviders(
		github.New(
			os.Getenv("GITHUB_CLIENT_ID"),
			os.Getenv("GITHUB_CLIENT_SECRET"),
			"http://localhost:8090/auth/github/callback",
		),
	)
}

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	api := e.Group("/api")

	if err := function.ConfigDockerClient(); err != nil {
		log.WithError(err).Error("failed to create docker client")
		os.Exit(1)
	}

	if err := service.ConfigK8Client(); err != nil {
		log.WithError(err).Error("failed to create k8s client")
		os.Exit(1)
	}

	api.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	api.POST("/functions", handler.PostFunctionHandler, myMiddleware.IsAuthenticated)

	api.GET("/functions/:name", handler.GetFunctionHandler)

	api.GET("/functions", handler.ListFunctionsHandler, myMiddleware.IsAuthenticated)

	// Start the server on port 8090.
	e.Logger.Fatal(e.Start(":8090"))
}
