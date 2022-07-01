package selfrelease

import (
	"os"

	"github.com/devfbe/gipgee/docker"
)

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

func (c *GenerateKanikoDockerAuthCmd) Run() error {
	var authMap map[string]docker.UsernamePassword
	var registryPassword string
	var registry string
	var registryUsername string
	if c.Target == "release" {
		registryPassword = os.Getenv("GIPGEE_SELF_RELEASE_RELEASE_REGISTRY_PASSWORD")
		registry = os.Getenv("GIPGEE_SELF_RELEASE_REGISTRY")
		registryUsername = os.Getenv("GIPGEE_SELF_RELEASE_RELEASE_REGISTRY_USERNAME")

	} else if c.Target == "staging" {
		registryPassword = os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY_PASSWORD")
		registry = os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY")
		registryUsername = os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY_USERNAME")
	} else {
		panic("This part of the code should never be reached")
	}
	authMap = map[string]docker.UsernamePassword{
		registry: {
			UserName: registryUsername,
			Password: registryPassword,
		},
	}
	err := os.WriteFile(kanikoSecretsFilename, []byte(docker.CreateAuth(authMap)), 0600)
	return err
}
