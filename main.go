package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"runtime/debug"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.Default()

	r := router.Group("/api/v1")

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.POST("/github/webhook", handleWebhook)

	router.Run()
}

func handleWebhook(c *gin.Context) {
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
		repository, ok := payload["repository"].(map[string]any)
		if !ok {
			returnError(c, err.Error())
			return
		}

		repositoryName, ok := repository["name"].(string)
		if !ok {
			returnError(c, err.Error())
			return
		}

		repositoryAddress, ok := repository["html_url"].(string)
		if !ok {
			returnError(c, err.Error())
			return
		}

		sshAddress, ok := repository["ssh_url"].(string)
		if !ok {
			returnError(c, err.Error())
			return
		}

		branchName, err := extractBranchName(event, payload)
		if err != nil {
			returnError(c, err.Error())
			return
		}

		prId, err := extractPRID(event, payload)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		prNumber, err := extractPRNumber(event, payload)
		if err != nil {
			returnError(c, err.Error())
			return
		}

		PR := MyPullRequest{
			branchName:  branchName,
			repoName:    repositoryName,
			repoAddress: repositoryAddress,
			sshAddress:  sshAddress,
			prID:        prId,
			prNumber:    prNumber,
		}

		err = PR.PullChanges()

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

type Event struct {
	name       string
	action     string
	nameAction string
}

func extractEventType(c *gin.Context, payload map[string]any) (Event, error) {
	// Get the X-GitHub-Event header
	eventType := c.GetHeader("X-GitHub-Event")

	if eventType == "" {
		return Event{}, errors.New("missing X-GitHub-Event header")
	}

	event := Event{
		name: eventType,
	}

	if action, ok := payload["action"].(string); ok {
		// If there is 'action' key, return 'eventType.action'
		event.action = action
		event.nameAction = eventType + "." + action
		return event, nil
	}

	return event, nil
}
func extractBranchName(event Event, payload map[string]any) (string, error) {
	name := event.name
	var branchName string

	if name == "pull_request" {
		pullRequest, ok := payload["pull_request"].(map[string]any)
		if !ok {
			return "", errors.New(event.nameAction + " - Could not extract branch name from pull_request - no pull_request :" + fmt.Sprintf("%v", pullRequest))
		}

		head, ok := pullRequest["head"].(map[string]any)
		if !ok {
			return "", errors.New(event.nameAction + " - Could not extract branch name - no head :" + fmt.Sprintf("%v", head))
		}

		branchName, _ = head["ref"].(string)
	}

	if name == "workflow_run" {
		workflowRun, ok := payload["workflow_run"].(map[string]any)

		if !ok {
			return "", errors.New(event.nameAction + " - Could not extract branch name from workflow_run - no workflow_run :" + fmt.Sprintf("%v", workflowRun))
		}
		pullRequests, ok := workflowRun["pull_requests"].([]interface{})
		if !ok || len(pullRequests) == 0 {
			return "", errors.New(event.nameAction + " - Could not extract branch name from workflow_run - no pull_requests :" + fmt.Sprintf("%v", pullRequests))

		}

		head, ok := pullRequests[0].(map[string]any)["head"].(map[string]any)
		if !ok {
			return "", errors.New(event.nameAction + " - Could not extract branch name from workflow_run - no head :" + fmt.Sprintf("%v", head))
		}
		branchName, _ = head["ref"].(string)
	}

	return branchName, nil
}

func extractPRID(event Event, payload map[string]any) (float64, error) {
	name := event.name
	var prId float64
	errorMessage := errors.New(event.nameAction + " - Could not extract PR ID")

	if name == "pull_request" {
		pullRequest, ok := payload["pull_request"].(map[string]any)
		if !ok {
			return 0, errorMessage
		}

		prId, _ = pullRequest["id"].(float64)
	}

	if name == "workflow_run" {
		workflowRun, ok := payload["workflow_run"].(map[string]any)
		if !ok {
			return 0, errorMessage
		}
		pullRequests, ok := workflowRun["pull_requests"].([]interface{})
		if !ok || len(pullRequests) == 0 {
			return 0, errorMessage

		}

		pullRequest, ok := pullRequests[0].(map[string]any)
		if !ok {
			return 0, errorMessage
		}
		prId, _ = pullRequest["id"].(float64)
	}

	return prId, nil
}

func extractPRNumber(event Event, payload map[string]any) (float64, error) {
	name := event.name
	var prNumber float64
	errorMessage := errors.New(event.nameAction + " - Could not extract PR Number")

	if name == "pull_request" {
		pullRequest, ok := payload["pull_request"].(map[string]any)
		if !ok {
			return 0, errorMessage
		}

		prNumber, _ = pullRequest["number"].(float64)
	}

	if name == "workflow_run" {
		workflowRun, ok := payload["workflow_run"].(map[string]any)
		if !ok {
			return 0, errorMessage
		}
		pullRequests, ok := workflowRun["pull_requests"].([]interface{})
		if !ok || len(pullRequests) == 0 {
			return 0, errorMessage

		}

		pullRequest, ok := pullRequests[0].(map[string]any)
		if !ok {
			return 0, errorMessage
		}
		prNumber, _ = pullRequest["number"].(float64)
	}

	return prNumber, nil
}

func returnError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"message": message,
	})
	debug.PrintStack()
	fmt.Println("Error occured %+v\n", message)

}

