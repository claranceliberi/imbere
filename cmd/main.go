package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rssb/imbere/pkg/pull_request"
)

func main() {
	router := gin.Default()

	r := router.Group("/api/v1")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/github/webhook", pull_request.HandleWebhook)

	router.Run()
}

// function that receives repoaddress, branchname
// clone the branch from the repo
// install
// build
// start pm2 instance
// store details in db(sqlite)

// communicate to github the status (probably a separate function that would communicate every step)

// EVENTS
// pull_request.opened OR pull_request.reopened
// 		CREATE A RECORD IN DB WITH NECESSARY INFORMATION FOR THE PR

// workflow_run.completed
// 		UPDATE DB WITH WORKFLOW STATUS 0 or 1, meaning succeeded or not
// 		IF PR WAS LABELED 'IMBERE_DEPLOY' , DEPLOY

// pull_request.labeled
// 		UPDATE DB LABELED TO DEPLOY
// 		IF WORKFLOW WAS 1, DEPLOY

// pull_request.unlabeled
// UPDATE DB LABEL TO NOT DEPLOY

// pull_request.closed
// MARK PR CLOSED , NO MORE DEPLOY

// DATABASE TABLE
// ID, PR ID, PR_NUMBER(string), BRANCH(string), URL(string), WORKFLOW_SUCCEEDED(1,0), LABELED_TO_DEPLOY(1,0), ACTIVE(1,0), ....INFORMATION FROM PM2
