package imagebuild

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	c "github.com/devfbe/gipgee/config"
	"github.com/devfbe/gipgee/docker"
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

type ImageBuildPipelineGenerator interface {
	GeneratePipeline() *pm.Pipeline
}

type imageBuildPipelineGeneratorImpl struct {
	config        *c.Config
	imagesToBuild []string
	autoStart     bool
	pipelineFile  string
	configFile    string
	gipgeeImage   string
}

func NewBuildPipelineGenerator(config *c.Config, imagesToBuild []string, autoStart bool, pipelineFile string, configFile string, gipgeeImage string) ImageBuildPipelineGenerator {
	return &imageBuildPipelineGeneratorImpl{
		config:        config,
		imagesToBuild: imagesToBuild,
		autoStart:     autoStart,
		pipelineFile:  pipelineFile,
		configFile:    configFile,
		gipgeeImage:   gipgeeImage,
	}
}

func (pipelineGenerator *imageBuildPipelineGeneratorImpl) GeneratePipeline() *pm.Pipeline {

	allInOneStage := pm.Stage{Name: "🏗️ All in One 🧪"}
	pipelineJobs := make([]*pm.Job, 0)
	// If gitlab triggers a child pipeline with "strategy" depend and this child pipeline
	// only contains jobs that have "when: manual" setting, in some versions gitlab keeps
	// the parent pipeline in the "running" state and in other versions the trigger job
	// in the pipeline fails with "unknown failure". So in this case we just add
	// a "make gitlab happy job" that always runs so that the parent pipeline does not crash.
	if !pipelineGenerator.autoStart {
		job := pm.Job{
			Name:  "Make gitlab happy",
			Stage: &allInOneStage,
			Script: []string{
				`echo "This job is just there to avoid that the parent pipeline fails. This workaround is necessary if all jobs in the generated pipeline are manual triggered jobs which do not automatically start."`,
			},
		}
		pipelineJobs = append(pipelineJobs, &job)
	}

	var gipgeeImageCoordinates pm.ContainerImageCoordinates

	if pipelineGenerator.gipgeeImage == "" {
		gipgeeImageCoordinates = pm.ContainerImageCoordinates{
			Registry:   "docker.io",
			Repository: "devfbe/gipgee",
			Tag:        "latest",
		}
	} else {
		coords, err := pm.ContainerImageCoordinatesFromString(pipelineGenerator.gipgeeImage)
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

	for _, imageToBuild := range pipelineGenerator.imagesToBuild {
		log.Printf("Building image build jobs for image '%s'\n", imageToBuild)
		kanikoScript := make([]string, 0)
		ignoredPaths := ""
		if pipelineGenerator.config.Quirks.KanikoMoveVarQuirk {
			kanikoScript = append(kanikoScript, "mv /var /var-orig")
			ignoredPaths = "--ignore-path=/var-orig"
		}
		containerFile := *pipelineGenerator.config.Images[imageToBuild].ContainerFile
		baseImage := pipelineGenerator.config.Images[imageToBuild].BaseImage.String()
		destination := pipelineGenerator.config.Images[imageToBuild].StagingLocation.String()

		kanikoScript = append(kanikoScript, "./.gipgee/gipgee image-build generate-kaniko-auth --config-file-name='"+pipelineGenerator.configFile+"' --image-id '"+imageToBuild+"'")
		kanikoScript = append(kanikoScript, "/kaniko/executor "+ignoredPaths+" --context ${CI_PROJECT_DIR} --dockerfile ${CI_PROJECT_DIR}/"+containerFile+" --build-arg=GIPGEE_BASE_IMAGE="+baseImage+" --build-arg=GIPGEE_IMAGE_ID="+imageToBuild+" --destination "+destination)

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
		imageConfig := pipelineGenerator.config.Images[imageToBuild]
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
					"GIPGEE_CONFIG_FILE_NAME": pipelineGenerator.configFile,
				},
			}
			releaseJobNeeds = append(releaseJobNeeds, pm.JobNeeds{Job: &stagingTestJob})
		}

		authMap := make(map[string]docker.UsernamePassword, 0)

		if imageConfig.BaseImage.Credentials != nil {
			up, err := pipelineGenerator.config.GetUserNamePassword(*imageConfig.BaseImage.Credentials)
			if err != nil {
				panic(err)
			}
			authMap[*imageConfig.BaseImage.Registry] = docker.UsernamePassword{
				UserName: up.Username,
				Password: up.Password,
			}
		}
		if imageConfig.StagingLocation.Credentials != nil {
			up, err := pipelineGenerator.config.GetUserNamePassword(*imageConfig.StagingLocation.Credentials)
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
			up, err := pipelineGenerator.config.GetUserNamePassword(*imageConfig.StagingLocation.Credentials)
			if err != nil {
				panic(err)
			}
			skopeoSrcCredentials = fmt.Sprintf("--src-username '%s' --src-password '%s'", up.Username, up.Password)
		}

		for _, releaseLocation := range imageConfig.ReleaseLocations {
			skopeoDestCredentials := ""
			if releaseLocation.Credentials != nil {
				up, err := pipelineGenerator.config.GetUserNamePassword(*releaseLocation.Credentials)
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

		pipelineJobs = append(pipelineJobs, &buildStagingImageJob, &performReleaseJob)
		for _, j := range releaseJobNeeds {
			pipelineJobs = append(pipelineJobs, j.Job)
		}
	}

	pipelineJobs = append(pipelineJobs, &copyGipgeeToArtifact)

	pipeline := pm.Pipeline{
		Stages: []*pm.Stage{&allInOneStage},
		Jobs:   pipelineJobs,
		Variables: map[string]interface{}{
			"DOCKER_AUTH_CONFIG": generateDockerAuthConfig(pipelineGenerator.config),
		},
	}

	return &pipeline

}

