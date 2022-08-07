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

	skopeoResultLocation := "gipgee-skopeo-result.json"
	skopeoUpdateCheckJob := pm.Job{
		Name:  "🛃 Skopeo update check",
		Stage: &ai1Stage,
		Image: &config.SkopeoImage,
		Script: []string{
			"./gipgee update-check perform-skopeo-update-check",
		},
		Variables: &map[string]interface{}{
			"GIPGEE_CONFIG_FILE_NAME":                params.ConfigFileName,
			"GIPGEE_UPDATE_CHECK_SKOPEO_RESULT_PATH": skopeoResultLocation,
		},
		Needs: []pm.JobNeeds{
			{
				Job:       &copyGipgeeAsArtifact,
				Artifacts: true,
			},
		},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{skopeoResultLocation},
		},
	}

	pipelineJobs = append(pipelineJobs, &skopeoUpdateCheckJob)

	imageUpdateCheckResultFiles := map[string][]string{}

	for imageId, imageConfig := range params.Config.Images {

		var locations []*config.ImageLocation

		locations = append(locations, imageConfig.ReleaseLocations...)

		for idx, location := range locations {
			resultFileLocation := getImageUpdateCheckResultFileName(imageId, idx)
			imageUpdateCheckResultFiles[imageId] = append(imageUpdateCheckResultFiles[imageId], resultFileLocation)

			if len(*imageConfig.UpdateCheckCommand) > 0 {
				pipelineJobs = append(pipelineJobs, &pm.Job{
					Name:   fmt.Sprintf("🛂 Update check %s/%d", imageId, idx),
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
						{
							Job:       &skopeoUpdateCheckJob,
							Artifacts: false,
						},
					},
					Variables: &map[string]interface{}{
						"GIPGEE_CONFIG_FILE_NAME":              params.ConfigFileName,
						"GIPGEE_UPDATE_CHECK_RESULT_FILE_PATH": resultFileLocation,
						"DOCKER_AUTH_CONFIG":                   generateDockerAuthConfig(imageId, params.Config),
					},
					Artifacts: &pm.JobArtifacts{
						Paths: []string{resultFileLocation},
					},
				})
			} else {
				log.Printf("Not generating update check job(s) for image '%s' because update check command is empty\n", imageId)
			}
		}
	}

	rebuildPipelineDependencies := make([]pm.JobNeeds, 0)
	for _, j := range pipelineJobs {
		rebuildPipelineDependencies = append(rebuildPipelineDependencies, pm.JobNeeds{
			Job:       j,
			Artifacts: true,
		})
	}
	generateRebuildPipelineJob := pm.Job{
		Name:  "🛠️ Generate pipeline for rebuilds",
		Stage: &ai1Stage,
		Script: []string{
			"./gipgee update-check generate-image-rebuild-file",
			"./gipgee image-build generate-pipeline --image-selection-file=gipgee-image-rebuild-file.json",
		},
		Needs: rebuildPipelineDependencies,
		Variables: &map[string]interface{}{
			"GIPGEE_CONFIG_FILE_NAME":                params.ConfigFileName,
			"GIPGEE_UPDATE_CHECK_SKOPEO_RESULT_PATH": skopeoResultLocation,
		},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{
				".gipgee-gitlab-ci.yml",
			},
		},
	}

	pipelineJobs = append(pipelineJobs, &generateRebuildPipelineJob)

	if params.SkipRebuild {
		log.Println("Skip rebuild activated, not generating trigger job which starts the rebuild pipeline.")
	} else {
		triggerJob := pm.Job{
			Name:  "🛫 Trigger rebuild pipeline",
			Stage: &ai1Stage,
			Trigger: &pm.JobTrigger{
				Include: &pm.JobTriggerInclude{
					Artifact: ".gipgee-gitlab-ci.yml",
					Job:      &generateRebuildPipelineJob,
				},
				Strategy: "depend",
			},
			Needs: []pm.JobNeeds{
				{
					Job:       &generateRebuildPipelineJob,
					Artifacts: true,
				},
			},
		}

		pipelineJobs = append(pipelineJobs, &triggerJob)
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
