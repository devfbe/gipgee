package selfrelease

import (
	"os"

	"github.com/devfbe/gipgee/docker"
	git "github.com/devfbe/gipgee/git"
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

// Stage 1
func (cmd *GeneratePipelineCmd) Run() error {
	ai1Stage := pm.Stage{Name: "🔨 all in one"}
	golangImage := pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "golang", Tag: "1.18.3"}
	alpineImage := pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "alpine", Tag: "latest"}
	linterImage := pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "golangci/golangci-lint", Tag: "v1.46.2"}
	securityScannerImage := pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "securego/gosec", Tag: "2.12.0"}
	kanikoImage := pm.ContainerImageCoordinates{Registry: "gcr.io", Repository: "kaniko-project/executor", Tag: "debug"} // FIXME: use fixed version
	skopeoImage := pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "alpine", Tag: "latest"}              // TODO own skopeo image
	registry := os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY")
	repository := os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REPOSITORY")
	tag := git.GetCurrentGitRevisionHex()
	stagingImage := pm.ContainerImageCoordinates{Registry: registry, Repository: repository, Tag: tag}

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
			"./gipgee self-release generate-kaniko-docker-auth --target staging",
		},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{kanikoSecretsFilename},
		},
		Needs: []pm.JobNeeds{{
			Job:       &buildJob,
			Artifacts: true,
		}},
	}

	kanikoBuildJob := pm.Job{ // FIXME: add busybox
		Name:  "🐋 Build staging image using kaniko",
		Image: &kanikoImage,
		Stage: &ai1Stage,
		Script: []string{
			"mkdir -p /kaniko/.docker",
			"cp -v ${CI_PROJECT_DIR}/" + kanikoSecretsFilename + " /kaniko/.docker/config.json",
			"/kaniko/executor --context ${CI_PROJECT_DIR} --dockerfile ${CI_PROJECT_DIR}/Containerfile --destination " + stagingImage.String(),
		},
		Needs: []pm.JobNeeds{
			{Job: &generateAuthFileJob, Artifacts: true},
			{Job: &buildJob, Artifacts: true},
			{Job: &lintJob, Artifacts: false},
			{Job: &securityScanJob, Artifacts: false},
			{Job: &testJob, Artifacts: false},
		},
	}

	buildIntegrationTestPipeline := pm.Job{
		Name:  "🪄 Build integration test pipeline (with new staging image)",
		Stage: &ai1Stage,
		Image: &stagingImage,
		Needs: []pm.JobNeeds{
			{Job: &kanikoBuildJob},
		},
		Script: []string{
			"gipgee self-release generate-integration-test-pipeline",
		},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{
				SelfReleaseIntegrationTestPipelineFileName,
			},
		},
	}

	RunIntegrationTestPipeline := pm.Job{
		Name:  "▶️ Run integration test pipeline",
		Stage: &ai1Stage,
		Needs: []pm.JobNeeds{
			{Job: &buildIntegrationTestPipeline, Artifacts: true},
		},
		Trigger: &pm.JobTrigger{
			Include: &pm.JobTriggerInclude{
				Artifact: SelfReleaseIntegrationTestPipelineFileName,
				Job:      &buildIntegrationTestPipeline,
			},
			Strategy: "depend",
		},
	}

	PerformSelfRelease := pm.Job{
		Name:  "🤗 Release staging image",
		Stage: &ai1Stage,
		Image: &skopeoImage,
		Script: []string{
			"apk add skopeo",
			"echo 'i would run skopeo now'",
		},
		Needs: []pm.JobNeeds{
			{Job: &RunIntegrationTestPipeline, Artifacts: false},
		},
	}

	stagingRegistryAuth := docker.CreateAuth(map[string]docker.UsernamePassword{
		os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY"): {
			Password: os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY_PASSWORD"),
			UserName: os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY_USERNAME"),
		},
	})

	pipeline := pm.Pipeline{
		Stages: []*pm.Stage{&ai1Stage},
		Variables: map[string]interface{}{
			"GOPROXY":            "direct",
			"DOCKER_AUTH_CONFIG": stagingRegistryAuth,
		},
		Jobs: []*pm.Job{
			&testJob,
			&buildJob,
			&lintJob,
			&securityScanJob,
			&generateAuthFileJob,
			&kanikoBuildJob,
			&buildIntegrationTestPipeline,
			&RunIntegrationTestPipeline,
			&PerformSelfRelease,
		},
	}

	err := pipeline.WritePipelineToFile("gipgee-pipeline.yml")
	if err != nil {
		panic(err)
	}
	return nil
}
