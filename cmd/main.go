package main

import (
	"net/http"

	"faas-api/internal/middleware/authenticator"
	"faas-api/internal/router"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func main() {

	auth, err := authenticator.New()
	if err != nil {
		log.Fatalf("Failed to initialize the authenticator: %v", err)
	}

	rtr := router.New(auth)

	rtr.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// if err := function.ConfigDockerClient(); err != nil {
	// 	log.WithError(err).Error("failed to create docker client")
	// 	os.Exit(1)
	// }

	// if err := service.ConfigK8Client(); err != nil {
	// 	log.WithError(err).Error("failed to create k8s client")
	// 	os.Exit(1)
	// }

	rtr.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	log.Print("Server listening on http://localhost:8090/")
	if err := http.ListenAndServe("0.0.0.0:8090", rtr); err != nil {
		log.Fatalf("There was an error with the http server: %v", err)
	}

}
