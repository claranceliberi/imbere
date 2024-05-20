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

func (service *PullRequestService) PrepareDir() (string, error) {
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

func (service *PullRequestService) PullChanges() error {
	_, err := exec.LookPath("git")

	if err != nil {
		return err
	}

	// prepare cloning dir
	dirPath, err := service.PrepareDir()
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

	return nil
}

func (pr *PullRequest) CommunicateProgress(status string) error {
	logger := log.Default()
	logger.Println(status)
	return nil
}
