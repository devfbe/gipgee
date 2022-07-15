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
	authMap := make(map[string]docker.UsernamePassword, 0)
	if imgCfg.BaseImage.Credentials != nil {
		up, err := cfg.GetUserNamePassword(*imgCfg.BaseImage.Credentials)
		if err != nil {
			panic(err)
		}
		authMap[*imgCfg.BaseImage.Registry] = docker.UsernamePassword{
			UserName: up.Username,
			Password: up.Password,
		}
		log.Printf("Added base image registry auth for registry '%s'\n", *imgCfg.BaseImage.Registry)
	} else {
		log.Printf("No base image registry auth configured for registry '%s'\n", *imgCfg.BaseImage.Registry)
	}

	for _, releaseLoc := range imgCfg.ReleaseLocations {
		if releaseLoc.Credentials != nil {
			up, err := cfg.GetUserNamePassword(*releaseLoc.Credentials)
			if err != nil {
				panic(err)
			}
			authMap[*releaseLoc.Registry] = docker.UsernamePassword{
				UserName: up.Username,
				Password: up.Password,
			}
			log.Printf("Added release location registry auth for registry '%s'\n", *releaseLoc.Registry)
		} else {
			log.Printf("No release location registry auth configured for '%s' (release location '%s')\n", *releaseLoc.Registry, releaseLoc.String())
		}

	}
	if imgCfg.StagingLocation.Credentials != nil {
		up, err := cfg.GetUserNamePassword(*imgCfg.StagingLocation.Credentials)
		if err != nil {
			panic(err)
		}
		authMap[*imgCfg.StagingLocation.Registry] = docker.UsernamePassword{
			UserName: up.Username,
			Password: up.Password,
		}
		log.Printf("Added staging location registry auth for registry '%s'\n", *imgCfg.StagingLocation.Registry)
	} else {
		log.Printf("No staging location registry auth configured for '%s'\n", *imgCfg.StagingLocation.Registry)
	}

	dockerAuth := docker.CreateAuth(authMap)
	err = os.WriteFile(KanikoSecretsFilename, []byte(dockerAuth), 0600)
	if err != nil {
		panic(err)
	}
	log.Printf("Wrote kaniko docker auth to '%s'", KanikoSecretsFilename)

	return nil
}