// function that receives repoaddress, branchname
// clone the branch from the repo
// install
// build
// start pm2 instance
// store details in db(sqlite)

// communicate to github the status (probably a separate function that would communicate every step)

// PullRequest interface
type PullRequest interface {
	GetBranchName() string
	GetRepoName() string
	GetRepoAddress() string
	GetSSHAddress() string
	GetPRID() float64
	GetPrNumber() float64
	PrepareDir() (string, error)
	PullChanges() error
	CommunicateProgress(status string) error
}

// Deployment interface
type Deployment interface {
	NavigateToDir(dirPath string) error
	InstallDependencies() error
	BuildDependencies() error
	DeployToPM2() error
	StoreDeploymentDetails() error
	CommunicateStatus(status string) error
}

// Concrete type that implements PullRequest
type MyPullRequest struct {
	branchName  string
	repoName    string
	repoAddress string
	sshAddress  string
	prID        float64
	prNumber    float64
}

func (pr *MyPullRequest) GetBranchName() string {
	return pr.branchName
}

func (pr *MyPullRequest) GetRepoName() string {
	return pr.repoName
}

func (pr *MyPullRequest) GetRepoAddress() string {
	return pr.repoAddress
}
func (pr *MyPullRequest) GetSSHAddress() string {
	return pr.sshAddress
}

func (pr *MyPullRequest) GetPRID() float64 {
	return pr.prID
}

func (pr *MyPullRequest) GetPrNumber() float64 {
	return pr.prNumber
}

func (pr *MyPullRequest) PrepareDir() (string, error) {
	prIDStr := fmt.Sprintf("%d", int(pr.GetPrNumber()))
	// dirPath := "/var/lib/imbere/builds/" + pr.GetRepoName() + "/" + pr.GetBranchName() + "_" + prIDStr
	dirPath := "./builds/" + pr.GetRepoName() + "/" + pr.GetBranchName() + "_" + prIDStr

	// check if dir exists and remove it
	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		err := os.RemoveAll(dirPath)

		if err != nil {
			return "", err
		}
	}

	// create dir
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		return "", err
	}

	return dirPath, nil
}

func (pr *MyPullRequest) PullChanges() error {
	_, err := exec.LookPath("git")

	if err != nil {
		return err
	}

	// prepare cloning dir
	dirPath, err := pr.PrepareDir()
	if err != nil {
		return err
	}

	// go to project dir
	err = os.Chdir(dirPath)
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "clone", pr.GetRepoAddress(), ".")
	cmd.Env = append(os.Environ(),
		"GIT_SSH_COMMAND=ssh -i ./.ssh/key -F /dev/null",
	)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func (pr *MyPullRequest) CommunicateProgress(status string) error {
	logger := log.Default()
	logger.Println(status)
	return nil
}

// Concrete type that implements Deployment
type MyDeployment struct {
	// fields here
}

func (d *MyDeployment) NavigateToDir(dirPath string) error {
	// implementation here
	return nil
}

func (d *MyDeployment) InstallDependencies() error {
	// implementation here
	return nil
}

func (d *MyDeployment) BuildDependencies() error {
	// implementation here
	return nil
}

func (d *MyDeployment) DeployToPM2() error {
	// implementation here
	return nil
}

func (d *MyDeployment) StoreDeploymentDetails() error {
	// implementation here
	return nil
}

func (d *MyDeployment) CommunicateStatus(status string) error {
	// implementation here
	return nil
}

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
