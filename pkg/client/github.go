package client

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/github"
	"github.com/rssb/imbere/pkg/constants"
)

type GithubClient struct {
	client *github.Client
}

func NewGithubClient(installationID int64) *GithubClient {
	// Shared transport to reuse TCP connections.
	tr := http.DefaultTransport

	// Wrap the shared transport for use with the app ID 1 authenticating with installation ID 99.
	itr, err := ghinstallation.NewKeyFromFile(tr, constants.GITHUB_APP_ID, installationID, "keys/Imbere2_private_key.pem")
	if err != nil {
		log.Fatal(err)
	}

	// Use installation transport with github.com/google/go-github
	ghClient := github.NewClient(&http.Client{Transport: itr})

	client := GithubClient{
		client: ghClient,
	}

	return &client
}

func (gc *GithubClient) CreateComment(owner string, repo string, number int64, content string) (*int64, error) {

	comment := github.IssueComment{
		Body: &content,
	}

	log.Printf("creating comment")

	prComment, _, err := gc.client.Issues.CreateComment(context.Background(), owner, repo, int(number), &comment)

	if err != nil {
		return nil, fmt.Errorf("Could not create comment on pull request %v", err)
	}

	log.Printf("Comment created with ID: %d\n", *prComment.ID)

	return prComment.ID, nil

}

func (gc *GithubClient) EditComment(id int64, owner string, repo string, content string) (*int64, error) {

	comment := github.IssueComment{
		Body: &content,
	}

	prComment, _, err := gc.client.Issues.EditComment(context.Background(), owner, repo, id, &comment)

	if err != nil {
		return nil, fmt.Errorf("Could not create comment on pull request %v", err)
	}

	log.Printf("Comment edited with ID: %d\n", *prComment.ID)
	return prComment.ID, nil

}
