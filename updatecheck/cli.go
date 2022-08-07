package updatecheck

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

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

func CheckIfBaseImageLayersAreDiverged(imageId string, config *cfg.Config, rebuildNeededChan chan bool) {
	imageConfig := config.Images[imageId]

	baseImageLayers := runSkopeoInspect(imageId, imageConfig.BaseImage, config)
	// release location iteration
	for idx, releaseLocation := range imageConfig.ReleaseLocations {
		log.Printf("Getting layers of release location %d (%s)\n", idx, releaseLocation.String())
		releaseLocationLayers := runSkopeoInspect(imageId, releaseLocation, config)
		log.Printf("Comparing base image layers of base image '%s' and child image layers of '%s'\n", imageConfig.BaseImage.String(), releaseLocation.String())
		layersMatch := baseImageLayersMatch(baseImageLayers.Layers, releaseLocationLayers.Layers)
		if !layersMatch {
			log.Printf("Base image layer %v are not the start layers of the image at release location '%s' (%v). Returning that rebuild for image '%s' is necessary.\n", baseImageLayers.Layers, releaseLocation.String(), releaseLocationLayers, imageId)
			rebuildNeededChan <- true
			return
		} else {
			log.Printf("Image layers of '%s' contain the same start layers as base image '%s', no rebuild necessary for this release location", releaseLocation.String(), imageConfig.BaseImage.String())
		}
	}
	rebuildNeededChan <- false
}

func baseImageLayersMatch(baseImageLayers, childImageLayers []string) bool {
	for len(baseImageLayers) > len(childImageLayers) {
		return false
	}

	for idx, baseImageLayer := range baseImageLayers {
		if childImageLayers[idx] != baseImageLayer {
			return false
		}
	}
	return true
}

func (cmd *PerformSkopeoUpdateCheckCmd) Run() error {
	log.Println("Will perform update check here soon")
	config, err := cfg.LoadConfiguration(cmd.ConfigFileName)
	resultChanMap := make(map[string]chan bool, len(config.Images))
	if err != nil {
		panic(err)
	}
	for imageId := range config.Images {
		rebuildNeededChan := make(chan bool)
		resultChanMap[imageId] = rebuildNeededChan
		go CheckIfBaseImageLayersAreDiverged(imageId, config, rebuildNeededChan)
	}
	log.Println("Waiting for all skopeo jobs to be finished")
	resultMap := make(map[string]bool, len(config.Images))
	for imageId, resultChan := range resultChanMap {
		resultMap[imageId] = <-resultChan
	}
	log.Println("All skopeo inspect jobs finished")

	resultMapAsJson, err := json.Marshal(resultMap)
	if err != nil {
		panic(err)
	}

	log.Printf("Writing result map (%v) to %s\n", string(resultMapAsJson), cmd.SkopeoResultPath)
	err = os.WriteFile(cmd.SkopeoResultPath, resultMapAsJson, 0600)
	if err != nil {
		panic(err)
	}
	return nil
}

type UpdateCheckCmd struct {
	GeneratePipeline         GeneratePipelineCmd         `cmd:""`
	ExecUpdateCheck          ExecUpdateCheckCmd          `cmd:""`
	AutoUpdateCheck          AutoUpdateCheckCmd          `cmd:""`
	PerformSkopeoUpdateCheck PerformSkopeoUpdateCheckCmd `cmd:""`
	GenerateImageRebuildFile GenerateImageRebuildFileCmd `cmd:""`
}

type GenerateImageRebuildFileCmd struct {
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_CONFIG_FILE_NAME" default:"gipgee.yml"`
	SkopeoResultPath string `help:"Set the path of the skopeo result" env:"GIPGEE_UPDATE_CHECK_SKOPEO_RESULT_PATH" required:""`
}

func (cmd *GenerateImageRebuildFileCmd) Run() error {

	imagesToRebuild := make(map[string]bool, 0)

	log.Printf("Loading gipgee configuration file '%s'\n", cmd.ConfigFileName)
	config, err := cfg.LoadConfiguration(cmd.ConfigFileName)
	if err != nil {
		return err
	}
	log.Println("Configuration successfully loaded")
	log.Printf("Checking skopeo result, loading result file '%s'\n", cmd.SkopeoResultPath)
	resultMap := make(map[string]bool, 0)
	skopeoResults, err := os.ReadFile(cmd.SkopeoResultPath)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(skopeoResults, &resultMap)
	if err != nil {
		panic(err)
	}

	for key, value := range resultMap {
		if value {
			log.Printf("Skopeo check indicates rebuild needed for image '%s', adding to rebuild list.\n", key)
			imagesToRebuild[key] = true
		} else {
			log.Printf("Skopeo check didn't indicate rebuild for image '%s'\n", key)
		}
	}

	log.Println("Checking image update check results")
	for _, imageConfig := range config.Images {
		if len(*imageConfig.UpdateCheckCommand) > 0 {
			for idx, location := range imageConfig.ReleaseLocations {
				resultFileLocation := getImageUpdateCheckResultFileName(imageConfig.Id, idx)
				log.Printf("Trying to load resultfile '%s' for image '%s', target location '%d' (%s)\n", resultFileLocation, imageConfig.Id, idx, location.String())
				resultFile, err := os.ReadFile(resultFileLocation) // #nosec G304
				if err != nil {
					panic(err)
				}
				result := strings.TrimSuffix(string(resultFile), "\n")
				if result == "UPGRADE_NEEDED" {
					log.Printf("Result file '%s' contains UPGRADE_NEEDED, adding '%s' to image rebuild list (if not already exists)\n", resultFileLocation, imageConfig.Id)
					imagesToRebuild[imageConfig.Id] = true
				} else if result == "NO_UPGRADE_NEEDED" {
					log.Printf("Result file '%s' contains NO_UPGRADE_NEEDED, not adding '%s' to image rebuild list\n", resultFileLocation, imageConfig.Id)
				} else {
					panic(fmt.Errorf("'%s' is not a valid content for a image update check result file", result))
				}
			}
		} else {
			log.Printf("Skipping check for image '%s' because update check command is not defined\n", imageConfig.Id)
		}
	}

	imagesToRebuildJson, err := json.Marshal(imagesToRebuild)
	if err != nil {
		panic(err)
	}
	log.Printf("Images to rebuild (image selection file content: '%s')\n", imagesToRebuildJson)
	rebuildFileName := "gipgee-image-rebuild-file.json" // TODO make as param
	log.Printf("Writing json to file '%s'\n", rebuildFileName)
	err = os.WriteFile(rebuildFileName, imagesToRebuildJson, 0600)
	if err != nil {
		panic(err)
	}
	return nil
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
