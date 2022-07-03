package pipelinemodel

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

var (
	// From https://docs.gitlab.com/ee/ci/yaml/
	globalKeywords = []string{
		"default", "include", "stages", "variables", "workflow",
	}
)

type ContainerImageCoordinates struct {
	Registry   string
	Repository string
	Tag        string
}

func ContainerImageCoordinatesFromString(containerCoordinates string) (*ContainerImageCoordinates, error) {
	firstSlashIdx := strings.Index(containerCoordinates, "/")
	if firstSlashIdx == -1 {
		return nil, fmt.Errorf("didn't find / in container coordinates '%s' - cannot extract registry. Given coordinates must contain at least registry and repository", containerCoordinates)
	}

	registry := containerCoordinates[:firstSlashIdx]

	remaining := containerCoordinates[firstSlashIdx+1:]
	var repository string
	var tag string
	indexOfColon := strings.Index(remaining, ":")
	if indexOfColon == -1 {
		tag = "latest"
		repository = remaining
	} else {
		tag = remaining[indexOfColon+1:]
		repository = remaining[:indexOfColon]
	}

	return &ContainerImageCoordinates{
		Registry:   registry,
		Repository: repository,
		Tag:        tag,
	}, nil

}

func (c *ContainerImageCoordinates) String() string {
	return fmt.Sprintf("%s/%s:%s", c.Registry, c.Repository, c.Tag)
}

func (coordinates *ContainerImageCoordinates) MarshalYAML() (interface{}, error) {
	u, err := url.Parse(coordinates.Registry)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, coordinates.Repository)
	toReturn := u.String() + ":" + coordinates.Tag
	return toReturn, nil
}

func (pipeline *Pipeline) MarshalYAML() (interface{}, error) {
	pipelineMap := make(map[string]interface{})
	pipelineStages := make([]string, len(pipeline.Stages))
	for idx, stage := range pipeline.Stages {
		pipelineStages[idx] = stage.Name
	}
	pipelineMap["stages"] = pipelineStages
	for _, job := range pipeline.Jobs {
		pipelineMap[job.Name] = job
	}
	if len(pipeline.Variables) > 0 {
		pipelineMap["variables"] = pipeline.Variables
	}
	return pipelineMap, nil
}

type Pipeline struct {
	Stages    []*Stage
	Jobs      []*Job
	Variables map[string]interface{}
}

func (p *Pipeline) validate() error {
	for _, job := range p.Jobs {
		for _, globalKeyword := range globalKeywords {
			if job.Name == globalKeyword {
				return errors.New("Found job with name" + job.Name + " which is a reserved keyword. Pipeline validation failed.")
			}
		}
	}
	return nil
}

func (p *Pipeline) Render() string {

	err := p.validate()
	if err != nil {
		panic(err)
	}

	bytes, err := yaml.Marshal(p)
	if err != nil {
		panic(err)
	}

	return string(bytes)

}

type Stage struct {
	Name string
}

func (pipelineStage *Stage) MarshalYAML() (interface{}, error) {
	if pipelineStage.Name == "" {
		return nil, errors.New("Pipeline stage name must not be empty")
	}
	return pipelineStage.Name, nil
}

type JobArtifacts struct {
	Paths   []string `yaml:"paths,omitempty"`
	Exclude []string `yaml:"exclude,omitempty"`
	When    *string  `yaml:"when,omitempty"`
}

type JobAllowFailure struct {
	Allowed   *bool
	ExitCodes *[]int
}

type JobNeeds struct {
	Job       *Job
	Artifacts bool
}

// TODO in gitlab 14.10 it's possible to use a trigger:forward setting to configure if vars are forwarded
// we can really use this feature, so we should add this here later.

type JobTriggerInclude struct {
	Artifact string
	Job      *Job
}

func (jobTriggerInclude JobTriggerInclude) MarshalYAML() (interface{}, error) {
	return map[string]string{
		"artifact": jobTriggerInclude.Artifact,
		"job":      jobTriggerInclude.Job.Name,
	}, nil
}

type JobTrigger struct {
	Include  *JobTriggerInclude `yaml:"include"`
	Strategy string             `yaml:"strategy"`
}

func (jobNeeds JobNeeds) MarshalYAML() (interface{}, error) {
	if jobNeeds.Job == nil {
		return nil, errors.New("needs: needs job to be defined")
	}
	return map[string]interface{}{
		"job":       jobNeeds.Job.Name,
		"artifacts": jobNeeds.Artifacts,
	}, nil

}

type Job struct {
	Name          string                     `yaml:"-"`
	Stage         *Stage                     `yaml:"stage"`
	Script        []string                   `yaml:"script,omitempty"`
	AfterScript   []string                   `yaml:"after_script,omitempty"`
	BeforeScript  []string                   `yaml:"before_script,omitempty"`
	AllowFailure  *JobAllowFailure           `yaml:"allow_failure,omitempty"` // FIXME: use struct representing the simple bool value or exit_codes as child
	Artifacts     *JobArtifacts              `yaml:"artifacts,omitempty"`
	Image         *ContainerImageCoordinates `yaml:"image,omitempty"`
	Needs         []JobNeeds                 `yaml:"needs,omitempty"` // empty array explicitly allowed
	Interruptible *bool                      `yaml:"interruptible,omitempty"`
	Trigger       *JobTrigger                `yaml:"trigger,omitempty"`
	Variables     *map[string]interface{}    `yaml:"variables,omitempty"`
	/*
		cache 	List of files that should be cached between subsequent runs.
		coverage 	Code coverage settings for a given job.
		dast_configuration 	Use configuration from DAST profiles on a job level.
		dependencies 	Restrict which artifacts are passed to a specific job by providing a list of jobs to fetch artifacts from.
		environment 	Name of an environment to which the job deploys.
		except 	Control when jobs are not created.
		extends 	Configuration entries that this job inherits from.
		image 	Use Docker images.
		inherit 	Select which global defaults all jobs inherit.
		interruptible 	Defines if a job can be canceled when made redundant by a newer run.
		only 	Control when jobs are created.
		pages 	Upload the result of a job to use with GitLab Pages.
		parallel 	How many instances of a job should be run in parallel.
		release 	Instructs the runner to generate a release object.
		resource_group 	Limit job concurrency.
		retry 	When and how many times a job can be auto-retried in case of a failure.
		rules 	List of conditions to evaluate and determine selected attributes of a job, and whether or not it’s created.
		script 	Shell script that is executed by a runner.
		secrets 	The CI/CD secrets the job needs.
		services 	Use Docker services images.
		stage 	Defines a job stage.
		tags 	List of tags that are used to select a runner.
		timeout 	Define a custom job-level timeout that takes precedence over the project-wide setting.
		trigger 	Defines a downstream pipeline trigger.
		variables 	Define job variables on a job level.
		when
	*/
}

func (jaf *JobAllowFailure) MarshalYAML() (interface{}, error) {

	if jaf.ExitCodes == nil && jaf.Allowed == nil {
		return nil, errors.New("both exit codes and allowed are nil in JobAllowFailure")
	}

	if jaf.ExitCodes != nil && jaf.Allowed != nil {
		if !*jaf.Allowed {
			return nil, errors.New("allow failure defined exit codes but allowed is false")
		} else {
			return struct {
				ExitCodes *[]int `yaml:"exit_codes"`
			}{
				jaf.ExitCodes,
			}, nil
		}
	}

	return jaf.Allowed, nil
}
