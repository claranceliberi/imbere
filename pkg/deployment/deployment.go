package deployment

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/rssb/imbere/pkg/constants"
	"github.com/rssb/imbere/pkg/db"
	"github.com/rssb/imbere/pkg/process_monitor"
	"github.com/rssb/imbere/pkg/utils"
)

type DeploymentService struct {
	pr      *db.PullRequest
	prRepo  db.PullRequestRepo
	monitor *process_monitor.ProcessMonitor
}

func NewDeploymentService(pr *db.PullRequest, monitor *process_monitor.ProcessMonitor) *DeploymentService {
	return &DeploymentService{
		pr:      pr,
		monitor: monitor,
	}
}

func (service *DeploymentService) WorkingDirectory() string {
	return constants.BUILD_DIR + service.pr.GetDir()
}
func (service *DeploymentService) InstallDependencies() error {
	cmd := exec.Command("ni")
	cmd.Dir = service.WorkingDirectory()

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_INSTALLING_DEPENDENCIES, constants.PROCESS_OUTCOME_ONGOING)
	service.monitor.AddLog("Started Installing Dependencies")

	service.monitor.ListenToCmd(cmd)

	if err := cmd.Start(); err != nil {
		service.monitor.AddLog(fmt.Sprintf("install command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_INSTALLING_DEPENDENCIES, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	if err := cmd.Wait(); err != nil {
		service.monitor.AddLog(fmt.Sprintf("install command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_INSTALLING_DEPENDENCIES, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_INSTALLING_DEPENDENCIES, constants.PROCESS_OUTCOME_SUCCEEDED)
	service.monitor.AddLog("Finished Installing Dependencies")

	return nil
}

func (service *DeploymentService) Build() error {
	cmd := exec.Command("nr", "build")
	cmd.Dir = service.WorkingDirectory()

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_BUILDING_PROJECT, constants.PROCESS_OUTCOME_ONGOING)
	service.monitor.AddLog("Started Building")

	service.monitor.ListenToCmd(cmd)

	if err := cmd.Start(); err != nil {
		service.monitor.AddLog(fmt.Sprintf("build command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_BUILDING_PROJECT, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	if err := cmd.Wait(); err != nil {
		service.monitor.AddLog(fmt.Sprintf("build command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_BUILDING_PROJECT, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_BUILDING_PROJECT, constants.PROCESS_OUTCOME_SUCCEEDED)
	service.monitor.AddLog("Finished Building")

	return nil
}

func (service *DeploymentService) DeployToPM2() error {
	var cmd *exec.Cmd

	// If there's an active deployment, we restart it. This approach ensures
	// that we only have a single instance of the app running, even when there
	// are changes to the pull request.
	if service.pr.Deployed {
		cmd = exec.Command("sh", "-c", "pm2 restart "+service.pr.GetPrId())
	} else {
		cmd = exec.Command("sh", "-c", "pm2 start 'nr start' --name "+service.pr.GetPrId()+" --namespace "+constants.PM2_NAMESPACE)
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_ONGOING)
	service.monitor.AddLog("Started Deploying")

	cmd.Dir = service.WorkingDirectory()
	var port int32

	if service.pr.Deployed {
		port = service.pr.DeploymentPort
	} else {
		var portErr error
		port, portErr = utils.GetFreePort()

		if portErr != nil {
			service.monitor.AddLog(fmt.Sprintf("deploy failed - failed to get port %s \n", portErr))
			service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
			return portErr
		}
	}

	cmd.Env = append(os.Environ(), "PORT="+strconv.Itoa(int(port)))

	service.monitor.ListenToCmd(cmd)

	if startErr := cmd.Start(); startErr != nil {
		service.monitor.AddLog(fmt.Sprintf("deploy command failed with %s in %s \n", startErr, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
		return startErr
	}

	if waitErr := cmd.Wait(); waitErr != nil {
		service.monitor.AddLog(fmt.Sprintf("deploy command failed with %s in %s \n", waitErr, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
		return waitErr
	}

	// update db record , indicating that the pr is currently deployed
	pr, deployErr := service.prRepo.Deploy(service.pr.PrID, port)
	// Assign the latest port to the new pull request. This update will be reflected across all instances, ensuring that external clients receive the most recent port information.
	service.pr = pr 

	if deployErr != nil {
		service.monitor.AddLog(fmt.Sprintf("saving deployment status failed with %s in %s \n", deployErr, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
		return deployErr
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_SUCCEEDED)
	service.monitor.AddLog("Finished Deploying")

	return nil
}
