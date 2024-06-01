package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rssb/imbere/pkg/db"
	"github.com/rssb/imbere/pkg/webhook"
)

func main() {
	db.DbInit() //

	router := gin.Default()

	r := router.Group("/api/v1")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/github/webhook", webhook.HandleWebhook)

	router.Run()
}
