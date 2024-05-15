package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"

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
	var payload map[string]interface{}

	if err := c.ShouldBindJSON(&payload); err != nil {
		fmt.Println("Error occured %+v\n", err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	}

	// get event type
	eventType, err := extractEventType(c, payload)

	if err != nil {
		c.JSON(http.StatusAccepted, gin.H{
			"message": err.Error(),
		})
		return
	}

	repository, ok := payload["repository"].(map[string]interface{})
	if !ok {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "missing repository data",
		})
		return
	}

	repositoryName, ok := repository["name"].(string)
	if !ok {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "missing repository name",
		})
		return
	}

	repositoryAddress, ok := repository["html_url"].(string)
	if !ok {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "missing repository address",
		})
		return
	}

	sshAddress, ok := repository["ssh_url"].(string)
	if !ok {
		c.JSON(http.StatusAccepted, gin.H{
			"message": "missing ssh address",
		})
		return
	}

	fmt.Println("Payload: %+v\n", payload)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "not yet started",
	})
}

// type MyPullRequest struct {
// 	branchName  string
// 	repoName    string
// 	repoAddress string
// 	sshAddress  string
// 	prID        string
// }

type Event struct {
	name       string
	action     string
	nameAction string
}

func extractEventType(c *gin.Context, payload map[string]interface{}) (Event, error) {
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
	GetPRID() string
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
	prID        string
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

func (pr *MyPullRequest) GetPRID() string {
	return pr.repoAddress
}
func (pr *MyPullRequest) PrepareDir() (string, error) {
	dirPath := "/var/lib/imbere/" + pr.GetRepoName() + "/" + pr.GetBranchName() + "_" + pr.GetPRID()

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
// ID, PR(string), BRANCH(string), URL(string), WORKFLOW_SUCCEEDED(1,0), LABELED_TO_DEPLOY(1,0), ACTIVE(1,0), ....INFORMATION FROM PM2
