package imagebuild

import (
	"fmt"
	"log"
	"os"

	"github.com/devfbe/gipgee/config"
	"github.com/devfbe/gipgee/docker"
)

const (
	KanikoSecretsFilename = "/kaniko/.docker/config.json" // #nosec G101
)

func (params *GenerateKanikoAuthCmd) Run() error {

	err := os.MkdirAll("/kaniko/.docker", 0700)
	if err != nil {
		panic(err)
	}

	cfg, err := config.LoadConfiguration(params.ConfigFile)

	if err != nil {
		panic(err)
	}

	imgCfg, exists := cfg.Images[params.ImageId]
	if !exists {
		panic(fmt.Errorf("image config '%s' does not exist - this should never happen here", params.ImageId))
	}
	// first of all, ensure that the (potentially read only) base image pull secrets are configured if defined
	if imgCfg.BaseImage.Credentials != nil {
		fmt.Println("FIXME implement base image credentials here")
	}

	switch target := params.Target; target {
	case "release":
		for _, releaseLoc := range imgCfg.ReleaseLocations {
			if releaseLoc.Credentials != nil {
				fmt.Printf("FIXME add credentials for release location %s\n", releaseLoc.String())
			}
		}
	case "staging":
		if imgCfg.StagingLocation.Credentials != nil {
			usernamePassword, err := cfg.GetUserNamePassword(*imgCfg.StagingLocation.Credentials)
			if err != nil {
				panic(err)
			}
			dockerAuth := docker.CreateAuth(map[string]docker.UsernamePassword{*imgCfg.StagingLocation.Registry: {UserName: usernamePassword.Username, Password: usernamePassword.Password}})
			err = os.WriteFile(KanikoSecretsFilename, []byte(dockerAuth), 0600)
			if err != nil {
				panic(err)
			}
			log.Printf("Wrote kaniko docker auth to '%s'", KanikoSecretsFilename)

		} else {
			log.Printf("Warning, no staging location credentials defined for image '%s'\n", imgCfg.Id)
		}
	default:
		panic("this code should never be reached")
	}

	return nil
}
