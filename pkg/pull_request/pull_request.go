package pull_request

import (
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/rssb/imbere/pkg/constants"
	"github.com/rssb/imbere/pkg/db"
	"github.com/rssb/imbere/pkg/deployment"
	"github.com/rssb/imbere/pkg/process_monitor"
)

// Concrete type that implements PullRequest

type Event struct {
	name   string
	action string
}

// gives a combination of event name(type) and action
// ie. workflow_run.completed or pull_request.opened
func (event *Event) GetNameAction() string {
	return event.name + "." + event.action
}

type PullRequestService struct {
	pr      *db.PullRequest
	monitor *process_monitor.ProcessMonitor
}

func NewPullRequestService(pr *db.PullRequest, processMonitor *process_monitor.ProcessMonitor) *PullRequestService {

	return &PullRequestService{
		pr:      pr,
		monitor: processMonitor,
	}
}

func (service *PullRequestService) log(content string) {
	service.monitor.AddLog(content)
}

func (service *PullRequestService) removeDir() error {
	dirPath := constants.BUILD_DIR + service.pr.GetDir()

	if _, err := os.Stat(dirPath); !os.IsNotExist(err) {
		removeDirErr := os.RemoveAll(dirPath)

		if removeDirErr != nil {
			service.log(fmt.Sprintf("Failed to remove directory: %s", err.Error()))
			return removeDirErr
		}
	}
	service.log(fmt.Sprintf("Directory %s  removed successfully", dirPath))
	return nil
}

// When a pull request (PR) is created, a corresponding directory is generated that contains the changes introduced by the PR.
// If a directory for the PR already exists, it is first deleted and then recreated to reflect the latest changes.
// This approach ensures that any new pushes to the PR are incorporated, keeping the deployed code up-to-date.
// The directory can later be deployed to any environment, enabling continuous integration and delivery.
func (service *PullRequestService) prepareDir() (string, error) {
	dirPath := constants.BUILD_DIR + service.pr.GetDir()

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_PREPARING_DIR, constants.PROCESS_OUTCOME_ONGOING)

	// check if dir exists and remove it
	if err := service.removeDir(); err != nil {
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_PREPARING_DIR, constants.PROCESS_OUTCOME_FAILED)
		return "", err
	}

	// create dir
	err := os.MkdirAll(dirPath, 0755)
	if err != nil {
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_PREPARING_DIR, constants.PROCESS_OUTCOME_FAILED)
		service.log(fmt.Sprintf("Failed to create directory: %s", err.Error()))
		return "", err
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_PREPARING_DIR, constants.PROCESS_OUTCOME_SUCCEEDED)
	service.log(fmt.Sprintf("Directory %s created successfully", dirPath))

	return dirPath, nil
}

// Saves pr information in database, if Pull Request (PR ) already exists it will be updated
// we track PR by ts pr_id
func (service *PullRequestService) save() error {
	prRepo := db.PullRequestRepo{}
	return prRepo.Save(service.pr)
}

// In this section, we perform three key operations:
// 1. Create a new directory for the incoming pull request.
// 2. Pull the latest changes from the pull request into this directory.
// 3. Save the current state of the pull request in the database.
// These steps ensure that we have the most recent code changes isolated in a separate directory and the pull request's status is accurately tracked in the database.
func (service *PullRequestService) PullChanges() error {
	_, err := exec.LookPath("git")
	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_STARTED, constants.PROCESS_OUTCOME_ONGOING)

	// set pr as active, this is to indicate that the pr is currently being processed
	// this is important to avoid multiple processing of the same pr
	service.pr.IsDeploying = true
	service.save()

	if err != nil {
		service.log(fmt.Sprintf("Git not found: %s", err.Error()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_STARTED, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	// prepare cloning dir
	// the process about directory creation is communicated inside this method
	dirPath, err := service.prepareDir()
	if err != nil {
		return err
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_PULLING_CHANGES, constants.PROCESS_OUTCOME_ONGOING)

	cmd := exec.Command("git", "clone", "-b", service.pr.BranchName, "--single-branch", service.pr.RepoAddress, dirPath)
	cmd.Env = append(os.Environ(),
		"GIT_SSH_COMMAND=ssh -i ./.ssh/key -F /dev/null",
	)

	service.monitor.ListenToCmd(cmd) // listen for outcome

	err = cmd.Start()
	if err != nil {
		service.log(fmt.Sprintf("Failed to clone repository: %s", err.Error()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_PULLING_CHANGES, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	err = cmd.Wait()
	if err != nil {
		service.log(fmt.Sprintf("Failed to clone repository: %s", err.Error()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_PULLING_CHANGES, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	service.log("Repository cloned successfully")
	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_PULLING_CHANGES, constants.PROCESS_OUTCOME_SUCCEEDED)

	err = service.save()
	if err != nil {
		return err
	}

	return nil
}

func (service *PullRequestService) MarkActive() {
	service.pr.Active = true
	service.save()
}

func (service *PullRequestService) Deploy() error {
	if service.pr.IsDeploying {
		log.Println("There is a deployment in progress for this PR ID: %s, skipping...", service.pr.GetPrId())
	} else {

		err := service.PullChanges()

		if err != nil {
			return err
		}

		deploymentService := deployment.NewDeploymentService(service.pr, service.monitor)

		err = deploymentService.InstallDependencies()
		if err != nil {
			return err
		}

		err = deploymentService.Build()
		if err != nil {
			return err
		}

		err = deploymentService.Deploy()
		if err != nil {
			return err
		}

	}

	return nil
}

func (service *PullRequestService) UnDeploy() error {

	err := service.removeDir()

	if err != nil {
		return err
	}

	deploymentService := deployment.NewDeploymentService(service.pr, service.monitor)

	return deploymentService.UnDeploy()
}

func (service *PullRequestService) UpdateLabelToDeploy(isLabelPresent bool) error {
	service.pr.LabeledToDeploy = isLabelPresent

	if isLabelPresent && !service.pr.Deployed {
		service.Deploy()
	}

	return service.save()
}

func HandlePR(event Event, payload map[string]interface{}) error {

	PR, err := CreateOrAssociatePullRequestFromPayload(event, payload)

	if err != nil {
		return err
		// utils.ReturnError(c, err.Error())
	}

	nameAction := event.GetNameAction()
	isPullRequestOpenedOrReopened := nameAction == "pull_request.opened" || nameAction == "pull_request.reopened"
	isWorkflowRunCompleted := nameAction == "workflow_run.completed"
	shouldDeploy := isWorkflowRunCompleted && PR.LabeledToDeploy
	isPullRequestClosed := nameAction == "pull_request.closed"
	isPullRequestLabeled := nameAction == "pull_request.labeled"
	isPullRequestUnlabeled := nameAction == "pull_request.unlabeled"
	processMonitor := process_monitor.NewProcessMonitor(PR)

	prService := NewPullRequestService(PR, processMonitor)

	if isPullRequestOpenedOrReopened {
		
	} else if shouldDeploy {
		return prService.Deploy()
	} else if isPullRequestClosed {
		return prService.UnDeploy()
	} else if isPullRequestLabeled || isPullRequestUnlabeled {
		labelName, err := extractLabelName(event, payload)
		if err != nil {
			return err
		}

		if labelName == constants.DEPLOYMENT_LABEL {
			if isPullRequestLabeled {
				return prService.UpdateLabelToDeploy(true)
			} else if isPullRequestUnlabeled {
				return prService.UpdateLabelToDeploy(false)
			}
		}

	}

	return nil
}

func CommunicateProgress(status string) error {
	logger := log.Default()
	logger.Println(status)
	return nil
}
