package utils

import (
	"fmt"
	"log"
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	md "github.com/nao1215/markdown"
	"github.com/phayes/freeport"
	"github.com/rssb/imbere/pkg/constants"
)

func ReturnError(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, gin.H{
		"message": message,
	})
	debug.PrintStack()
	fmt.Println("Error occured %+v\n", message)

}

func GetFreePort() (int32, error) {
	port, err := freeport.GetFreePort()
	if err != nil {
		log.Fatal(err)
		return 0, err
	}

	log.Printf("port %d", port)

	return int32(port), nil
}


func ParseProgressToMD(progress constants.ProcessProgress, outcome constants.ProcessOutcome) *md.Markdown {
	markdown := md.NewMarkdown(nil)
	markdown.H3("Progress Status")

	progressSteps := []constants.ProcessProgress{
		constants.PROCESS_PROGRESS_STARTED,
		constants.PROCESS_PROGRESS_PREPARING_DIR,
		constants.PROCESS_PROGRESS_PULLING_CHANGES,
		constants.PROCESS_PROGRESS_INSTALLING_DEPENDENCIES,
		constants.PROCESS_PROGRESS_BUILDING_PROJECT,
		constants.PROCESS_PROGRESS_DEPLOYING,
		constants.PROCESS_PROGRESS_COMPLETED,
	}

	for _, step := range progressSteps {
		if step == progress {
			switch outcome {
			case constants.PROCESS_OUTCOME_SUCCEEDED:
				markdown.PlainTextf(fmt.Sprintf("✅ %s", GetProgressStepName(step)))
			case constants.PROCESS_OUTCOME_FAILED:
				markdown.PlainTextf(fmt.Sprintf("❌ %s", GetProgressStepName(step)))
			case constants.PROCESS_OUTCOME_ONGOING:
				markdown.PlainTextf(fmt.Sprintf("⏳ %s", GetProgressStepName(step)))
			}
		} else if step < progress {
			markdown.PlainTextf(fmt.Sprintf("✅ %s", GetProgressStepName(step)))
		} else {
			markdown.PlainTextf(fmt.Sprintf("⚪ %s", GetProgressStepName(step)))
		}
	}

	return markdown
}

func GetProgressStepName(step constants.ProcessProgress) string {
	switch step {
	case constants.PROCESS_PROGRESS_STARTED:
		return "Started"
	case constants.PROCESS_PROGRESS_PREPARING_DIR:
		return "Preparing Directory"
	case constants.PROCESS_PROGRESS_PULLING_CHANGES:
		return "Pulling Changes"
	case constants.PROCESS_PROGRESS_INSTALLING_DEPENDENCIES:
		return "Installing Dependencies"
	case constants.PROCESS_PROGRESS_BUILDING_PROJECT:
		return "Building Project"
	case constants.PROCESS_PROGRESS_DEPLOYING:
		return "Deploying"
	case constants.PROCESS_PROGRESS_COMPLETED:
		return "Completed"
	default:
		return "Unknown"
	}
}
