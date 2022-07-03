package imagebuild

import (
	c "github.com/devfbe/gipgee/config"
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

func GenerateReleasePipeline(config *c.Config, imagesToBuild []string, autoStart bool, gipgeeImage string) *pm.Pipeline {
	allInOneStage := pm.Stage{Name: "🏗️ All in One 🧪"}
	kanikoImage := pm.ContainerImageCoordinates{Registry: "gcr.io", Repository: "kaniko-project/executor", Tag: "debug"} // FIXME: use fixed version

	var gipgeeImageCoordinates pm.ContainerImageCoordinates

	if gipgeeImage == "" {
		gipgeeImageCoordinates = pm.ContainerImageCoordinates{
			Registry:   "docker.io",
			Repository: "devfbe/gipgee",
			Tag:        "latest",
		}
	} else {
		coords, err := pm.ContainerImageCoordinatesFromString(gipgeeImage)
		if err != nil {
			panic(err)
		}
		gipgeeImageCoordinates = *coords
	}

	copyGipgeeToArtifact := pm.Job{
		Name:  "🧰 provide gipgee binary as artifact",
		Image: &gipgeeImageCoordinates,
		Stage: &allInOneStage,
		Script: []string{
			"mkdir .gipgee && cd .gipgee && cp $(which gipgee) gipgee",
		},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{".gipgee"},
		},
	}

	stagingBuildJobs := make([]*pm.Job, len(imagesToBuild))
	for idx, imageToBuild := range imagesToBuild {
		stagingBuildJobs[idx] = &pm.Job{
			Name:  "🐋 Build staging image " + imageToBuild + " using kaniko",
			Image: &kanikoImage,
			Stage: &allInOneStage,
			Script: []string{
				"mkdir -p /kaniko/.docker",
				"./.gipgee/gipgee help",
				//"cp -v ${CI_PROJECT_DIR}/" + kanikoSecretsFilename + " /kaniko/.docker/config.json",
				"/kaniko/executor --context ${CI_PROJECT_DIR} --dockerfile ${CI_PROJECT_DIR}/" + *config.Images[imageToBuild].ContainerFile + " --no-push",
			},
			Needs: []pm.JobNeeds{{
				Job:       &copyGipgeeToArtifact,
				Artifacts: true,
			}},
		}
	}

	stagingBuildJobs = append(stagingBuildJobs, &copyGipgeeToArtifact)

	pipeline := pm.Pipeline{
		Stages: []*pm.Stage{&allInOneStage},
		Jobs:   stagingBuildJobs,
	}
	return &pipeline

}

func (params *ImageBuildCmd) Run() error {
	config, err := c.LoadConfiguration(params.ConfigFileName)
	if err != nil {
		return err
	}

	/*
		FIXME: select depending on git diff
	*/

	imagesToBuild := make([]string, 0)
	for key := range config.Images {
		imagesToBuild = append(imagesToBuild, key)
	}

	pipeline := GenerateReleasePipeline(config, imagesToBuild, true, params.GipgeeImage) // True only on manual pipeline..
	err = pipeline.WritePipelineToFile(params.PipelineFileName)
	if err != nil {
		panic(err)
	}
	return nil
}
