package updatecheck

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sync"

	cfg "github.com/devfbe/gipgee/config"
)

type AutoUpdateCheckCmd struct {
	ImageId        string `arg:"" required:""`
	ResultFilePath string `arg:"" required:""`
}

func (*AutoUpdateCheckCmd) Help() string {
	return "This command tries to auto detect your currently used package manager and performs an update check if successful"
}

func (cmd *AutoUpdateCheckCmd) Run() error {
	return NewAutoUpdateChecker(cmd.ImageId, cmd.ResultFilePath).Run()
}

type PerformSkopeoUpdateCheckCmd struct {
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_CONFIG_FILE_NAME" default:"gipgee.yml"`
	SkopeoResultPath string `help:"Set the path of the skopeo result" env:"GIPGEE_UPDATE_CHECK_SKOPEO_RESULT_PATH" required:""`
}

type skopeoInspectOutput struct {
	Layers []string `json:"Layers"`
}

func runSkopeoInspect(imageId string, imageLocation *cfg.ImageLocation, config *cfg.Config) *skopeoInspectOutput {
	imageConfig := config.Images[imageId]

	log.Printf("Getting image layers of image '%s', location: '%s'\n", imageId, imageLocation)
	skopeoInspectImageCmdSlice := []string{"skopeo", "inspect", "-n", fmt.Sprintf("docker://%s", imageConfig.BaseImage.String())}
	if imageLocation.Credentials != nil {
		up, err := config.GetUserNamePassword(*imageLocation.Credentials)
		if err != nil {
			panic(err)
		}
		skopeoInspectImageCmdSlice = append(skopeoInspectImageCmdSlice, []string{"--username", up.Username, "--password", up.Password}...)
	}

	skopeoInspectImageCmd := exec.Command(skopeoInspectImageCmdSlice[0], skopeoInspectImageCmdSlice[1:]...) // #nosec G204
	skopeoInspectImageCmd.Stderr = os.Stderr
	output, err := skopeoInspectImageCmd.Output()
	if err != nil {
		panic(err)
	}

	baseImageLayers := skopeoInspectOutput{}
	err = json.Unmarshal(output, &baseImageLayers)
	if err != nil {
		panic(err)
	}
	return &baseImageLayers
}

func CheckIfBaseImageLayersAreDiverged(wg *sync.WaitGroup, imageId string, config *cfg.Config) {
	defer wg.Done()
	imageConfig := config.Images[imageId]

	baseImageLayers := runSkopeoInspect(imageId, imageConfig.BaseImage, config)
	log.Printf("%v\n", baseImageLayers)
	// release location iteration
	for idx, releaseLocation := range imageConfig.ReleaseLocations {
		log.Printf("Getting layers of release location %d (%s)\n", idx, releaseLocation.String())
		releaseLocationLayers := runSkopeoInspect(imageId, releaseLocation, config)
		log.Printf("%v\n", releaseLocationLayers)
	}

}

func (cmd *PerformSkopeoUpdateCheckCmd) Run() error {
	log.Println("Will perform update check here soon")
	config, err := cfg.LoadConfiguration(cmd.ConfigFileName)
	if err != nil {
		panic(err)
	}
	wg := sync.WaitGroup{}
	for imageId := range config.Images {
		wg.Add(1)
		go CheckIfBaseImageLayersAreDiverged(&wg, imageId, config)
	}
	log.Println("Waiting for all skopeo jobs to be finished")
	wg.Wait()
	log.Println("All skopeo inspect jobs finished")
	return nil
}

type UpdateCheckCmd struct {
	GeneratePipeline         GeneratePipelineCmd         `cmd:""`
	ExecUpdateCheck          ExecUpdateCheckCmd          `cmd:""`
	AutoUpdateCheck          AutoUpdateCheckCmd          `cmd:""`
	PerformSkopeoUpdateCheck PerformSkopeoUpdateCheckCmd `cmd:""`
}

type GeneratePipelineCmd struct {
	PipelineFileName string `help:"Set the name of the pipeline file" env:"GIPGEE_PIPELINE_FILENAME" default:".gipgee-gitlab-ci.yml"`
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_CONFIG_FILE_NAME" default:"gipgee.yml"`
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
	commandArgsString = append(commandArgsString, cmd.ImageId, cmd.ResultFilePath)
	executionCmd := exec.Command(commandString, commandArgsString...) // #nosec G204
	executionCmd.Stderr = os.Stderr
	executionCmd.Stdout = os.Stdout
	err = executionCmd.Run()
	return err
}
