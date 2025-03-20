package main

import (
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	api := e.Group("/api")

	// Define the POST endpoint.
	// e.POST("/api/functions", function.PostFunctionHandler)

	// health check endpoint
	api.GET("/health", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	// Start the server on port 8080.
	e.Logger.Fatal(e.Start(":8090"))
}
