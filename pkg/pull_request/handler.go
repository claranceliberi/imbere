package pull_request

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func HandleWebhook(c *gin.Context) {
	var payload map[string]any

	if err := c.ShouldBindJSON(&payload); err != nil {

		returnError(c, err.Error())
	}

	// fmt.Println("Got Payload %d", payload)

	// get event type
	event, err := extractEventType(c, payload)

	if err != nil {
		returnError(c, err.Error())
		return
	}

	nameAction := event.nameAction

	isHandledEVentAction := nameAction == "workflow_run.completed" || nameAction == "pull_request.closed" || nameAction == "pull_request.opened" || nameAction == "pull_request.reopened"

	if isHandledEVentAction {

		PR, err := createPullRequestFromPayload(event, payload)

		if err != nil {
			returnError(c, err.Error())
		}

		prService := PullRequestService{
			pr: PR,
		}

		prService.PullChanges()

		if err != nil {
			returnError(c, err.Error())
		}
		c.JSON(http.StatusCreated, gin.H{
			"message": "not yet started",
		})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"message": "not yet started",
	})
}

func createPullRequestFromPayload(event Event, payload map[string]interface{}) (*PullRequest, error) {
	repository, ok := payload["repository"].(map[string]interface{})
	if !ok {
		return &PullRequest{}, fmt.Errorf("failed to parse repository from payload")
	}

	repositoryName, ok := repository["name"].(string)
	if !ok {
		return &PullRequest{}, fmt.Errorf("failed to parse repository name from payload")
	}

	repositoryAddress, ok := repository["html_url"].(string)
	if !ok {
		return &PullRequest{}, fmt.Errorf("failed to parse repository address from payload")
	}

	sshAddress, ok := repository["ssh_url"].(string)
	if !ok {
		return &PullRequest{}, fmt.Errorf("failed to parse ssh address from payload")
	}

	branchName, err := extractBranchName(event, payload)
	if err != nil {
		return &PullRequest{}, err
	}

	prId, err := extractPRID(event, payload)
	if err != nil {
		return &PullRequest{}, err
	}

	prNumber, err := extractPRNumber(event, payload)
	if err != nil {
		return &PullRequest{}, err
	}

	prUrl, err := extractUrl(event, payload)
	if err != nil {
		return &PullRequest{}, err
	}

	PR := PullRequest{
		BranchName:  branchName,
		RepoName:    repositoryName,
		RepoAddress: repositoryAddress,
		SSHAddress:  sshAddress,
		PrID:        prId,
		PrNumber:    prNumber,
		PrUrl:       prUrl,
	}

	return &PR, nil
}

func returnError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"message": message,
	})
	debug.PrintStack()
	fmt.Println("Error occured %+v\n", message)

}
