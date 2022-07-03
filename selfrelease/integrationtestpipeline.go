package selfrelease

import (
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

const (
	SelfReleaseIntegrationTestPipelineFileName = "gipgee-self-release-integration-test-pipeline.yaml"
)

// Nested integration test pipeline, wait for the result of this pipeline and continue
func (cmd *GenerateIntegrationTestPipelineCmd) Run() error {

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

	err := pipeline.WritePipelineToFile(SelfReleaseIntegrationTestPipelineFileName)
	if err != nil {
		panic(err)
	}

	return nil
}
