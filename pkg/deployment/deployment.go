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
	service.log("Started Installing Dependencies")

	service.monitor.ListenToCmd(cmd)

	if err := cmd.Start(); err != nil {
		service.log(fmt.Sprintf("install command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_INSTALLING_DEPENDENCIES, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	if err := cmd.Wait(); err != nil {
		service.log(fmt.Sprintf("install command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_INSTALLING_DEPENDENCIES, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_INSTALLING_DEPENDENCIES, constants.PROCESS_OUTCOME_SUCCEEDED)
	service.log("Finished Installing Dependencies")

	return nil
}

func (service *DeploymentService) Build() error {
	cmd := exec.Command("nr", "build")
	cmd.Dir = service.WorkingDirectory()

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_BUILDING_PROJECT, constants.PROCESS_OUTCOME_ONGOING)
	service.log("Started Building")

	service.monitor.ListenToCmd(cmd)

	if err := cmd.Start(); err != nil {
		service.log(fmt.Sprintf("build command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_BUILDING_PROJECT, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	if err := cmd.Wait(); err != nil {
		service.log(fmt.Sprintf("build command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_BUILDING_PROJECT, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_BUILDING_PROJECT, constants.PROCESS_OUTCOME_SUCCEEDED)
	service.log("Finished Building")

	return nil
}

func (service *DeploymentService) Deploy() error {
	var port int32

	if service.pr.Deployed {
		port = service.pr.DeploymentPort
	} else {
		var portErr error
		port, portErr = utils.GetFreePort()

		if portErr != nil {
			service.log(fmt.Sprintf("deploy failed - failed to get port %s \n", portErr))
			service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
			return portErr
		}

	}

	err := service.deployToPM2(port)

	if err != nil {
		return err
	}

	// update db record , indicating that the pr is currently deployed
	pr, deployErr := service.prRepo.Deploy(service.pr.PrID, port)
	// Assign the latest port to the new pull request. This update will be reflected across all instances, ensuring that external clients receive the most recent port information.
	service.pr = pr

	if deployErr != nil {
		service.log(fmt.Sprintf("saving deployment status failed with %s in %s \n", deployErr, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
		return deployErr
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_COMPLETED, constants.PROCESS_OUTCOME_SUCCEEDED)
	service.log("Finished Deploying")

	return nil
}

func (service *DeploymentService) deployToPM2(port int32) error {
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
	service.log("Started Deploying")

	cmd.Dir = service.WorkingDirectory()

	service.pr.DeploymentPort = port
	service.monitor.SetPort(port)
	cmd.Env = append(os.Environ(), "PORT="+strconv.Itoa(int(port)))

	service.monitor.ListenToCmd(cmd)

	if startErr := cmd.Start(); startErr != nil {
		service.log(fmt.Sprintf("deploy command failed with %s in %s \n", startErr, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
		return startErr
	}

	if waitErr := cmd.Wait(); waitErr != nil {
		service.log(fmt.Sprintf("deploy command failed with %s in %s \n", waitErr, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
		return waitErr
	}

	return nil
}

func (service *DeploymentService) UnDeploy() error {
	err := service.unDeployFromPM2()

	if err != nil {
		return err
	}

	pr, deployErr := service.prRepo.UnDeploy(service.pr.PrID)

	if deployErr != nil {
		err := fmt.Sprintf("error while updating deployment record in db : %s", deployErr)
		service.log(err)
		return fmt.Errorf(err)
	}

	service.pr = pr
	service.log(fmt.Sprintf("successful undeployed pr ID: %d", pr.PrID))

	return nil
}

func (service *DeploymentService) log(content string) {
	service.monitor.AddLog(content)

}

func (service *DeploymentService) unDeployFromPM2() error {
	var cmd *exec.Cmd

	if service.pr.Deployed {
		cmd = exec.Command("sh", "-c", "pm2 delete "+service.pr.GetPrId())
	} else {
		err := fmt.Sprintf("there was no deployment with name %s to delete from pm2", service.pr.GetPrId())
		service.log(err)
		return fmt.Errorf(err)
	}

	service.monitor.ListenToCmd(cmd)

	if startErr := cmd.Start(); startErr != nil {
		err := fmt.Sprintf("error while starting command to undeploy on pm2 : %s", startErr)
		service.log(err)
		return fmt.Errorf(err)
	}

	if waitErr := cmd.Wait(); waitErr != nil {
		err := fmt.Sprintf("error while executing command to undeploy on pm2 : %s", waitErr)
		service.log(err)
		return fmt.Errorf(err)
	}

	return nil
}
