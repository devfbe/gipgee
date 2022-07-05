package updatecheck

import (
	"fmt"
	"github.com/devfbe/gipgee/config"
	cfg "github.com/devfbe/gipgee/config"
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

type PipelineParams struct {
	SkipRebuild      bool
	GipgeeImage      string
	Config           *config.Config
	releaseOrStaging string
}

func (pipeline *PipelineParams) CheckReleaseImages() bool {
	return pipeline.releaseOrStaging == "release"
}

func (pipeline *PipelineParams) CheckStagingImages() bool {
	return pipeline.releaseOrStaging == "staging"
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
	for imageId, imageConfig := range params.Config.Images {

		var locations []*cfg.ImageLocation

		if params.CheckReleaseImages() {
			locations = append(locations, imageConfig.ReleaseLocations...)
		}

		if params.CheckStagingImages() {
			locations = append(locations, imageConfig.StagingLocation)
		}

		for idx, location := range locations {
			var updateCheckCommand []string
			updateCheckCommand = append(updateCheckCommand, *imageConfig.UpdateCheckCommand...)
			updateCheckCommand = append(updateCheckCommand, imageId)
			pipelineJobs = append(pipelineJobs, &pm.Job{
				Name:   fmt.Sprintf("Update check job %d - image %s", idx, *location.Repository), // Todo maybe idx of location if multiple??
				Stage:  &ai1Stage,
				Script: updateCheckCommand,
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
				Variables: nil,
			})
		}
	}

	pipeline := pm.Pipeline{
		Stages:    []*pm.Stage{&ai1Stage},
		Jobs:      pipelineJobs,
		Variables: map[string]interface{}{},
	}
	return &pipeline
}
