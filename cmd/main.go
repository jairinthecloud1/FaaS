package main

import (
	function "github.com/jairinthecloud1/FaaS/internal/pkg"
	"github.com/joho/godotenv"
	echo "github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		panic("Error loading .env file")
	}
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// Define the POST endpoint.
	e.POST("/api/functions", function.PostFunctionHandler)

	// Start the server on port 8080.
	e.Logger.Fatal(e.Start(":8090"))
}
