package deployment

// Deployment interface
type Deployment interface {
	NavigateToDir(dirPath string) error
	InstallDependencies() error
	BuildDependencies() error
	DeployToPM2() error
	StoreDeploymentDetails() error
	CommunicateStatus(status string) error
}

// Concrete type that implements Deployment
type MyDeployment struct {
	// fields here
}

func (d *MyDeployment) NavigateToDir(dirPath string) error {
	// implementation here
	return nil
}

func (d *MyDeployment) InstallDependencies() error {
	// implementation here
	return nil
}

func (d *MyDeployment) BuildDependencies() error {
	// implementation here
	return nil
}

func (d *MyDeployment) DeployToPM2() error {
	// implementation here
	return nil
}

func (d *MyDeployment) StoreDeploymentDetails() error {
	// implementation here
	return nil
}

func (d *MyDeployment) CommunicateStatus(status string) error {
	// implementation here
	return nil
}
