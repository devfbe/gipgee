package main

import (
	"log"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/devfbe/gipgee/config"
	"github.com/devfbe/gipgee/imagebuild"
	"github.com/devfbe/gipgee/initialize"
	"github.com/devfbe/gipgee/selfrelease"
	"github.com/devfbe/gipgee/updatecheck"
)

type runCmd struct {
	PipelineFile   string `help:"Set the name of the pipeline file" env:"GIPGEE_PIPELINE_FILENAME" default:".gipgee-gitlab-ci.yml"`
	ConfigFileName string `help:"Set the name of the gipgee config file" env:"GIPGEE_CONFIG_FILENAME" default:"gipgee.yml"`
	GipgeeImage    string `help:"Overwrite the gipgee container image" env:"GIPGEE_OVERWRITE_GIPGEE_IMAGE" optional:""`
}

func (cmd *runCmd) Help() string {
	return "Use this method in your gitlab pipeline and let gipgee what pipeline to create based on the env vars gitlab sets."
}

func decideImagesToBuild(cfg *config.Config) []string {
	images := make([]string, 0)
	for key := range cfg.Images {
		images = append(images, key)
	}
	return images
}

func (cmd *runCmd) Run() error {

	pipelineSourceEnvVar := os.Getenv("CI_PIPELINE_SOURCE")
	gipgeeUpdateCheckVar := os.Getenv("GIPGEE_UPDATE_CHECK")
	log.Println("Checking if CI_PIPELINE_SOURCE is == 'schedule' and GIPGEE_UPDATE_CHECK is == true")
	// see https://docs.gitlab.com/ee/ci/variables/predefined_variables.html

	if pipelineSourceEnvVar == "schedule" && strings.ToLower(gipgeeUpdateCheckVar) == "true" {
		log.Println("Pipeline source is schedule and GIPGEE_UPDATE_CHECK is true, generating update check pipeline")

		cfg, err := config.LoadConfiguration(cmd.ConfigFileName)
		if err != nil {
			panic(err)
		}

		params := updatecheck.PipelineParams{
			SkipRebuild:    false, // only for self release integration test true
			GipgeeImage:    cmd.GipgeeImage,
			ConfigFileName: cmd.ConfigFileName,
			Config:         cfg,
		}
		pipeline := updatecheck.GeneratePipeline(params)
		return pipeline.WritePipelineToFile(cmd.PipelineFile)
	} else {
		log.Println("Detected no update check pipeline schedule, assuming image build pipeline. Checking if feature branch or default branch pipeline")
		defaultBranch := os.Getenv("CI_DEFAULT_BRANCH")
		commitBranch := os.Getenv("CI_DEFAULT_BRANCH")
		if commitBranch == defaultBranch {
			log.Printf("Detected that the commit branch '%s' is the default branch '%s'. Generating image build pipeline that releases the images", commitBranch, defaultBranch)
			cfg, err := config.LoadConfiguration(cmd.ConfigFileName)
			if err != nil {
				panic(err)
			}
			gen := imagebuild.NewBuildPipelineGenerator(cfg, decideImagesToBuild(cfg), true, cmd.PipelineFile, cmd.ConfigFileName, cmd.GipgeeImage)
			pipeline := gen.GeneratePipeline()
			return pipeline.WritePipelineToFile(cmd.PipelineFile)
		} else {
			log.Printf("Detected that the commit branch '%s' is not the default branch '%s'. Generating image build pipeline that builds and tests but does not release the images", commitBranch, defaultBranch)
			cfg, err := config.LoadConfiguration(cmd.ConfigFileName)
			if err != nil {
				panic(err)
			}
			gen := imagebuild.NewBuildPipelineGenerator(cfg, decideImagesToBuild(cfg), false, cmd.PipelineFile, cmd.ConfigFileName, cmd.GipgeeImage)
			pipeline := gen.GeneratePipeline()
			return pipeline.WritePipelineToFile(cmd.PipelineFile)
		}
	}

}

var cli struct {
	Initialize  initialize.InitCmd         `cmd:""`
	SelfRelease selfrelease.SelfReleaseCmd `cmd:""`
	UpdateCheck updatecheck.UpdateCheckCmd `cmd:""`
	ImageBuild  imagebuild.ImageBuildCmd   `cmd:""`
	Run         runCmd                     `cmd:""`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
