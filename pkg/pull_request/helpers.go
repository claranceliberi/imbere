package pull_request

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"
)

// The Event type is derived from the incoming payload. This type standardizes 
// the payload structure,
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

// extract pull request branch from payload
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

// extract pull request id from payload
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

// extract pull request number from payload
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

// extra pull request github url from payload
func extractUrl(event Event, payload map[string]any) (string, error) {
	name := event.name
	var prUrl string
	errorMessage := errors.New(event.nameAction + " - Could not extract PR URL")

	if name == "pull_request" {
		pullRequest, ok := payload["pull_request"].(map[string]any)
		if !ok {
			return "", errorMessage
		}

		prUrl, _ = pullRequest["url"].(string)
	}

	if name == "workflow_run" {
		workflowRun, ok := payload["workflow_run"].(map[string]any)
		if !ok {
			return "", errorMessage
		}
		pullRequests, ok := workflowRun["pull_requests"].([]interface{})
		if !ok || len(pullRequests) == 0 {
			return "", errorMessage

		}

		pullRequest, ok := pullRequests[0].(map[string]any)
		if !ok {
			return "", errorMessage
		}
		prUrl, _ = pullRequest["url"].(string)
	}

	return prUrl, nil
}
