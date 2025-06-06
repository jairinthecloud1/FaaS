package router

import (
	"encoding/gob"
	"net/http"
	"os"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	handler "faas-api/internal"
	"faas-api/internal/function"
	"faas-api/internal/service"
	"faas-api/platform/authenticator"
	"faas-api/platform/middleware"
	"faas-api/web/app/app"
	"faas-api/web/app/callback"
	"faas-api/web/app/home"
	"faas-api/web/app/login"
	"faas-api/web/app/logout"
	"faas-api/web/app/user"

	log "github.com/sirupsen/logrus"
)

// New registers the routes and returns the router.
func New(auth *authenticator.Authenticator) *gin.Engine {
	if err := function.ConfigDockerClient(); err != nil {
		log.WithError(err).Error("failed to create docker client")
		os.Exit(1)
	}

	if err := service.ConfigK8Client(); err != nil {
		log.WithError(err).Error("failed to create k8s client")
		os.Exit(1)
	}

	router := gin.Default()

	// To store custom types in our cookies,
	// we must first register them using gob.Register
	gob.Register(map[string]interface{}{})

	secret := os.Getenv("COOKIE_SECRET")
	if secret == "" {
		log.Fatal("COOKIE_SECRET environment variable is not set")
		os.Exit(1)
	}

	store := cookie.NewStore([]byte(secret))

	cookieDomain := os.Getenv("COOKIE_DOMAIN")
	cookieSecure := os.Getenv("COOKIE_SECURE")

	// Default: local development
	secure := false
	sameSite := http.SameSiteLaxMode
	if cookieDomain == "" || cookieDomain == "localhost" {
		cookieDomain = ""
		secure = false
		sameSite = http.SameSiteLaxMode
	} else {
		// In cluster/production
		secure = cookieSecure == "true"
		if secure {
			sameSite = http.SameSiteNoneMode
		} else {
			sameSite = http.SameSiteLaxMode
		}
	}

	store.Options(sessions.Options{
		Domain:   cookieDomain,
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		Secure:   secure,
		HttpOnly: true,
		SameSite: sameSite,
	})
	router.Use(sessions.Sessions("auth-session", store))

	router.Static("/public", "web/static")
	router.LoadHTMLGlob("web/template/*")

	api := router.Group("/api")

	api.GET("/", home.Handler)
	api.GET("/login", login.Handler(auth))
	api.GET("/callback", callback.Handler(auth))
	api.GET("/user", middleware.IsAuthenticated, user.Handler)
	api.GET("/logout", logout.Handler)

	api.GET("/health", func(c *gin.Context) {
		c.String(200, "OK")
	})

	protectedAPI := api.Group("", middleware.IsAuthenticated)

	protectedAPI.GET("/app", app.Handler)

	protectedAPI.POST("/functions", handler.PostFunctionHandler)

	protectedAPI.GET("/functions/:name", handler.GetFunctionHandler)

	protectedAPI.GET("/functions", handler.ListFunctionsHandler)

	return router
}