func generateDockerAuthConfig(config *c.Config) string {
	env, exists := os.LookupEnv("DOCKER_AUTH_CONFIG")
	dockerAuthConfig := &docker.DockerAuths{Auths: make(map[string]docker.DockerAuth)}
	if exists {
		log.Println("Extending existing env var DOCKER_AUTH_CONFIG with the necessary pull secrets for the build pipeline")
		dockerAuthConfig = docker.LoadAuthConfigFromString(env)
	} else {
		log.Println("Creating new DOCKER_AUTH_CONFIG env var for the build pipeline")
	}

	// in the image build pipeline we - currently - only need the staging location as DOCKER_AUTH_CONFIG because
	// only the test jobs download the images via gitlab. The release to staging skopeo job or the kaniko build
	// both craft their credentials manually and do not depend on the DOCKER_AUTH_CONFIG
	for imageId, imageConfig := range config.Images {
		if imageConfig.StagingLocation.Credentials != nil {
			_, exists := dockerAuthConfig.Auths[*imageConfig.StagingLocation.Registry]
			if exists {
				log.Printf("Image id '%s': auth for staging registry '%s' already exists in DOCKER_AUTH_CONFIG, not adding / overwriting (again)\n", imageId, *imageConfig.StagingLocation.Registry)
				// Maybe check if the corresponding auth is the same as already configured and if not to yield a warning?
			} else {
				configUp, err := config.GetUserNamePassword(*imageConfig.StagingLocation.Credentials)
				if err != nil {
					panic(err)
				}
				up := docker.UsernamePassword{
					UserName: configUp.Username,
					Password: configUp.Password,
				}
				dockerAuthConfig.Auths[*imageConfig.StagingLocation.Registry] = up.ToDockerAuth()
				log.Printf("Image id '%s': auth for staging registry '%s' added to DOCKER_AUTH_CONFIG", imageId, *imageConfig.StagingLocation.Registry)
			}
		} else {
			log.Printf("Image id '%s' has no staging location auth configured, nothing to add to DOCKER_AUTH_CONFIG", imageId)
		}
	}

	return dockerAuthConfig.ToJsonString()
}

func (params *GeneratePipelineCmd) Run() error {
	config, err := c.LoadConfiguration(params.ConfigFileName)
	if err != nil {
		return err
	}

	/*
		FIXME: select depending on git diff
	*/

	imagesToBuild := make([]string, 0)
	if params.ImageSelectionFile != "" {
		log.Printf("Image selection file defined ('%s'), loading image selection.\n", params.ImageSelectionFile)
		bytes, err := os.ReadFile(params.ImageSelectionFile)
		if err != nil {
			panic(err)
		}
		imagesToLoadMap := make(map[string]bool, 0)
		err = json.Unmarshal(bytes, &imagesToLoadMap)
		if err != nil {
			panic(err)
		}
		for key := range imagesToLoadMap {
			log.Printf("Added image with id '%s'\n", key)
			imagesToBuild = append(imagesToBuild, key)
		}
	} else {
		for key := range config.Images {
			imagesToBuild = append(imagesToBuild, key)
		}
	}

	var generator = NewBuildPipelineGenerator(
		config, imagesToBuild, true, params.PipelineFile, params.ConfigFileName, params.GipgeeImage,
	)

	pipeline := generator.GeneratePipeline()

	err = pipeline.WritePipelineToFile(params.PipelineFile)
	if err != nil {
		panic(err)
	}
	return nil
}
