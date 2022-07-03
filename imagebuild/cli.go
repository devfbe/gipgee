package imagebuild

type ImageBuildCmd struct {
	PipelineFileName string `help:"Set the name of the pipeline file" env:"GIPGEE_IMAGE_BUILD_PIPELINE_FILENAME" default:".gipgee-gitlab-ci.yml"`
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_IMAGE_BUILD_CONFIG_FILENAME" default:"gipgee.yml"`
	GipgeeImage      string `help:"Overwrite the gipgee container image" env:"GIPGEE_OVERWRITE_GIPGEE_IMAGE" optional:""`
}

type GenerateKanikoDockerAuthCmd struct {
	Target string `enum:"staging,release" required:""`
}

func (r *ImageBuildCmd) Help() string {
	return "Generate image build pipeline based on the config gipgee config file"
}
