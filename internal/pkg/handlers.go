package function

import (
	"fmt"
	"net/http"

	echo "github.com/labstack/echo/v4"
)

func PostFunctionHandler(c echo.Context) error {

	fileBytes, runtime, name, envVars, err := ProcessRequestData(c)
	if err != nil {
		return fmt.Errorf("error processing request data: %w", err)
	}

	function := FunctionRequest{
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
