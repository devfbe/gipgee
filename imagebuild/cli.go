package imagebuild

import (
	"os"
	"os/exec"

	cfg "github.com/devfbe/gipgee/config"
)

type ImageBuildCmd struct {
	GenerateKanikoAuth   GenerateKanikoAuthCmd   `cmd:""`
	GeneratePipeline     GeneratePipelineCmd     `cmd:""`
	ExecStagingImageTest ExecStagingImageTestCmd `cmd:""`
}

type GeneratePipelineCmd struct {
	PipelineFile       string `help:"Set the name of the pipeline file" env:"GIPGEE_PIPELINE_FILENAME" default:".gipgee-gitlab-ci.yml"`
	ConfigFileName     string `help:"Set the name of the gipgee config file" env:"GIPGEE_CONFIG_FILE_NAME" default:"gipgee.yml"`
	GipgeeImage        string `help:"Overwrite the gipgee container image" env:"GIPGEE_OVERWRITE_GIPGEE_IMAGE" optional:""`
	ImageSelectionFile string `help:"Specify an image selection file which manually selects the images to be rebuilt. Used by the update check pipeline, there should be no need to use it manually"`
}

type GenerateKanikoAuthCmd struct {
	ConfigFileName string `required:"" env:"GIPGEE_CONFIG_FILE_NAME"`
	ImageId        string `required:""`
}

func (*GeneratePipelineCmd) Help() string {
	return "Generate image build pipeline based on the config gipgee config file"
}

func (*GenerateKanikoAuthCmd) Help() string {
	return "Only for gipgee internal use in the image build pipeline"
}

type ExecStagingImageTestCmd struct {
	ImageId        string `arg:""`
	ConfigFileName string `required:"" env:"GIPGEE_CONFIG_FILE_NAME"`
}

func (cmd *ExecStagingImageTestCmd) Run() error {
	config, err := cfg.LoadConfiguration(cmd.ConfigFileName)
	if err != nil {
		panic(err)
	}
	imageTestCommand := config.Images[cmd.ImageId].TestCommand
	commandString := (*imageTestCommand)[0]
	commandArgsString := make([]string, 0)
	if len(*imageTestCommand) > 1 {
		commandArgsString = append(commandArgsString, (*imageTestCommand)[1:]...)
	}
	commandArgsString = append(commandArgsString, cmd.ImageId)
	executionCmd := exec.Command(commandString, commandArgsString...) // #nosec G204
	executionCmd.Stderr = os.Stderr
	executionCmd.Stdout = os.Stdout
	err = executionCmd.Run()
	return err
}
