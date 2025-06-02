package middleware

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// IsAuthenticated is a middleware that checks if
// the user has already been authenticated previously.
func IsAuthenticated(ctx *gin.Context) {

	profile := sessions.Default(ctx).Get("profile")
	if profile == nil {
		ctx.Redirect(http.StatusSeeOther, "/")
	}
	log.Println("user profile:", profile)
	username := profile.(map[string]interface{})["name"].(string)
	sub := profile.(map[string]interface{})["sub"].(string)

	provider := strings.Split(sub, "|")[0]

	ctx.Set("username", strings.ToLower(username))
	ctx.Set("provider", provider)
	ctx.Next()
}
