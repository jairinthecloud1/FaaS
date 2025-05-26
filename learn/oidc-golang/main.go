package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/labstack/echo/v4"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
)

// logRequest logs details about the request and its cookies
func logRequest(prefix string, r *http.Request) {
	log.Printf("%s %s %s", prefix, r.Method, r.URL.String())
	for _, c := range r.Cookies() {
		log.Printf("  Cookie: %s = %s", c.Name, c.Value)
	}
}

// logSession logs the session content (for gorilla/sessions store)
func logSession(prefix string, r *http.Request, store *sessions.CookieStore) {
	session, err := store.Get(r, gothic.SessionName)
	if err != nil {
		log.Printf("%s [WARN] getting session: %v", prefix, err)
	} else {
		out, _ := json.MarshalIndent(session.Values, "", "  ")
		log.Printf("%s Session[%s]: %s", prefix, session.ID, out)
	}
}

// AuthMiddleware checks session for authenticated user and sets "username" in context
func AuthMiddleware(store *sessions.CookieStore) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			logRequest("[AuthMiddleware]", c.Request())
			logSession("[AuthMiddleware]", c.Request(), store)

			// Try to get user data from session
			user, err := gothic.GetFromSession("github", c.Request())
			if err != nil || user == "" {
				log.Printf("[AuthMiddleware] No user in session or error: %v", err)
				return c.String(http.StatusUnauthorized, "Unauthorized. Login at /auth/github first.")
			}
			log.Printf("[AuthMiddleware] User in session: %s", user)

			// Try to complete the auth
			completeUser, err := gothic.CompleteUserAuth(c.Response(), c.Request())
			if err != nil {
				log.Printf("[AuthMiddleware] CompleteUserAuth failed: %v", err)
				return c.String(http.StatusUnauthorized, "Session expired or unauthorized. Please login.")
			}
			log.Printf("[AuthMiddleware] Authenticated: %#v", completeUser)
			c.Set("username", completeUser.NickName)
			return next(c)
		}
	}
}

func main() {
	// Prepare provider
	clientID := os.Getenv("GITHUB_CLIENT_ID")
	clientSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	callbackURL := "http://localhost:8090/auth/github/callback?provider=github"

	if clientID == "" || clientSecret == "" {
		log.Fatal("GITHUB_CLIENT_ID and GITHUB_CLIENT_SECRET must be set in env")
	}

	goth.UseProviders(
		github.New(
			clientID,
			clientSecret,
			callbackURL,
		),
	)

	key := "secret"      // Replace with your SESSION_SECRET or similar
	maxAge := 86400 * 30 // 30 days
	isProd := false      // Set to true when serving over https

	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true // HttpOnly should always be enabled
	store.Options.Secure = isProd

	gothic.Store = store
	e := echo.New()

	// Print config for debugging
	log.Printf("Using goth provider GitHub with callback=%s", callbackURL)
	log.Printf("Using Gothic SessionName=%q", gothic.SessionName)

	// Start GitHub OAuth login
	e.GET("/auth/github", func(c echo.Context) error {
		logRequest("[/auth/github]", c.Request())
		logSession("[/auth/github]", c.Request(), store)
		provider := c.QueryParam("provider")
		if provider == "" {
			log.Printf("[/auth/github] provider missing, adding 'github'")
			c.Request().URL.RawQuery = "provider=github"
		} else {
			log.Printf("[/auth/github] provider=%q", provider)
		}
		gothic.BeginAuthHandler(c.Response().Writer, c.Request())

		return nil
	})

	// OAuth callback
	e.GET("/auth/github/callback", func(c echo.Context) error {
		logRequest("[/auth/github/callback]", c.Request())
		logSession("[/auth/github/callback]", c.Request(), store)
		provider := c.QueryParam("provider")
		log.Printf("[/auth/github/callback] provider: %q", provider)

		user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
		if err != nil {
			log.Printf("[/auth/github/callback] Auth failed: %v", err)
			return c.String(http.StatusUnauthorized, "Auth failed: "+err.Error())
		}
		log.Printf("[/auth/github/callback] User: %#v", user)

		// Show session contents after auth
		logSession("[/auth/github/callback](post-auth)", c.Request(), store)

		return c.JSON(http.StatusOK, map[string]string{
			"username": user.NickName,
			"email":    user.Email,
			"name":     user.Name,
		})
	})

	// Protected route: only accessible after login
	e.GET("/api/protected", func(c echo.Context) error {
		logRequest("[/api/protected]", c.Request())
		logSession("[/api/protected]", c.Request(), store)
		usernameIfc := c.Get("username")
		var username string
		if usernameIfc != nil {
			username = fmt.Sprintf("%v", usernameIfc)
		}
		log.Printf("[/api/protected] username=%q", username)

		return c.JSON(http.StatusOK, map[string]string{
			"message": "Welcome to the protected route!",
			"user":    username,
		})
	}, AuthMiddleware(store))

	// Home route
	e.GET("/", func(c echo.Context) error {
		logRequest("[/]", c.Request())
		logSession("[/]", c.Request(), store)
		return c.HTML(http.StatusOK, `<a href="/auth/github?provider=github">Login with GitHub</a>`)
	})

	log.Println("Starting on :8090")
	e.Logger.Fatal(e.Start(":8090"))
}
