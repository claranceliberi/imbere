package event_monitor

import (
	"github.com/gin-gonic/gin"
	"github.com/rssb/imbere/pkg/db"
)

type GitHubEventMonitor struct {
}

func (s *GitHubEventMonitor) Webhook(c *gin.Context) (*db.PullRequest, error) {
	return nil, nil
}
