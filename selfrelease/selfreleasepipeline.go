package selfrelease

import (
	"fmt"
	"os"

	"github.com/devfbe/gipgee/config"
	"github.com/devfbe/gipgee/docker"
	git "github.com/devfbe/gipgee/git"
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

const (
	IntegrationTestImageBuildPipelineYamlFileName  = ".gipgee-integrationtest-imagebuild-pipeline.yaml"
	IntegrationTestUpdateCheckPipelineYamlFileName = ".gipgee-integrationtest-updatecheck-pipeline.yaml"
	IntegrationTestConfigFileName                  = "integrationtest/gipgee.yaml"
)

// Stage 1
func (cmd *GeneratePipelineCmd) Run() error {
	ai1Stage := pm.Stage{Name: "🔨 gipgee self release 🌀"}
	registry := os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY")
	repository := os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REPOSITORY")
	tag := git.GetCurrentGitRevisionHex("")
	stagingImage := pm.ContainerImageCoordinates{Registry: registry, Repository: repository, Tag: tag}

	stagingRegistryAuth := docker.CreateAuth(map[string]docker.UsernamePassword{
		os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY"): {
			Password: os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY_PASSWORD"),
			UserName: os.Getenv("GIPGEE_SELF_RELEASE_STAGING_REGISTRY_USERNAME"),
		},
	})

	testJob := pm.Job{
		Name:  "🧪 Test",
		Image: &config.GolangImage,
		Stage: &ai1Stage,
		Script: []string{
			"go test -race -covermode=atomic -coverprofile=coverage.txt ./...",
		},
		Artifacts: &pm.JobArtifacts{},
	}
	buildJob := pm.Job{
		Name:   "🏗️ Build",
		Image:  &config.GolangImage,
		Stage:  &ai1Stage,
		Script: []string{"CGO_ENABLED=0 go build"},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{"gipgee"},
		},
	}
	lintJob := pm.Job{
		Name:   "📝 Lint",
		Image:  &config.LinterImage,
		Stage:  &ai1Stage,
		Script: []string{"golangci-lint run"},
	}

	securityScanJob := pm.Job{
		Name:   "🛡️ Security Scan",
		Image:  &config.SecurityScannerImage,
		Stage:  &ai1Stage,
		Script: []string{"gosec ./..."},
	}
	generateAuthFileJob := pm.Job{
		Name:  "⚙️ Generate Kaniko docker auth file",
		Image: &config.AlpineImage,
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

	kanikoBuildJob := pm.Job{
		Name:  "🐋 Build staging image using kaniko",
		Image: &config.KanikoImage,
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

	buildIntegrationTestImageBuildPipeline := pm.Job{
		Name:  "🪄 Build image build integration test pipeline",
		Stage: &ai1Stage,
		Image: &stagingImage,
		Needs: []pm.JobNeeds{
			{Job: &kanikoBuildJob},
		},
		Script: []string{
			"gipgee image-build generate-pipeline",
		},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{
				IntegrationTestImageBuildPipelineYamlFileName,
			},
		},
		Variables: &map[string]interface{}{
			"GIPGEE_PIPELINE_FILENAME":      IntegrationTestImageBuildPipelineYamlFileName,
			"GIPGEE_CONFIG_FILE_NAME":       IntegrationTestConfigFileName,
			"DOCKER_AUTH_CONFIG":            stagingRegistryAuth,
			"GIPGEE_OVERWRITE_GIPGEE_IMAGE": stagingImage.String(),
		},
	}

	buildIntegrationTestUpdateCheckPipeline := pm.Job{
		Name:  "🚀 Build update check integration test pipeline",
		Stage: &ai1Stage,
		Image: &stagingImage,
		Needs: []pm.JobNeeds{
			{Job: &kanikoBuildJob},
		},
		Script: []string{
			"gipgee update-check generate-pipeline --skip-rebuild",
		},
		Artifacts: &pm.JobArtifacts{
			Paths: []string{
				IntegrationTestUpdateCheckPipelineYamlFileName,
			},
		},
		Variables: &map[string]interface{}{
			"GIPGEE_PIPELINE_FILENAME":      IntegrationTestUpdateCheckPipelineYamlFileName,
			"GIPGEE_CONFIG_FILE_NAME":       IntegrationTestConfigFileName,
			"DOCKER_AUTH_CONFIG":            stagingRegistryAuth,
			"GIPGEE_OVERWRITE_GIPGEE_IMAGE": stagingImage.String(),
		},
	}

	runIntegrationTestImageBuildPipeline := pm.Job{
		Name:  "▶️ Run image build integration test pipeline",
		Stage: &ai1Stage,
		Needs: []pm.JobNeeds{
			{Job: &buildIntegrationTestImageBuildPipeline, Artifacts: true},
		},
		Trigger: &pm.JobTrigger{
			Include: &pm.JobTriggerInclude{
				Artifact: IntegrationTestImageBuildPipelineYamlFileName,
				Job:      &buildIntegrationTestImageBuildPipeline,
			},
			Strategy: "depend",
		},
	}

	runIntegrationTestUpdateCheckPipeline := pm.Job{
		Name:  "▶️ Run image update check integration test pipeline",
		Stage: &ai1Stage,
		Needs: []pm.JobNeeds{
			{Job: &buildIntegrationTestUpdateCheckPipeline, Artifacts: true},
			{Job: &runIntegrationTestImageBuildPipeline, Artifacts: false},
		},
		Trigger: &pm.JobTrigger{
			Include: &pm.JobTriggerInclude{
				Artifact: IntegrationTestUpdateCheckPipelineYamlFileName,
				Job:      &buildIntegrationTestUpdateCheckPipeline,
			},
			Strategy: "depend",
		},
	}

	skopeoCmd := "skopeo copy"
	skopeoCmd += " --dest-creds ${GIPGEE_SELF_RELEASE_RELEASE_REGISTRY_USERNAME}:${GIPGEE_SELF_RELEASE_RELEASE_REGISTRY_PASSWORD}"
	skopeoCmd += " --src-creds ${GIPGEE_SELF_RELEASE_STAGING_REGISTRY_USERNAME}:${GIPGEE_SELF_RELEASE_STAGING_REGISTRY_PASSWORD}"
	skopeoCmd += fmt.Sprintf(" docker://${GIPGEE_SELF_RELEASE_STAGING_REGISTRY}/${GIPGEE_SELF_RELEASE_STAGING_REPOSITORY}:%s", git.GetCurrentGitRevisionHex("."))
	skopeoCmd += " docker://${GIPGEE_SELF_RELEASE_REGISTRY}/${GIPGEE_SELF_RELEASE_REPOSITORY}:${GIPGEE_SELF_RELEASE_TAG}"

	performSelfRelease := pm.Job{
		Name:   "🤗 Release staging image",
		Stage:  &ai1Stage,
		Image:  &config.SkopeoImage,
		Script: []string{skopeoCmd},
		Needs: []pm.JobNeeds{
			{Job: &runIntegrationTestUpdateCheckPipeline, Artifacts: false},
			{Job: &runIntegrationTestImageBuildPipeline, Artifacts: false},
		},
	}

	pipeline := pm.Pipeline{
		Stages: []*pm.Stage{&ai1Stage},
		Variables: map[string]interface{}{
			"GOPROXY": "direct",
		},
		Jobs: []*pm.Job{
			&testJob,
			&buildJob,
			&lintJob,
			&securityScanJob,
			&generateAuthFileJob,
			&kanikoBuildJob,
			&buildIntegrationTestUpdateCheckPipeline,
			&buildIntegrationTestImageBuildPipeline,
			&runIntegrationTestImageBuildPipeline,
			&runIntegrationTestUpdateCheckPipeline,
			&performSelfRelease,
		},
	}

	err := pipeline.WritePipelineToFile("gipgee-pipeline.yml")
	if err != nil {
		panic(err)
	}
	return nil
}
