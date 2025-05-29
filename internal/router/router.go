// platform/router/router.go

package router

import (
	"encoding/gob"
	"faas-api/internal/middleware/authenticator"
	"faas-api/internal/router/callback"
	"faas-api/internal/router/login"
	"faas-api/internal/router/logout"
	"faas-api/internal/router/user"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

// New registers the routes and returns the router.
func New(auth *authenticator.Authenticator) *gin.Engine {
	router := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	router.GET("/login", login.Handler(auth))
	router.GET("/callback", callback.Handler(auth))
	router.GET("/user", user.Handler)
	router.GET("/logout", logout.Handler)

	// api := router.Group("/api")

	// api.POST("/functions", handler.PostFunctionHandler, middleware.IsAuthenticated)

	// api.GET("/functions/:name", handler.GetFunctionHandler, middleware.IsAuthenticated)

	// api.GET("/functions", handler.ListFunctionsHandler, middleware.IsAuthenticated)

	return router
}
