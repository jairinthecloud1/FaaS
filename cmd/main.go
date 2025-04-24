package main

import (
	"os"

	handler "faas-api/internal"
	"faas-api/internal/function"
	myMiddleware "faas-api/internal/middleware"
	"faas-api/internal/service"

	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	log "github.com/sirupsen/logrus"
)

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
