package updatecheck

import (
	"fmt"
	"os"
	"os/exec"

	cfg "github.com/devfbe/gipgee/config"
)

type UpdateCheckCmd struct {
	GeneratePipeline GeneratePipelineCmd `cmd:""`
	ExecUpdateCheck  ExecUpdateCheckCmd  `cmd:""`
}

type GeneratePipelineCmd struct {
	PipelineFileName string `help:"Set the name of the pipeline file" env:"GIPGEE_UPDATE_CHECK_PIPELINE_FILENAME" default:".gipgee-gitlab-ci.yml"`
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_UPDATE_CHECK_CONFIG_FILENAME" default:"gipgee.yml"`
	GipgeeImage      string `help:"Overwrite the gipgee container image" env:"GIPGEE_OVERWRITE_GIPGEE_IMAGE" optional:""`
	SkipRebuild      bool   `help:"Just run the update check pipeline, skip the rebuild of images (used for testing)" default:"false"`
}

func (cmd *GeneratePipelineCmd) Run() error {

	config, err := cfg.LoadConfiguration(cmd.ConfigFileName)
	if err != nil {
		panic(err)
	}

	params := PipelineParams{
		SkipRebuild:    cmd.SkipRebuild,
		GipgeeImage:    cmd.GipgeeImage,
		Config:         config,
		ConfigFileName: cmd.ConfigFileName,
	}
	pipeline := GeneratePipeline(params)

	err = pipeline.WritePipelineToFile(cmd.PipelineFileName)
	if err != nil {
		panic(err)
	}

	fmt.Println("UpdateCheckCmd release")
	return nil
}

func (cmd *GeneratePipelineCmd) Help() string {
	return "Generates the update check pipeline"
}

type ExecUpdateCheckCmd struct {
	ImageId        string `arg:""`
	ConfigFileName string `required:"" env:"GIPGEE_CONFIG_FILE_NAME"`
	ResultFilePath string `required:"" env:"GIPGEE_UPDATE_CHECK_RESULT_FILE_PATH"`
}

func (cmd *ExecUpdateCheckCmd) Run() error {
	config, err := cfg.LoadConfiguration(cmd.ConfigFileName)
	if err != nil {
		panic(err)
	}
	updateCheckCommand := config.Images[cmd.ImageId].UpdateCheckCommand
	commandString := (*updateCheckCommand)[0]
	commandArgsString := make([]string, 0)
	if len(*updateCheckCommand) > 1 {
		commandArgsString = append(commandArgsString, (*updateCheckCommand)[1:]...)
	}
	commandArgsString = append(commandArgsString, cmd.ResultFilePath)
	executionCmd := exec.Command(commandString, commandArgsString...) // #nosec G204
	executionCmd.Stderr = os.Stderr
	executionCmd.Stdout = os.Stdout
	err = executionCmd.Run()
	return err
}
