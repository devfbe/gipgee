package updatecheck

import (
	"fmt"
	"log"
	"os"

	"github.com/devfbe/gipgee/config"
	"github.com/devfbe/gipgee/docker"
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

type PipelineParams struct {
	SkipRebuild    bool
	GipgeeImage    string
	Config         *config.Config
	ConfigFileName string
}

func GeneratePipeline(params PipelineParams) *pm.Pipeline {
	ai1Stage := pm.Stage{Name: "🚦 All in One"}
	var pipelineJobs []*pm.Job
	var gipgeeImage *pm.ContainerImageCoordinates
	if params.GipgeeImage != "" {
		var err error // next line doesn't allow := ... is this a golang bug?
		gipgeeImage, err = pm.ContainerImageCoordinatesFromString(params.GipgeeImage)
		if err != nil {
			panic(err)
		}
	} else {
		gipgeeImage = &pm.ContainerImageCoordinates{
			Registry:   "docker.io",
			Repository: "devfbe/gipgee",
			Tag:        "latest",
		}
	}

	// The copyGipgeeAsArtifact job copies the gipgee binary, which is statically linked, to the gitlab artifacts.
	// This helps us e.g. in update check jobs (in the images the user has built) where we can then simply run our
	// gipgee which contains additional code for the update checks / kaniko auth generation / ...
	copyGipgeeAsArtifact := pm.Job{
		Name:  "Copy gipgee to artifacts",
		Stage: &ai1Stage,
		Image: gipgeeImage,
		Script: []string{
			"cp $(which gipgee) gipgee",
		},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{"gipgee"},
		},
	}
	pipelineJobs = append(pipelineJobs, &copyGipgeeAsArtifact)

	imageUpdateCheckResultFiles := map[string][]string{}

	for imageId, imageConfig := range params.Config.Images {

		var locations []*config.ImageLocation

		locations = append(locations, imageConfig.ReleaseLocations...)

		for idx, location := range locations {

			resultFileLocation := fmt.Sprintf("/tmp/gipgee-update-check-result-%s-release-location-%d", imageId, idx)
			imageUpdateCheckResultFiles[imageId] = append(imageUpdateCheckResultFiles[imageId], resultFileLocation)

			if len(*imageConfig.UpdateCheckCommand) > 0 {
				pipelineJobs = append(pipelineJobs, &pm.Job{
					Name:   fmt.Sprintf("Update check %s/%d", imageId, idx),
					Stage:  &ai1Stage,
					Script: []string{fmt.Sprintf("./gipgee update-check exec-update-check %s", imageId)},
					Image: &pm.ContainerImageCoordinates{
						Registry:   *location.Registry,
						Repository: *location.Repository,
						Tag:        *location.Tag,
					},
					Needs: []pm.JobNeeds{
						{
							Job:       &copyGipgeeAsArtifact,
							Artifacts: true,
						},
					},
					Variables: &map[string]interface{}{
						"GIPGEE_CONFIG_FILE_NAME":              params.ConfigFileName,
						"GIPGEE_UPDATE_CHECK_RESULT_FILE_PATH": resultFileLocation,
						"DOCKER_AUTH_CONFIG":                   generateDockerAuthConfig(imageId, params.Config),
					},
				})
			} else {
				log.Printf("Not generating update check job(s) for image '%s' because update check command is empty\n", imageId)
			}
		}
	}

	pipeline := pm.Pipeline{
		Stages:    []*pm.Stage{&ai1Stage},
		Jobs:      pipelineJobs,
		Variables: map[string]interface{}{},
	}
	return &pipeline
}

func generateDockerAuthConfig(imageId string, cfg *config.Config) string {
	env, exists := os.LookupEnv("DOCKER_AUTH_CONFIG")
	dockerAuthConfig := &docker.DockerAuths{Auths: make(map[string]docker.DockerAuth)}
	if exists {
		log.Printf("Extending existing env var DOCKER_AUTH_CONFIG with the necessary pull secrets for image '%s'\n", imageId)
		dockerAuthConfig = docker.LoadAuthConfigFromString(env)
	} else {
		log.Printf("Creating new DOCKER_AUTH_CONFIG env var for the image update check for image '%s'\n", imageId)
	}
	for idx, releaseLocation := range cfg.Images[imageId].ReleaseLocations {
		if releaseLocation.Credentials != nil {
			log.Printf("Generating auth for registry '%s' for image '%s' - target location '%d'\n", *releaseLocation.Registry, imageId, idx)
			configUp, err := cfg.GetUserNamePassword(*releaseLocation.Credentials)
			if err != nil {
				panic(err)
			}
			up := docker.UsernamePassword{
				UserName: configUp.Username,
				Password: configUp.Password,
			}
			dockerAuthConfig.Auths[*releaseLocation.Registry] = up.ToDockerAuth()
		}
	}
	return dockerAuthConfig.ToJsonString()
}
