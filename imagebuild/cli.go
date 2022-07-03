package imagebuild

type ImageBuildCmd struct {
	GenerateKanikoAuth GenerateKanikoAuthCmd `cmd:""`
	GeneratePipeline   GeneratePipelineCmd   `cmd:""`
}

type GeneratePipelineCmd struct {
	PipelineFileName string `help:"Set the name of the pipeline file" env:"GIPGEE_IMAGE_BUILD_PIPELINE_FILENAME" default:".gipgee-gitlab-ci.yml"`
	ConfigFileName   string `help:"Set the name of the gipgee config file" env:"GIPGEE_IMAGE_BUILD_CONFIG_FILENAME" default:"gipgee.yml"`
	GipgeeImage      string `help:"Overwrite the gipgee container image" env:"GIPGEE_OVERWRITE_GIPGEE_IMAGE" optional:""`
}

type GenerateKanikoAuthCmd struct {
	ConfigFileName string `required:""`
	ImageId        string `required:""`
	Target         string `enum:"staging,release" required:""`
}

func (*GeneratePipelineCmd) Help() string {
	return "Generate image build pipeline based on the config gipgee config file"
}

func (*GenerateKanikoAuthCmd) Help() string {
	return "Only for gipgee internal use in the image build pipeline"
}
