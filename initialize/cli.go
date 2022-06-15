package initialize

import (
	"fmt"
	"strconv"
)

type InitCmd struct {
	NoConfig         bool   `help:"Don't generate configuration yaml"`
	NoPipeline       bool   `help:"Don't generate pipeline yaml"`
	PipelineFileName string `help:"Set the name of the pipeline file" env:"GIPGEE_INIT_PIPELINE_FILENAME" default:".gitlab-ci.yml"`
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_INIT_CONFIG_FILENAME" default:"gipgee.yml"`
	Force            bool   `help:"Force generation, overwrite existing files." env:"GIPGEE_INIT_FORCE" default:"false"`
}

func (r *InitCmd) Run() error {
	fmt.Println("initialize " + r.PipelineFileName + "force is " + strconv.FormatBool(r.Force))
	return nil
}

func (r *InitCmd) Help() string {
	return "Initialize a new gipgee configuration and pipeline"
}
