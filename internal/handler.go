package handler

import (
	"faas-api/internal/function"
	"faas-api/internal/k8/namespace"
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

	// get username from context
	username := c.Get("username").(string)
	provider := c.Get("provider").(string)
	if username == "" {
		return c.JSON(http.StatusUnauthorized, "username is required")
	}

	namespace, err := namespace.CreateOrGetNamespace(c.Request().Context(), service.Clientset, username, provider)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, fmt.Sprintf("error creating or getting namespace: %s", err.Error()))
	}

	// send namespace to function.Serve

	result, err := function.Serve(namespace)
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

	username := c.Get("username").(string)
	provider := c.Get("provider").(string)

	function, err := service.GetKnativeService(service.Clientset, namespace.BuildNameSpaceName(username, provider), functionName)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, function)
}

func ListFunctionsHandler(c echo.Context) error {
	username := c.Get("username").(string)
	provider := c.Get("provider").(string)

	functions, err := service.ListKnativeServices(service.Clientset, namespace.BuildNameSpaceName(username, provider))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, functions)
}
