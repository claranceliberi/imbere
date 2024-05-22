package constants

const MAIN_DIR = "/Users/claranceliberi/projects/rssb/imbere/"
const BUILD_DIR = MAIN_DIR + "builds/"

const PM2_NAMESPACE = "IMBERE"

// IMBERE2 github app id
const GITHUB_APP_ID = 903361

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
)

type ProcessOutcome int

const (
	PROCESS_OUTCOME_ONGOING ProcessOutcome = iota
	PROCESS_OUTCOME_SUCCEEDED
	PROCESS_OUTCOME_FAILED
)