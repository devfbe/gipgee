package selfrelease

import "github.com/devfbe/gipgee/updatecheck"

type GenerateKanikoDockerAuthCmd struct {
	Target string `enum:"staging,release" required:""`
}

type GeneratePipelineCmd struct {
}

type SelfReleaseCmd struct {
	GenerateKanikoDockerAuth GenerateKanikoDockerAuthCmd `cmd:""`
	GeneratePipeline         GeneratePipelineCmd         `cmd:""`
	Exec                     ExecCmd                     `cmd:""`
}

const (
	kanikoSecretsFilename = "gipgee-kaniko-auth.json" // #nosec G101
)

// Below are the exec commands. The gipgee uses this commands to execute commands like the
// update check command, defined in the image config / defaults. This is safer than trying to
// render commands to the yaml file (you have to handle escaping and quoting for the shell in yaml then)
// because you can just run a process and pass the args as slice to the command instead of serializing
// them to one string in the yaml file.
type ExecCmd struct {
	UpdateCheck updatecheck.ExecUpdateCheckCmd `cmd:""`
}
