package handler

import (
	"faas-api/internal/function"
	"faas-api/internal/k8/namespace"
	"faas-api/internal/service"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func PostFunctionHandler(c *gin.Context) {

	fileBytes, runtime, name, envVars, err := function.ProcessRequestData(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("failed to process request data: %v", err)})
		return
	}

	function := function.FunctionRequest{
		Runtime: runtime,
		Name:    name,
		EnvVars: envVars,
		File:    fileBytes,
	}

	if err := function.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid function request: %v", err)})
		return
	}

	username := c.GetString("username")
	provider := c.GetString("provider")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	namespace, err := namespace.CreateOrGetNamespace(c, service.Clientset, username, provider)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create or get namespace: %v", err)})
		return
	}

	result, err := function.Serve(namespace)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to serve function: %v", err)})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "Function deployed successfully",
		"result":  result,
	})

}

func GetFunctionHandler(c *gin.Context) {
	functionName := c.Param("name")
	if functionName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "function name is required"})
		return
	}

	// get username from context
	username := c.GetString("username")
	provider := c.GetString("provider")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	function, err := service.GetKnativeService(service.Clientset, namespace.BuildNameSpaceName(username, provider), functionName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to get function: %v", err)})
		return
	}

	if function == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "function not found"})
		return
	}
	c.JSON(http.StatusOK, function)
}

func ListFunctionsHandler(c *gin.Context) {
	// get username from context
	username := c.GetString("username")
	provider := c.GetString("provider")
	if username == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	functions, err := service.ListKnativeServices(service.Clientset, namespace.BuildNameSpaceName(username, provider))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to list functions: %v", err)})
		return
	}
	if functions == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "no functions found"})
		return
	}

	c.JSON(http.StatusOK, functions)
}
