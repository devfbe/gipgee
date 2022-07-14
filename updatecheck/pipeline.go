package updatecheck

import (
	"fmt"

	"github.com/devfbe/gipgee/config"
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
	for imageId, imageConfig := range params.Config.Images {

		var locations []*config.ImageLocation

		locations = append(locations, imageConfig.ReleaseLocations...)

		for idx, location := range locations {
			if imageConfig.UpdateCheckCommand != nil {
				pipelineJobs = append(pipelineJobs, &pm.Job{
					Name:   fmt.Sprintf("Update check job %d - image %s", idx, *location.Repository), // Todo maybe idx of location if multiple??
					Stage:  &ai1Stage,
					Script: []string{fmt.Sprintf("./gipgee exec update-check %s", imageId)},
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
						"GIPGEE_CONFIG_FILE_NAME": params.ConfigFileName,
					},
				})
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
