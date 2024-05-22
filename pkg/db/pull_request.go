package db;

import (
	"fmt"

	"gorm.io/gorm"
)

type PullRequestRepo struct {
	db *gorm.DB
}

type PullRequest struct {
	gorm.Model
	PrID              int64  `gorm:"type:bigint;not null"`
	PrNumber          int64  `gorm:"type:bigint;not null"`
	BranchName        string `gorm:"type:text;not null"`
	PrUrl             string `gorm:"type:text;not null"`
	RepoName          string `gorm:"type:text;not null"`
	RepoAddress       string `gorm:"type:text;not null"`
	SSHAddress        string `gorm:"type:text;not null"`
	InstallationID    int64  `gorm:"type:bigint;not null"`
	OwnerName         string `gorm:"type:text;not null"`
	OwnerID           int64  `gorm:"type:bigint;not null"`
	CommentID         int64  `gorm:"type:bigint;not null;default:0"`
	WorkflowSucceeded bool   `gorm:"type:bool;not null;default:false"` // did workflow succeed from github
	LabeledToDeploy   bool   `gorm:"type:bool;not null;default:false"` // is PR labeled to be deployed on github
	Active            bool   `gorm:"type:bool;not null;default:false"` // active pull request
	Deployed          bool   `gorm:"type:bool;not null;default:false"` // is deployed (accessible over internet)
	DeploymentPort    int32  `gorm:"type:bigint"`                      // deployment service port
}

func (pr *PullRequest) GetPrId() string {
	return fmt.Sprintf("%d", int(pr.PrID))
}

func (pr *PullRequest) GetPrNumber() string {
	return fmt.Sprintf("%d", int(pr.PrNumber))
}

func (pr *PullRequest) GetDir() string {
	return pr.RepoName + "/" + pr.BranchName + "_" + pr.GetPrNumber()
}

func (repo *PullRequestRepo) prepareDbConnection() {
	repo.db = dbCon()
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
		InstallationID:    pr.InstallationID,
		WorkflowSucceeded: pr.WorkflowSucceeded,
		LabeledToDeploy:   pr.LabeledToDeploy,
		Active:            pr.Active,
		Deployed:          pr.Deployed,
		DeploymentPort:    pr.DeploymentPort,
		OwnerName:         pr.OwnerName,
		OwnerID:           pr.OwnerID,
		CommentID:         pr.CommentID,
	}).FirstOrCreate(pr)

	if result.Error != nil {
		return result.Error
	}

	return nil
}

func (repo *PullRequestRepo) GetByPrID(prId int64) (*PullRequest, error) {
	repo.prepareDbConnection()

	var pr PullRequest

	result := repo.db.Where("pr_id", prId)

	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}

		return nil, result.Error
	}

	result.First(&pr)

	return &pr, nil
}

func (repo *PullRequestRepo) Deploy(prId int64, port int32) (*PullRequest, error) {

	pr, err := repo.GetByPrID(prId)

	if err != nil {
		return nil, err
	}

	pr.Deployed = true
	pr.DeploymentPort = port

	err = repo.Save(pr)

	if err != nil {
		return nil, err
	}

	return pr, nil
}

func (repo *PullRequestRepo) UnDeploy(prId int64) (*PullRequest, error) {

	pr, err := repo.GetByPrID(prId)

	if err != nil {
		return nil, err
	}

	pr.Deployed = false
	pr.DeploymentPort = 0

	err = repo.Save(pr)

	if err != nil {
		return nil, err
	}

	return pr, nil
}
