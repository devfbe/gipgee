package selfrelease

type GenerateKanikoDockerAuthCmd struct {
	Target string `enum:"staging,release" required:""`
}

type GeneratePipelineCmd struct {
}

type SelfReleaseCmd struct {
	GenerateKanikoDockerAuth GenerateKanikoDockerAuthCmd `cmd:""`
	GeneratePipeline         GeneratePipelineCmd         `cmd:""`
}

const (
	kanikoSecretsFilename = "gipgee-kaniko-auth.json" // #nosec G101
)
