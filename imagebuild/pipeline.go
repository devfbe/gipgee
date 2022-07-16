package imagebuild

import (
	"fmt"

	c "github.com/devfbe/gipgee/config"
	"github.com/devfbe/gipgee/docker"
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

		kanikoScript := make([]string, 0)
		ignoredPaths := ""
		if config.Quirks.KanikoMoveVarQuirk {
			kanikoScript = append(kanikoScript, "mv /var /var-orig")
			ignoredPaths = "--ignore-path=/var-orig"
		}
		kanikoScript = append(kanikoScript, "./.gipgee/gipgee image-build generate-kaniko-auth --config-file='"+params.ConfigFile+"' --image-id '"+imageToBuild+"'")
		kanikoScript = append(kanikoScript, "/kaniko/executor "+ignoredPaths+" --context ${CI_PROJECT_DIR} --dockerfile ${CI_PROJECT_DIR}/"+*config.Images[imageToBuild].ContainerFile+" --build-arg=GIPGEE_BASE_IMAGE="+config.Images[imageToBuild].BaseImage.String()+" --build-arg=GIPGEE_IMAGE_ID="+imageToBuild+" --destination "+config.Images[imageToBuild].StagingLocation.String())

		buildStagingImageJob := pm.Job{
			Name:   "🐋 Build staging image " + imageToBuild + " using kaniko",
			Image:  &c.KanikoImage,
			Stage:  &allInOneStage,
			Script: kanikoScript,
			Needs: []pm.JobNeeds{{
				Job:       &copyGipgeeToArtifact,
				Artifacts: true,
			}},
		}
		imageConfig := config.Images[imageToBuild]
		stagingImageCoordinates, err := pm.ContainerImageCoordinatesFromString(imageConfig.StagingLocation.String())

		if err != nil {
			panic(err)
		}
		releaseJobNeeds := []pm.JobNeeds{
			{
				Job:       &copyGipgeeToArtifact,
				Artifacts: true,
			},
		}
		if len(*imageConfig.TestCommand) > 0 {
			stagingTestJob := pm.Job{
				Name:   "🧪 Test staging image " + imageToBuild,
				Image:  stagingImageCoordinates,
				Stage:  &allInOneStage,
				Script: []string{fmt.Sprintf("./.gipgee/gipgee image-build exec-staging-image-test %s", imageToBuild)},
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
				Variables: &map[string]interface{}{
					"GIPGEE_CONFIG_FILE_NAME": params.ConfigFile,
				},
			}
			releaseJobNeeds = append(releaseJobNeeds, pm.JobNeeds{Job: &stagingTestJob})
		}

		authMap := make(map[string]docker.UsernamePassword, 0)

		if imageConfig.BaseImage.Credentials != nil {
			up, err := config.GetUserNamePassword(*imageConfig.BaseImage.Credentials)
			if err != nil {
				panic(err)
			}
			authMap[*imageConfig.BaseImage.Registry] = docker.UsernamePassword{
				UserName: up.Username,
				Password: up.Password,
			}
		}
		if imageConfig.StagingLocation.Credentials != nil {
			up, err := config.GetUserNamePassword(*imageConfig.StagingLocation.Credentials)
			if err != nil {
				panic(err)
			}
			authMap[*imageConfig.StagingLocation.Registry] = docker.UsernamePassword{
				UserName: up.Username,
				Password: up.Password,
			}
		}

		releaseScript := []string{}
		skopeoSrcCredentials := ""
		if imageConfig.StagingLocation.Credentials != nil {
			up, err := config.GetUserNamePassword(*imageConfig.StagingLocation.Credentials)
			if err != nil {
				panic(err)
			}
			skopeoSrcCredentials = fmt.Sprintf("--src-username '%s' --src-password '%s'", up.Username, up.Password)
		}

		for _, releaseLocation := range imageConfig.ReleaseLocations {
			skopeoDestCredentials := ""
			if releaseLocation.Credentials != nil {
				up, err := config.GetUserNamePassword(*releaseLocation.Credentials)
				if err != nil {
					panic(err)
				}
				skopeoDestCredentials = fmt.Sprintf("--dest-username '%s' --dest-password '%s'", up.Username, up.Password)
			}
			releaseScript = append(releaseScript, fmt.Sprintf("skopeo copy %s %s docker://%s docker://%s", skopeoSrcCredentials, skopeoDestCredentials, imageConfig.StagingLocation.String(), releaseLocation.String()))
		}
		performReleaseJob := pm.Job{
			Name:   "✨ Release staging image " + imageToBuild,
			Stage:  &allInOneStage,
			Image:  &c.SkopeoImage,
			Script: releaseScript,
			Needs:  releaseJobNeeds,
		}

		stagingBuildJobs = append(stagingBuildJobs, &buildStagingImageJob, &performReleaseJob)
		for _, j := range releaseJobNeeds {
			stagingBuildJobs = append(stagingBuildJobs, j.Job)
		}
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
