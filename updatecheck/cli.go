package updatecheck

import "fmt"

type UpdateCheckCmd struct {
	PipelineFileName string `help:"Set the name of the pipeline file" env:"GIPGEE_UPDATE_CHECK_PIPELINE_FILENAME" default:".gipgee-gitlab-ci.yml"`
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_UPDATE_CHECK_CONFIG_FILENAME" default:"gipgee.yml"`
	GipgeeImage      string `help:"Overwrite the gipgee container image" env:"GIPGEE_OVERWRITE_GIPGEE_IMAGE" optional:""`
}

func (r *UpdateCheckCmd) Run() error {
	fmt.Println("UpdateCheckCmd release")
	return nil
}

func (r *UpdateCheckCmd) Help() string {
	return "Generates the update check pipeline"
}
