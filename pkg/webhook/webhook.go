package webhook

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rssb/imbere/pkg/constants"
	"github.com/rssb/imbere/pkg/pull_request"
	"github.com/rssb/imbere/pkg/utils"
)

// HandleWebhook is the main entry point for handling incoming github webhooks.
// It parses the payload, extracts the event type, and handles the event if it's one of the supported types.
// If the event is handled, it creates a PullRequest from the payload, and then pulls the changes(which later triggers deployment).
// If the event is not handled, it simply returns a 202 Accepted response.
func HandleWebhook(c *gin.Context) {
	var payload map[string]any

	if err := c.ShouldBindJSON(&payload); err != nil {

		utils.ReturnError(c, err.Error())
	}

	// get event type
	event, err := pull_request.ExtractEventType(c, payload)

	if err != nil {
		utils.ReturnError(c, err.Error())
		return
	}

	nameAction := event.GetNameAction()

	fmt.Println(nameAction)

	isHandledEVentAction := constants.ALLOWED_EVENT_ACTIONS[nameAction]

	if isHandledEVentAction {

		err := pull_request.HandlePR(event, payload)

		if err != nil {
			utils.ReturnError(c, err.Error())
		}

	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "not yet started",
	})
}

// PULL REQUESTS
// once any accepted event is triggered on webhook, we do the following
//	 1. we extract information from the payload
//		a. if event is related to new PR or changes in PR (pull_request.open, pull_request.reopen, workflow_run.completed) or labeled deployment IMBERE_DEPLOY
//			a.1. we create directory for the pr or replace it if it exists
//			a.2. we pull latest changes from the pr
//			a.3. we store/update the information for the pr in db
//			a.4. we notify the deployment service to deploy
// 		b. if the event is related to closed pr or unlabeled IMBERE_DEPLOY
// 			b.1 we delete the directory that contains changes
//			b.2 we update information in db
// 			b.3 we notify the deployment service to undeploy
//	Note: Every process is communicated to github

// DEPLOYMENTS
//	a. ON_DEPLOY
//		a.1 install packages
// 		b.2 build project
//		b.3 check if there was no deployment dedicated to that PR
//		b.4 if it existed kill existing deployment
// 		b.5  deploy (to given deployment service)
//		b.6 store/update deployment information in db
//	 b. ON_UNDEPLOY
//		b.1 undeploy (from given deployment service)
// 		b.2 update deployment info in db
//	Note: Every process is communicated to github

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
