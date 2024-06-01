package pull_request

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rssb/imbere/pkg/db"
)

// createPullRequestFromPayload creates a PullRequest from the given payload.
// It extracts the repository information, branch name, PR ID, PR number, and PR URL from the payload.
// If any of these extractions fail, it returns an error.
func CreateOrAssociatePullRequestFromPayload(event Event, payload map[string]interface{}) (*db.PullRequest, error) {
	repository, ok := payload["repository"].(map[string]interface{})
	if !ok {
		return &db.PullRequest{}, fmt.Errorf("failed to parse repository from payload")
	}

	repositoryName, ok := repository["name"].(string)
	if !ok {
		return &db.PullRequest{}, fmt.Errorf("failed to parse repository name from payload")
	}

	repositoryAddress, ok := repository["html_url"].(string)
	if !ok {
		return &db.PullRequest{}, fmt.Errorf("failed to parse repository address from payload")
	}

	sshAddress, ok := repository["ssh_url"].(string)
	if !ok {
		return &db.PullRequest{}, fmt.Errorf("failed to parse ssh address from payload")
	}

	ownerName, ownerId, err := extractRepoOwnerInfo(payload)

	if err != nil {
		return &db.PullRequest{}, err
	}

	branchName, err := extractBranchName(event, payload)
	if err != nil {
		return &db.PullRequest{}, err
	}

	prId, err := extractPRID(event, payload)
	if err != nil {
		return &db.PullRequest{}, err
	}

	prNumber, err := extractPRNumber(event, payload)
	if err != nil {
		return &db.PullRequest{}, err
	}

	prUrl, err := extractUrl(event, payload)
	if err != nil {
		return &db.PullRequest{}, err
	}

	installationID, err := extractInstallationID(event, payload)
	if err != nil {
		return &db.PullRequest{}, err
	}

	prRepo := db.PullRequestRepo{}
	// Try to get the PR by its ID
	pr, err := prRepo.GetByPrID(prId)

	if err == nil && pr == nil {
		pr = &db.PullRequest{
			PrID: prId,
		}
	} else if err != nil {
		return nil, err
	}

	// If the PR exists, update its fields
	pr.BranchName = branchName
	pr.RepoName = repositoryName
	pr.RepoAddress = repositoryAddress
	pr.SSHAddress = sshAddress
	pr.PrNumber = prNumber
	pr.PrUrl = prUrl
	pr.InstallationID = installationID
	pr.OwnerName = ownerName
	pr.OwnerID = ownerId

	return pr, nil
}

// The Event type is derived from the incoming payload. This type standardizes
// the payload structure,
func ExtractEventType(c *gin.Context, payload map[string]any) (Event, error) {
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
		return event, nil
	}

	return event, nil
}

func extractValueFromPayload(payload map[string]interface{}, path ...string) (interface{}, error) {
	var temp interface{} = payload

	for _, p := range path {
		switch v := temp.(type) {
		case map[string]interface{}:
			var ok bool
			temp, ok = v[p]
			if !ok {
				return nil, errors.New("Could not extract value - path does not exist")
			}
		case []interface{}:
			index, err := strconv.Atoi(p)
			if err != nil || index < 0 || index >= len(v) {
				return nil, errors.New("Could not extract value - invalid array index")
			}
			temp = v[index]
		default:
			return nil, errors.New("Could not extract value - path does not exist")
		}
	}

	return temp, nil
}

func extractBranchName(event Event, payload map[string]interface{}) (string, error) {
	var branchName string
	var err error
	var temp interface{}

	if event.name == "pull_request" {
		temp, err = extractValueFromPayload(payload, "pull_request", "head", "ref")
	} else if event.name == "workflow_run" {
		temp, err = extractValueFromPayload(payload, "workflow_run", "pull_requests", "0", "head", "ref")
	}

	if err != nil {
		return "", errors.New(event.GetNameAction() + " - Could not extract branch name: " + err.Error())
	}

	branchName, ok := temp.(string)
	if !ok {
		return "", errors.New(event.GetNameAction() + " - Could not extract branch name: value is not a string")
	}

	return branchName, nil
}

