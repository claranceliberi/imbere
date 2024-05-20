package pull_request

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

// Concrete type that implements PullRequest

type Event struct {
	name       string
	action     string
	nameAction string
}

type PullRequestService struct {
	pr *PullRequest
}

// When a pull request (PR) is created, a corresponding directory is generated that contains the changes introduced by the PR.
// If a directory for the PR already exists, it is first deleted and then recreated to reflect the latest changes.
// This approach ensures that any new pushes to the PR are incorporated, keeping the deployed code up-to-date.
// The directory can later be deployed to any environment, enabling continuous integration and delivery.
func (service *PullRequestService) prepareDir() (string, error) {
	mainDir := "/Users/claranceliberi/projects/rssb/imbere"
	prIDStr := fmt.Sprintf("%d", int(service.pr.PrNumber))
	// dirPath := "/var/lib/imbere/builds/" + service.pr.GetRepoName() + "/" + service.pr.GetBranchName() + "_" + prIDStr
	dirPath := mainDir + "/builds/" + service.pr.RepoName + "/" + service.pr.BranchName + "_" + prIDStr

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

// Saves pr information in database, if Pull Request (PR ) already exists it will be updated
// we track PR by ts pr_id
func (service *PullRequestService) save() error {
	prRepo := PullRequestRepo{}
	return prRepo.Save(service.pr)
}

// In this section, we perform three key operations:
// 1. Create a new directory for the incoming pull request.
// 2. Pull the latest changes from the pull request into this directory.
// 3. Save the current state of the pull request in the database.
// These steps ensure that we have the most recent code changes isolated in a separate directory and the pull request's status is accurately tracked in the database.
func (service *PullRequestService) PullChanges() error {
	_, err := exec.LookPath("git")

	if err != nil {
		return err
	}

	// prepare cloning dir
	dirPath, err := service.prepareDir()
	if err != nil {
		return err
	}

	cmd := exec.Command("git", "clone", service.pr.RepoAddress, dirPath)
	cmd.Env = append(os.Environ(),
		"GIT_SSH_COMMAND=ssh -i ./.ssh/key -F /dev/null",
	)

	err = cmd.Run()
	if err != nil {
		return err
	}

	err = service.save()

	if err != nil {
		return err
	}

	return nil
}

func (pr *PullRequest) CommunicateProgress(status string) error {
	logger := log.Default()
	logger.Println(status)
	return nil
}
