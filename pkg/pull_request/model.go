package pull_request

import (
	"github.com/rssb/imbere/pkg/db"
	"gorm.io/gorm"
)

type PullRequestRepo struct {
	db *gorm.DB
}

type PullRequest struct {
	gorm.Model
	PrID              float64 `gorm:"type:float;not null"`
	PrNumber          float64 `gorm:"type:float;not null"`
	BranchName        string  `gorm:"type:text;not null"`
	PrUrl             string  `gorm:"type:text;not null"`
	RepoName          string  `gorm:"type:text;not null"`
	RepoAddress       string  `gorm:"type:text;not null"`
	SSHAddress        string  `gorm:"type:text;not null"`
	WorkflowSucceeded bool    `gorm:"type:bool;not null;default:false"`
	LabeledToDeploy   bool    `gorm:"type:bool;not null;default:false"`
	Active            bool    `gorm:"type:bool;not null;default:false"`
}

func (repo *PullRequestRepo) prepareDbConnection() {
	repo.db = db.DbCon()
}

func (repo *PullRequestRepo) Save(pr *PullRequest) error {
	repo.prepareDbConnection()

	result := repo.db.Where(PullRequest{PrID: pr.PrID}).Assign(PullRequest{
		PrNumber:          pr.PrNumber,
		BranchName:        pr.BranchName,
		PrUrl:             pr.PrUrl,
		RepoName:          pr.RepoName,
		RepoAddress:       pr.RepoAddress,
		SSHAddress:        pr.SSHAddress,
		WorkflowSucceeded: pr.WorkflowSucceeded,
		LabeledToDeploy:   pr.LabeledToDeploy,
		Active:            pr.Active,
	}).FirstOrCreate(pr)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (repo *PullRequestRepo) GetByPrID(prId float64) (*PullRequest, error) {
	repo.prepareDbConnection()

	var pr PullRequest

	result := repo.db.Where("pr_id", prId).First(&pr)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, result.Error
	}

	return &pr, nil
}
