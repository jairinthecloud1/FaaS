package router

import (
	"encoding/gob"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"faas-api/platform/authenticator"
	"faas-api/platform/middleware"
	"faas-api/web/app/callback"
	"faas-api/web/app/home"
	"faas-api/web/app/login"
	"faas-api/web/app/logout"
	"faas-api/web/app/user"
)

// New registers the routes and returns the router.
func New(auth *authenticator.Authenticator) *gin.Engine {
	router := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("auth-session", store))

	router.Static("/public", "web/static")
	router.LoadHTMLGlob("web/template/*")

	router.GET("/", home.Handler)
	router.GET("/login", login.Handler(auth))
	router.GET("/callback", callback.Handler(auth))
	router.GET("/user", middleware.IsAuthenticated, user.Handler)
	router.GET("/logout", logout.Handler)

	router.GET("/test", middleware.IsAuthenticated, func(ctx *gin.Context) {
		username, ok := ctx.Get("username")
		if !ok {
			ctx.JSON(400, gin.H{"error": "username not found"})
			return
		}
		provider, ok := ctx.Get("provider")
		if !ok {
			ctx.JSON(400, gin.H{"error": "provider not found"})
			return
		}
		ctx.JSON(200, gin.H{
			"username": username,
			"provider": provider,
		})
	})

	// router.POST("/functions", middleware.IsAuthenticated, handler.PostFunctionHandler)

	return router
}

// e := echo.New()
// e.Use(middleware.Logger())
// e.Use(middleware.Recover())
// api := e.Group("/api")

// if err := function.ConfigDockerClient(); err != nil {
// 	log.WithError(err).Error("failed to create docker client")
// 	os.Exit(1)
// }

// if err := service.ConfigK8Client(); err != nil {
// 	log.WithError(err).Error("failed to create k8s client")
// 	os.Exit(1)
// }

// api.GET("/health", func(c echo.Context) error {
// 	return c.String(200, "OK")
// })

// api.POST("/functions", handler.PostFunctionHandler, myMiddleware.IsAuthenticated)

// api.GET("/functions/:name", handler.GetFunctionHandler, myMiddleware.IsAuthenticated)

// api.GET("/functions", handler.ListFunctionsHandler, myMiddleware.IsAuthenticated)

// // Start the server on port 8090.
// e.Logger.Fatal(e.Start(":8090"))
