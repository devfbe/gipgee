package imagebuild

import (
	c "github.com/devfbe/gipgee/config"
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

func GenerateReleasePipeline(config *c.Config, imagesToBuild []string, autoStart bool, params *GeneratePipelineCmd) *pm.Pipeline {
	allInOneStage := pm.Stage{Name: "🏗️ All in One 🧪"}

	var gipgeeImageCoordinates pm.ContainerImageCoordinates

	if params.GipgeeImage == "" {
		gipgeeImageCoordinates = pm.ContainerImageCoordinates{
			Registry:   "docker.io",
			Repository: "devfbe/gipgee",
			Tag:        "latest",
		}
	} else {
		coords, err := pm.ContainerImageCoordinatesFromString(params.GipgeeImage)
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

	stagingBuildJobs := make([]*pm.Job, 0)
	for _, imageToBuild := range imagesToBuild {

		buildStagingImageJob := pm.Job{
			Name:  "🐋 Build staging image " + imageToBuild + " using kaniko",
			Image: &c.KanikoImage,
			Stage: &allInOneStage,
			Script: []string{
				"./.gipgee/gipgee image-build generate-kaniko-auth --config-file='" + params.ConfigFile + "' --target=staging --image-id '" + imageToBuild + "'",
				"/kaniko/executor --context ${CI_PROJECT_DIR} --dockerfile ${CI_PROJECT_DIR}/" + *config.Images[imageToBuild].ContainerFile + " --build-arg=GIPGEE_BASE_IMAGE=" + config.Images[imageToBuild].BaseImage.String() + " --destination " + config.Images[imageToBuild].StagingLocation.String(),
			},
			Needs: []pm.JobNeeds{{
				Job:       &copyGipgeeToArtifact,
				Artifacts: true,
			}},
		}

		stagingImageCoordinates, err := pm.ContainerImageCoordinatesFromString(config.Images[imageToBuild].StagingLocation.String())

		if err != nil {
			panic(err)
		}

		stagingTestJob := pm.Job{
			Name:  "🧪 Test staging image " + imageToBuild,
			Image: stagingImageCoordinates,
			Stage: &allInOneStage,
			Script: []string{
				"echo 'TODO, GIPGEE SHOULD EXECUTE THE COMMAND HERE FIXME'",
			},
			Needs: []pm.JobNeeds{
				{
					Job:       &buildStagingImageJob,
					Artifacts: false,
				},
				{
					Job:       &copyGipgeeToArtifact,
					Artifacts: true,
				},
			},
		}

		performReleaseJob := pm.Job{
			Name:  "✨ Release staging image " + imageToBuild,
			Stage: &allInOneStage,
			Image: &c.SkopeoImage,
			Script: []string{
				"apk add skopeo",
				"echo 'i would run skopeo now'",
			},
			Needs: []pm.JobNeeds{
				{Job: &stagingTestJob, Artifacts: false},
			},
		}

		stagingBuildJobs = append(stagingBuildJobs, &buildStagingImageJob, &stagingTestJob, &performReleaseJob)
	}

	stagingBuildJobs = append(stagingBuildJobs, &copyGipgeeToArtifact)

	pipeline := pm.Pipeline{
		Stages: []*pm.Stage{&allInOneStage},
		Jobs:   stagingBuildJobs,
	}
	return &pipeline

}

func (params *GeneratePipelineCmd) Run() error {
	config, err := c.LoadConfiguration(params.ConfigFile)
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

	pipeline := GenerateReleasePipeline(config, imagesToBuild, true, params) // True only on manual pipeline..
	err = pipeline.WritePipelineToFile(params.PipelineFile)
	if err != nil {
		panic(err)
	}
	return nil
}
