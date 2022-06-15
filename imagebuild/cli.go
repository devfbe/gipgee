package imagebuild

type ImageBuildCmd struct {
	PipelineFileName string `help:"Set the name of the pipeline file" env:"GIPGEE_IMAGE_BUILD_PIPELINE_FILENAME" default:".gitlab-ci.yml"`
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_IMAGE_BUILD_CONFIG_FILENAME" default:"gipgee.yml"`
}

func (r *ImageBuildCmd) Help() string {
	return "Generate image build pipeline based on the config gipgee config file"
}
