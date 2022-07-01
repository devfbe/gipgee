package config

import (
	"errors"
	"os"
	"path/filepath"
	"strconv"

	yaml "gopkg.in/yaml.v3"
)

type Credentials struct {
	UsernameVarName *string `yaml:"usernameVarName"`
	PasswordVarName *string `yaml:"passwordVarName"`
	AuthEnvVar      *string `yaml:"authEnvVar"`
	AuthFile        *string `yaml:"authFile"`
}

type Config struct {
	Version             int                     `yaml:"version"`
	Defaults            Defaults                `yaml:"defaults"`
	RegistryCredentials map[string]*Credentials `yaml:"registryCredentials"`
	Images              map[string]*Image       `yaml:"images"`
}

type BuildArg struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"` // TODO think about template engine here?
}

type Defaults struct {
	DefaultStagingRegistry              *string        `yaml:"defaultStagingRegistry,omitempty"`
	DefaultReleaseRegistry              *string        `yaml:"defaultReleaseRegistry,omitempty"`
	DefaultBaseImageRegistry            *string        `yaml:"defaultBaseImageRegistry,omitempty"`
	DefaultContainerFile                *string        `yaml:"defaultContainerFile,omitempty"`
	DefaultStagingRegistryCredentials   *string        `yaml:"defaultStagingRegistryCredentials,omitempty"`
	DefaultBaseImageRegistryCredentials *string        `yaml:"defaultBaseImageRegistryCredentials,omitempty"`
	DefaultReleaseRegistryCredentials   *string        `yaml:"defaultReleaseRegistryCredentials"`
	DefaultUpdateCheckCommand           *[]string      `yaml:"defaultUpdateCheckCommand,omitempty"`
	DefaultTestCommand                  *[]string      `yaml:"defaultTestCommand,omitempty"`
	DefaultAssetsToWatch                *[]string      `yaml:"defaultAssetsToWatch,omitempty"`
	DefaultBaseImage                    *ImageLocation `yaml:"defaultBaseImage,omitempty"`
	DefaultBuildArgs                    *[]BuildArg    `yaml:"defaultBuildArgs,omitempty"`
}

type ImageLocation struct {
	Registry    *string `yaml:"registry"`
	Repository  *string `yaml:"repository"`
	Tag         *string `yaml:"tag"`
	Credentials *string `yaml:"credentials"`
}

type Image struct {
	ContainerFile      *string          `yaml:"containerFile,omitempty"`
	StagingLocation    *ImageLocation   `yaml:"stagingLocation,omitempty"`
	ReleaseLocations   []*ImageLocation `yaml:"releaseLocations"`
	BaseImage          *ImageLocation   `yaml:"baseImage"`
	UpdateCheckCommand *[]string        `yaml:"updateCheckCommand,omitempty"`
	TestCommand        *[]string        `yaml:"testCommand,omitempty"`
	AssetsToWatch      *[]string        `yaml:"assetsToWatch,omitempty"`
	BuildArgs          *[]BuildArg      `yaml:"buildArgs,omitempty"`
}

func LoadConfiguration(relativePath string) (*Config, error) {
	bytes, err := os.ReadFile(filepath.Clean(relativePath))
	if err != nil {
		return nil, err
	}
	config := Config{}
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	err = fillConfigWithDefaultsAndValidate(&config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}

func fillConfigWithDefaultsAndValidate(config *Config) error {
	for imageName, image := range config.Images {

		if image.ContainerFile == nil {
			if config.Defaults.DefaultContainerFile != nil {
				image.ContainerFile = config.Defaults.DefaultContainerFile
			} else {
				return errors.New("containerFile not defined in image " + imageName + " and no default defined")
			}
		}

		if image.StagingLocation == nil {
			if config.Defaults.DefaultStagingRegistry != nil {
				image.StagingLocation = &ImageLocation{
					Registry: config.Defaults.DefaultStagingRegistry,
					// TODO fill with values from environ
				}
			} else {
				return errors.New("staging registry not defined for image " + imageName + " and no default defined")
			}
		}

		if image.StagingLocation.Registry == nil {
			if config.Defaults.DefaultStagingRegistry != nil {
				image.StagingLocation.Registry = config.Defaults.DefaultStagingRegistry
			} else {
				return errors.New("staging registry not defined for image " + imageName + " and no default defined")
			}
		}

		if image.StagingLocation.Credentials == nil && config.Defaults.DefaultStagingRegistry != nil {
			image.StagingLocation.Credentials = config.Defaults.DefaultStagingRegistry
		}

		if len(image.ReleaseLocations) == 0 {
			return errors.New("no release locations defined for image " + imageName)
		}

		for idx, releaseLocation := range image.ReleaseLocations {
			if releaseLocation.Registry == nil && config.Defaults.DefaultReleaseRegistry != nil {
				releaseLocation.Registry = config.Defaults.DefaultReleaseRegistry
			} else if releaseLocation.Registry == nil {
				return errors.New("registry not defined in release location " + strconv.Itoa(idx) + " for image " + imageName)
			}

			if releaseLocation.Credentials == nil && config.Defaults.DefaultReleaseRegistryCredentials != nil {
				releaseLocation.Credentials = config.Defaults.DefaultReleaseRegistryCredentials
			}
		}

		if image.BaseImage == nil {
			image.BaseImage = &ImageLocation{}
		}

		if image.BaseImage.Registry == nil {
			if config.Defaults.DefaultBaseImage != nil && config.Defaults.DefaultBaseImage.Registry != nil {
				image.BaseImage.Registry = config.Defaults.DefaultBaseImage.Registry
			}
		}
		if image.BaseImage.Repository == nil {
			if config.Defaults.DefaultBaseImage != nil && config.Defaults.DefaultBaseImage.Repository != nil {
				image.BaseImage.Repository = config.Defaults.DefaultBaseImage.Repository
			}
		}
		if image.BaseImage.Tag == nil {
			if config.Defaults.DefaultBaseImage != nil && config.Defaults.DefaultBaseImage.Tag != nil {
				image.BaseImage.Tag = config.Defaults.DefaultBaseImage.Tag
			}
		}

		if image.BaseImage.Credentials == nil && config.Defaults.DefaultBaseImageRegistryCredentials != nil {
			image.BaseImage.Credentials = config.Defaults.DefaultBaseImageRegistryCredentials
		}

		if image.UpdateCheckCommand == nil {
			if config.Defaults.DefaultUpdateCheckCommand != nil {
				image.UpdateCheckCommand = config.Defaults.DefaultUpdateCheckCommand
			} else {
				return errors.New("image update check command not defined and no default given")
			}
		}

		if image.TestCommand == nil {
			if config.Defaults.DefaultTestCommand != nil {
				image.TestCommand = config.Defaults.DefaultTestCommand
			} else {
				return errors.New("image test command not defined and no default given")
			}
		}

		if image.AssetsToWatch == nil {
			if config.Defaults.DefaultAssetsToWatch != nil {
				image.AssetsToWatch = config.Defaults.DefaultAssetsToWatch
			} else {
				return errors.New("default assets to watch not defined and no default given")
			}
		}

		if image.BuildArgs == nil {
			if config.Defaults.DefaultBuildArgs != nil {
				image.BuildArgs = config.Defaults.DefaultBuildArgs
			}
		}
	}
	return nil
}
