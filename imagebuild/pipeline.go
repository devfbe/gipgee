package imagebuild

import (
	c "github.com/devfbe/gipgee/config"
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

// Nested integration test pipeline, wait for the result of this pipeline and continue
func PrevIntegPipeline() error {

	stubStage := pm.Stage{Name: "Stub integration test stage"}
	stubJob := pm.Job{
		Name:  "do nothing, just a stub",
		Stage: &stubStage,
		Script: []string{
			"echo 'doing nothing'",
		},
	}
	pipeline := pm.Pipeline{
		Stages: []*pm.Stage{&stubStage},
		Jobs:   []*pm.Job{&stubJob},
	}

	err := pipeline.WritePipelineToFile("xxxxx")
	if err != nil {
		panic(err)
	}

	return nil
}

func GenerateReleasePipeline(config *c.Config, imagesToBuild []string, autoStart bool) *pm.Pipeline {
	allInOneStage := pm.Stage{Name: "🏗️ All in One 🧪"}
	kanikoImage := pm.ContainerImageCoordinates{Registry: "gcr.io", Repository: "kaniko-project/executor", Tag: "debug"} // FIXME: use fixed version

	stagingBuildJobs := make([]*pm.Job, len(imagesToBuild))
	for idx, imageToBuild := range imagesToBuild {
		stagingBuildJobs[idx] = &pm.Job{
			Name:  "🐋 Build staging image " + imageToBuild + " using kaniko",
			Image: &kanikoImage,
			Stage: &allInOneStage,
			Script: []string{
				"mkdir -p /kaniko/.docker",
				//"cp -v ${CI_PROJECT_DIR}/" + kanikoSecretsFilename + " /kaniko/.docker/config.json",
				"/kaniko/executor --context ${CI_PROJECT_DIR} --dockerfile ${CI_PROJECT_DIR}/Containerfile --no-push",
			},
		}
	}

	pipeline := pm.Pipeline{
		Stages: []*pm.Stage{&allInOneStage},
		Jobs:   stagingBuildJobs,
	}
	return &pipeline

}

func (r *ImageBuildCmd) Run() error {
	config, err := c.LoadConfiguration(r.ConfigFileName)
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

	pipeline := GenerateReleasePipeline(config, imagesToBuild, true) // True only on manual pipeline..
	err = pipeline.WritePipelineToFile(r.PipelineFileName)
	if err != nil {
		panic(err)
	}
	return nil
}
