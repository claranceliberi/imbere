package process_monitor

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"

	"github.com/rssb/imbere/pkg/client"
	"github.com/rssb/imbere/pkg/constants"
	"github.com/rssb/imbere/pkg/db"
)

type ProcessMonitor struct {
	ID       int64
	Progress constants.ProcessProgress
	Status   constants.ProcessOutcome
	Logs     chan string
	client   *client.GithubClient
	pr       *db.PullRequest
}

func NewProcessMonitor(pr *db.PullRequest) *ProcessMonitor {

	processMonitor := &ProcessMonitor{
		ID:       pr.PrID,
		Progress: constants.PROCESS_PROGRESS_STARTED,
		Status:   constants.PROCESS_OUTCOME_ONGOING,
		Logs:     make(chan string),
		client:   client.NewGithubClient(pr.InstallationID),
		pr:       pr,
	}

	processMonitor.HandleLogs() // immediately start listening to logs

	return processMonitor
}

func (p *ProcessMonitor) UpdateProgress(progress constants.ProcessProgress, status constants.ProcessOutcome) {
	p.Progress = progress
	p.Status = status

	appURL := "http://localhost" // replace with your app's URL
	comment := fmt.Sprintf("[App Link](%s:%d)", appURL, p.pr.DeploymentPort)

	owner := p.pr.OwnerName
	repo := p.pr.RepoName
	prNumber := p.pr.PrNumber
	commentId := p.pr.CommentID

	if p.Progress == constants.PROCESS_PROGRESS_DEPLOYING && p.Status == constants.PROCESS_OUTCOME_SUCCEEDED {
		log.Printf("CommentId: %d, Owner: %s, Repo: %s, PR Number: %d, Comment: %s\n", commentId, owner, repo, prNumber, comment)
		var err error
		var id *int64
		if commentId == 0 {
			id, err = p.client.CreateComment(owner, repo, prNumber, comment)
		} else {
			id, err = p.client.EditComment(commentId, owner, repo, comment)
		}

		log.Printf("Id was created %s, or Error  %s", id, err)
	}

	fmt.Printf("Process ID: %s, Progress: %s, Status: %d\n", p.ID, p.Progress, p.Status)
	// To communicate the status to github
}

func (p *ProcessMonitor) AddLog(log string) {
	p.Logs <- log
}

func (p *ProcessMonitor) HandleLogs() {
	go func() {
		for log := range p.Logs {
			fmt.Printf("Process ID: %s, Log: %s\n", p.ID, log)
			// just in case I might want to save logs or display them
		}
	}()
}

func (p *ProcessMonitor) ListenToCmd(cmd *exec.Cmd) {
	stdout, _ := cmd.StdoutPipe()
	stderr, _ := cmd.StderrPipe()

	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			p.AddLog(scanner.Text())
		}
	}()

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			p.AddLog(scanner.Text())
		}
	}()

}
