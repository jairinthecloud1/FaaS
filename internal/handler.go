package handler

import (
	"faas-api/internal/function"
	"faas-api/internal/service"
	"fmt"
	"net/http"

	echo "github.com/labstack/echo/v4"
)



func PostFunctionHandler(c echo.Context) error {

	fileBytes, runtime, name, envVars, err := function.ProcessRequestData(c)
	if err != nil {
		return fmt.Errorf("error processing request data: %w", err)
	}

	function := function.FunctionRequest{
		Runtime: runtime,
		Name:    name,
		EnvVars: envVars,
		File:    fileBytes,
	}

	if err := function.Validate(); err != nil {
		return c.JSON(http.StatusBadRequest, err.Error())
	}

	result, err := function.Serve()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, result)
}

func GetFunctionHandler(c echo.Context) error {
	functionName := c.Param("name")
	if functionName == "" {
		return c.JSON(http.StatusBadRequest, "function name is required")
	}

	function, err := service.GetKnativeService(service.Clientset, "default", functionName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, function)
}

func ListFunctionsHandler(c echo.Context) error {
	functions, err := service.ListKnativeServices(service.Clientset, "default")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, functions)
}
