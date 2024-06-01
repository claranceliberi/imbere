package constants

const MAIN_DIR = "/Users/claranceliberi/projects/rssb/imbere/"
const BUILD_DIR = MAIN_DIR + "builds/"

const PM2_NAMESPACE = "IMBERE"

// This is the label text that will be added to github PR if they want it to be deployed
const DEPLOYMENT_LABEL = "IMBERE_DEPLOY"

// IMBERE2 github app id
const GITHUB_APP_ID = 903361

const IP_ADDRESS = "localhost"

type ProcessProgress int

const (
	PROCESS_PROGRESS_UNKNOWN ProcessProgress = iota
	PROCESS_PROGRESS_STARTED
	PROCESS_PROGRESS_PREPARING_DIR
	PROCESS_PROGRESS_PULLING_CHANGES
	PROCESS_PROGRESS_INSTALLING_DEPENDENCIES
	PROCESS_PROGRESS_BUILDING_PROJECT
	PROCESS_PROGRESS_DEPLOYING
	PROCESS_PROGRESS_COMPLETED
	PROCESS_PROGRESS_UN_DEPLOYING
)

type ProcessOutcome int

const (
	PROCESS_OUTCOME_NOT_YET ProcessOutcome = iota
	PROCESS_OUTCOME_ONGOING
	PROCESS_OUTCOME_SUCCEEDED
	PROCESS_OUTCOME_FAILED
)

var ALLOWED_EVENT_ACTIONS = map[string]bool{
	"workflow_run.completed": true,
	"pull_request.closed":    true,
	"pull_request.opened":    true,
	"pull_request.reopened":  true,
	"pull_request.labeled":   true,
	"pull_request.unlabeled": true,
}