func extractPRID(event Event, payload map[string]interface{}) (int64, error) {
	var err error
	var temp interface{}

	if event.name == "pull_request" {
		temp, err = extractValueFromPayload(payload, "pull_request", "id")
	} else if event.name == "workflow_run" {
		temp, err = extractValueFromPayload(payload, "workflow_run", "pull_requests", "0", "id")
	}

	if err != nil {
		return 0, errors.New(event.GetNameAction() + " - Could not extract pull request id: " + err.Error())
	}

	prId, ok := temp.(float64)
	if !ok {
		return 0, errors.New(event.GetNameAction() + " - Could not extract pull request id: value is not a int64")
	}

	return int64(prId), nil
}

func extractPRNumber(event Event, payload map[string]interface{}) (int64, error) {
	var err error
	var temp interface{}

	if event.name == "pull_request" {
		temp, err = extractValueFromPayload(payload, "pull_request", "number")
	} else if event.name == "workflow_run" {
		temp, err = extractValueFromPayload(payload, "workflow_run", "pull_requests", "0", "number")
	}

	if err != nil {
		return 0, errors.New(event.GetNameAction() + " - Could not extract pull request number: " + err.Error())
	}

	prNumber, ok := temp.(float64)
	if !ok {
		return 0, errors.New(event.GetNameAction() + " - Could not extract pull request number: value is not a int64")
	}

	return int64(prNumber), nil
}

func extractLabelName(event Event, payload map[string]interface{}) (string, error) {
	var err error
	var temp interface{}

	if event.name == "pull_request" {
		temp, err = extractValueFromPayload(payload, "label", "name")
	}

	if err != nil {
		return "", errors.New(event.GetNameAction() + " - Could not extract label name : " + err.Error())
	}

	labelName, ok := temp.(string)

	if !ok {
		return "", errors.New(event.GetNameAction() + " - Could not extract label name: value is not a string")
	}

	return labelName, nil
}

func extractUrl(event Event, payload map[string]interface{}) (string, error) {
	var err error
	var temp interface{}

	if event.name == "pull_request" {
		temp, err = extractValueFromPayload(payload, "pull_request", "url")
	} else if event.name == "workflow_run" {
		temp, err = extractValueFromPayload(payload, "workflow_run", "pull_requests", "0", "url")
	}

	if err != nil {
		return "", errors.New(event.GetNameAction() + " - Could not extract pull request url: " + err.Error())
	}

	prUrl, ok := temp.(string)
	if !ok {
		return "", errors.New(event.GetNameAction() + " - Could not extract pull request url: value is not a string")
	}

	return prUrl, nil
}

func extractInstallationID(event Event, payload map[string]interface{}) (int64, error) {
	var err error
	var temp interface{}

	temp, err = extractValueFromPayload(payload, "installation", "id")

	if err != nil {
		return 0, errors.New(event.GetNameAction() + " - Could not extract installation Id: " + err.Error())
	}

	installationId, ok := temp.(float64)
	if !ok {
		return 0, errors.New(event.GetNameAction() + " - Could not extract installation Id: value is not a string")
	}

	return int64(installationId), nil
}

func extractRepoOwnerInfo(payload map[string]interface{}) (string, int64, error) {
	ownerNameValue, err := extractValueFromPayload(payload, "repository", "owner", "login")
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse owner name from payload: %v", err)
	}

	ownerName, ok := ownerNameValue.(string)
	if !ok {
		return "", 0, fmt.Errorf("failed to parse owner name from payload: value is not a string")
	}

	ownerIDValue, err := extractValueFromPayload(payload, "repository", "owner", "id")
	if err != nil {
		return "", 0, fmt.Errorf("failed to parse owner ID from payload: %v", err)
	}

	ownerID, ok := ownerIDValue.(float64)
	if !ok {
		return "", 0, fmt.Errorf("failed to parse owner ID from payload: value is not an integer")
	}

	return ownerName, int64(ownerID), nil
}
