package selfrelease

import (
	"fmt"
	"os"

	pm "github.com/devfbe/gipgee/pipelinemodel"
)

func (r *GeneratePipelineCmd) Run() error {
	ai1Stage := pm.Stage{Name: "🔨 all in one"}
	golangImage := pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "golang", Tag: "1.18.3"}
	alpineImage := pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "alpine", Tag: "latest"}
	linterImage := pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "golangci/golangci-lint", Tag: "v1.46.2"}
	securityScannerImage := pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "securego/gosec", Tag: "2.12.0"}
	kanikoImage := pm.ContainerImageCoordinates{Registry: "gcr.io", Repository: "kaniko-project/executor", Tag: "debug"} // FIXME: use fixed version

	testJob := pm.Job{
		Name:  "🧪 Test",
		Image: &golangImage,
		Stage: &ai1Stage,
		Script: []string{
			"go test -race -covermode=atomic -coverprofile=coverage.txt ./...",
		},
		Artifacts: &pm.JobArtifacts{},
	}
	buildJob := pm.Job{
		Name:   "🏗️ Build",
		Image:  &golangImage,
		Stage:  &ai1Stage,
		Script: []string{"CGO_ENABLED=0 go build"},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{"gipgee"},
		},
	}
	lintJob := pm.Job{
		Name:   "📝 Lint",
		Image:  &linterImage,
		Stage:  &ai1Stage,
		Script: []string{"golangci-lint run"},
	}

	securityScanJob := pm.Job{
		Name:   "🛡️ Security Scan",
		Image:  &securityScannerImage,
		Stage:  &ai1Stage,
		Script: []string{"gosec ./..."},
	}
	generateAuthFileJob := pm.Job{
		Name:  "⚙️ Generate Kaniko docker auth file",
		Image: &alpineImage,
		Stage: &ai1Stage,
		Script: []string{
			"ls -la",
			"./gipgee self-release generate-kaniko-docker-auth",
		},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{kanikoSecretsFilename},
		},
		Needs: []pm.JobNeeds{{
			Job:       &buildJob,
			Artifacts: true,
		}},
	}

	registry := os.Getenv("GIPGEE_SELF_RELEASE_REGISTRY")
	repository := os.Getenv("GIPGEE_SELF_RELEASE_REPOSITORY")
	tag := os.Getenv("GIPGEE_SELF_RELEASE_TAG")
	fullDestination := fmt.Sprintf("%v/%v:%v", registry, repository, tag)
	kanikoBuildJob := pm.Job{
		Name:  "🐋 Build staging image using kaniko",
		Image: &kanikoImage,
		Stage: &ai1Stage,
		Script: []string{
			"mkdir -p /kaniko/.docker",
			"cp -v ${CI_PROJECT_DIR}/" + kanikoSecretsFilename + " /kaniko/.docker/config.json",
			"/kaniko/executor --context ${CI_PROJECT_DIR} --dockerfile ${CI_PROJECT_DIR}/Containerfile --destination " + fullDestination,
		},
		Needs: []pm.JobNeeds{
			{Job: &generateAuthFileJob, Artifacts: true},
			{Job: &buildJob, Artifacts: true},
			{Job: &lintJob, Artifacts: false},
			{Job: &securityScanJob, Artifacts: false},
			{Job: &testJob, Artifacts: false},
		},
	}

	pipeline := pm.Pipeline{
		Stages: []pm.Stage{ai1Stage},
		Variables: map[string]interface{}{
			"GOPROXY": "direct",
		},
		Jobs: []pm.Job{
			testJob,
			buildJob,
			lintJob,
			securityScanJob,
			generateAuthFileJob,
			kanikoBuildJob,
		},
	}
	yamlString := pipeline.Render()
	fmt.Print("Generated pipeline is:\n" + yamlString)

	err := os.WriteFile("gipgee-pipeline.yml", []byte(yamlString), 0600)
	if err != nil {
		return err
	}
	return nil
}
