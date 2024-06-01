package process_monitor

import (
	"bufio"
	"fmt"
	"log"
	"os/exec"
	"strconv"

	"github.com/rssb/imbere/pkg/client"
	"github.com/rssb/imbere/pkg/constants"
	"github.com/rssb/imbere/pkg/db"
	"github.com/rssb/imbere/pkg/utils"
)

type ProcessMonitor struct {
	ID       int64
	Progress constants.ProcessProgress
	Status   constants.ProcessOutcome
	Logs     chan string
	client   *client.GithubClient
	pr       *db.PullRequest
	prRepo   *db.PullRequestRepo
}

func NewProcessMonitor(pr *db.PullRequest) *ProcessMonitor {
	prRepo := &db.PullRequestRepo{}
	processMonitor := &ProcessMonitor{
		ID:       pr.PrID,
		Progress: constants.PROCESS_PROGRESS_STARTED,
		Status:   constants.PROCESS_OUTCOME_ONGOING,
		Logs:     make(chan string),
		client:   client.NewGithubClient(pr.InstallationID),
		pr:       pr,
		prRepo:   prRepo,
	}

	processMonitor.HandleLogs() // immediately start listening to logs

	return processMonitor
}

func (p *ProcessMonitor) SetPort(port int32) {
	p.pr.DeploymentPort = port
}

func (p *ProcessMonitor) UpdateProgress(progress constants.ProcessProgress, status constants.ProcessOutcome) {
	p.Progress = progress
	p.Status = status

	appURL := "http://" + constants.IP_ADDRESS + ":" + strconv.Itoa(int(p.pr.DeploymentPort))

	progressMarkdown := utils.ParseProgressToMD(p.Progress, p.Status)
	progressMarkdown.PlainText("")
	progressMarkdown.H2("Deployment Url")
	progressMarkdown.PlainText(appURL)
	progressMarkdown.H2("Status")

	isDeployed := (p.Progress == constants.PROCESS_PROGRESS_DEPLOYING && p.Status == constants.PROCESS_OUTCOME_SUCCEEDED) || (p.Progress == constants.PROCESS_PROGRESS_COMPLETED)
	isUnDeployed := p.Progress == constants.PROCESS_PROGRESS_UN_DEPLOYING && p.Status == constants.PROCESS_OUTCOME_SUCCEEDED

	if isDeployed {
		progressMarkdown.GreenBadgef("Deployed")
	} else if p.Status == constants.PROCESS_OUTCOME_FAILED {
		progressMarkdown.RedBadgef("Failed")
	} else if isUnDeployed {
		progressMarkdown.YellowBadgef("Undeployed")
	} else {
		progressMarkdown.YellowBadgef("Deploying")
	}

	owner := p.pr.OwnerName
	repo := p.pr.RepoName
	prNumber := p.pr.PrNumber
	commentId := p.pr.CommentID

	log.Printf("CommentId: %d, Owner: %s, Repo: %s, PR Number: %d, Comment: %s\n", commentId, owner, repo, prNumber, progressMarkdown.String())

	var err error
	var id *int64
	if commentId == 0 {
		id, err = p.client.CreateComment(owner, repo, prNumber, progressMarkdown.String())
		pullRequest := p.pr

		pullRequest.CommentID = *id
		p.prRepo.Save(pullRequest)
	} else {
		id, err = p.client.EditComment(commentId, owner, repo, progressMarkdown.String())
	}

	log.Printf("Id was created %d, or Error  %s", *id, err)

	fmt.Printf("Process ID: %d, Progress: %d, Status: %d\n", p.ID, p.Progress, p.Status)
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
