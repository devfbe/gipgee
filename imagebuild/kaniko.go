package imagebuild

import (
	"fmt"
	"os"

	"github.com/devfbe/gipgee/config"
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
		panic(fmt.Errorf("Image config '%s' does not exist - this should never happen here", params.ImageId))
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
			fmt.Printf("FIXME add credentials for release location %s\n", imgCfg.StagingLocation.String())
		}
	default:
		panic("this code should never be reached")
	}

	return nil
}
