package updatecheck

import (
	"fmt"
	cfg "github.com/devfbe/gipgee/config"
)

type UpdateCheckCmd struct {
	PipelineFileName string `help:"Set the name of the pipeline file" env:"GIPGEE_UPDATE_CHECK_PIPELINE_FILENAME" default:".gipgee-gitlab-ci.yml"`
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_UPDATE_CHECK_CONFIG_FILENAME" default:"gipgee.yml"`
	GipgeeImage      string `help:"Overwrite the gipgee container image" env:"GIPGEE_OVERWRITE_GIPGEE_IMAGE" optional:""`
	SkipRebuild      bool   `help:"Just run the update check pipeline, skip the rebuild of images (used for testing)" default:"false"`
	ReleaseOrStaging string `enum:"release,staging" default:"release" help:"Which images should be checked for updates? Default: the released images, not the latest from the staging registry"`
}

func (cmd *UpdateCheckCmd) Run() error {

	config, err := cfg.LoadConfiguration(cmd.ConfigFileName)
	if err != nil {
		panic(err)
	}

	params := PipelineParams{
		SkipRebuild:      cmd.SkipRebuild,
		GipgeeImage:      cmd.GipgeeImage,
		Config:           config,
		releaseOrStaging: cmd.ReleaseOrStaging,
	}
	pipeline := GeneratePipeline(params)

	err = pipeline.WritePipelineToFile(cmd.PipelineFileName)
	if err != nil {
		panic(err)
	}

	fmt.Println("UpdateCheckCmd release")
	return nil
}

func (cmd *UpdateCheckCmd) Help() string {
	return "Generates the update check pipeline"
}
