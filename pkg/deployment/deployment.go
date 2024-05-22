package deployment

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/rssb/imbere/pkg/constants"
	"github.com/rssb/imbere/pkg/db"
	"github.com/rssb/imbere/pkg/process_monitor"
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
	cmd.Dir = service.WorkingDirectory()
	port := int32(3333)
	cmd.Env = append(os.Environ(), "PORT="+string(port))

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_ONGOING)
	service.monitor.AddLog("Started Deploying")

	service.monitor.ListenToCmd(cmd)

	if err := cmd.Start(); err != nil {
		service.monitor.AddLog(fmt.Sprintf("deploy command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	if err := cmd.Wait(); err != nil {
		service.monitor.AddLog(fmt.Sprintf("deploy command failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	// update db record , indicating that the pr is currently deployed
	_, err := service.prRepo.Deploy(service.pr.PrID, port)

	if err != nil {
		service.monitor.AddLog(fmt.Sprintf("saving deployment status failed with %s in %s \n", err, service.WorkingDirectory()))
		service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_FAILED)
		return err
	}

	service.monitor.UpdateProgress(constants.PROCESS_PROGRESS_DEPLOYING, constants.PROCESS_OUTCOME_SUCCEEDED)
	service.monitor.AddLog("Finished Deploying")

	return nil
}
