package selfrelease

import (
	"os"

	"github.com/devfbe/gipgee/docker"
)

type GenerateKanikoDockerAuthCmd struct {
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

func (c *GenerateKanikoDockerAuthCmd) Run() error {
	registryPassword := os.Getenv("GIPGEE_SELF_RELEASE_PASSWORD")
	registry := os.Getenv("GIPGEE_SELF_RELEASE_REGISTRY")
	registryUsername := os.Getenv("GIPGEE_SELF_RELEASE_USERNAME")
	authMap := map[string]docker.UsernamePassword{
		"https://" + registry + "/v1/": { // FIXME
			UserName: registryUsername,
			Password: registryPassword,
		},
	}
	err := os.WriteFile(kanikoSecretsFilename, []byte(docker.CreateAuth(authMap)), 0600)
	return err
}
