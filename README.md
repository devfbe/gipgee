
<p align="center"><img src="https://raw.githubusercontent.com/devfbe/gipgee/main/branding/gipgeegopher-small.png" alt="gipgee logo" /></p>

# gipgee
The Gitlab Image Pipeline Generator (Enhanced Edition)

## Work in progress
* [✅] Basic image build pipeline
* [⚙️ WIP] Parent / Child images in image build pipelines
* [📅 Planned] Feature branch builds
* [📅 Planned] Update Check Pipeline
* [📅 Planned] Gipgee self contained update check for Alpine, RedHat and Debian based distributions

The current version of Gipgee is only able to generate it's own self release pipeline (which is partially used as integration test). Additionally, it can create a basic image build pipeline.

## Configuration
The configuration currently can change. If you want to try gipgee you have to setup the following pipeline / config

### .gitlab-ci.yml
```
stages:
  - build and run test pipeline

run pipeline generator:
  image: docker.io/devfbe/gipgee:latest
  stage: build and run test pipeline
  script:
    - gipgee image-build generate-pipeline
  artifacts:
    paths:
      - .gipgee-gitlab-ci.yml

run generated pipeline:
  stage: build and run test pipeline
  trigger:
    include:
      - artifact: .gipgee-gitlab-ci.yml
        job: run pipeline generator
    strategy: depend
  needs:
    - job: run pipeline generator
      artifacts: true
```

This pipeline creates and runs a image build pipeline (all branches, all events). 

### gipgee.yml
This is a sample configuration of one of my test projects.
```
version: 1
registryCredentials:
  # define credentials named dockerio, fed from the DOCKER_IO_USERNAME/DOCKER_IO_PASSWORD env vars
  dockerio:
    usernameVarName: DOCKER_IO_USERNAME
    passwordVarName: DOCKER_IO_PASSWORD
# You can defaults for all images, but you don't have to.
# If you define more than one image, it is often useful to define defaults 
defaults:
  # The staging registry will be used for the staging images
  # After a successful test, the images will be copied to the release location(s).
  # Currently, gipgee does not support cleaning up the staging location, so you
  # have to remove / clean the staging registry periodically. Not deleting the staging
  # image gives you the opportunity to download and debug the staging image if something goes
  # wrong with in the image test jobs
  defaultStagingRegistry: index.docker.io
  defaultReleaseRegistry: index.docker.io
  # An updateCheckCommand has to be defined for the upcoming update check pipeline feature.
  # Just add gipgee-update-check or sth. like that to this array to make the config parser happy 
  defaultUpdateCheckCommand: ["gipgee-update-check"]
  # This test command will be executed in the staging image test jobs. It's an array so that you
  # can pass multiple arguments safely. gipgee will automatically append the image id as last parameter.
  defaultTestCommand: ["./testContainerImage.sh"]
  # Leave this as is, it's a feature that will be used for the feature branch image builds later.
  defaultAssetsToWatch: []
  # The default container file to use (commonly known as Dockerfile)
  # The build arg GIPGEE_BASE_IMAGE with the value of the image id will be passed
  # for the image build. It is highly recommended to use the GIPGEE_BASE_IMAGE build arg in the
  # FROM directive in your Containerfile (ARG GIPGEE_BASE_IMAGE followed by FROM $GIPGEE_BASE_IMAGE)
  defaultContainerFile: Containerfile
  # The default base image, you can define registry, repository and tag or only parts
  # Of the default base images.
  defaultBaseImage:
    registry: index.docker.io
  # the default credentials to use for the staging / release registry.
  defaultStagingRegistryCredentials: dockerio
  defaultReleaseRegistryCredentials: dockerio
images:

  gipgee-alpine-test:
    baseImage:
      repository: alpine
      tag: latest
    stagingLocation:
      repository: devfbe/gipgee-test
    releaseLocations:
      - repository: devfbe/gipgee-test
        tag: gipgee-alpine-test

  gipgee-redhat-ubi-test:
    baseImage:
      registry: registry.access.redhat.com
      repository: ubi8/ubi
      tag: latest
    stagingLocation:
      repository: devfbe/gipgee-test
    releaseLocations:
      - repository: devfbe/gipgee-test
        tag: gipgee-redhat-ubi-test

  gipgee-debian-test:
    baseImage:
      repository: debian
      tag: latest
    stagingLocation:
      repository: devfbe/gipgee-test
    releaseLocations:
      - repository: devfbe/gipgee-test
        tag: gipgee-debian-test
```

## What will Gipgee be?
The Gipgee will be a tool for dynamically creating container image build and update pipelines for gitlab, based on a yaml configuration.

In most companies I worked for you often have the problem that your container (base) images are build by a pipeline in GitLab, pushed to the registry and then it remains there untouched and without updates. Often there is no time to maintain the images and apply proper quality checks for container images. The Gipgee tries to solve this problem for you, if you are using GitLab.

Gitlab has a feature called [dynamic child pipelines](https://docs.gitlab.com/ee/ci/pipelines/parent_child_pipelines.html#dynamic-child-pipelines). Dynamic child pipelines are pipelines that can be created by a job in your pipeline and then be triggered. Here, the Gipgee comes in. You can commit a gipgeee.yaml configuration file which describes how to build your container image - which base image has to be used, which target locations you want to upload the image to and which tests you want to perform in the image before releasing it. After committing this yaml file and your Containerfile (for most people known as Dockerfile) you can run the Gipgee in your pipeline. The Gipgee will then genereate a child pipeline, based on the supplied yaml file which performs one of the following tasks:
* Rebuild the images
* Update Check

### Rebuild the images
The image rebuild pipeline will be created in different ways, depending on the surrounding context. 
#### Manual trigger
If you manually trigger the pipeline for your default branch, then a build pipeline for all images defined in the gipgee.yaml will be created. The first job of each image build job chain will be set to `when: manual`, which means that you have to click "play" on the corresponding job.
You can force a start of all jobs when triggering a pipeline for the default branch by defining the env var `GIPGEE_FORCE_AUTOSTART=true`. The pipeline will createt and test a staging image, and then release it.
#### Feature branch
In a feature branch, the gipgee will check which files have been changed in the branch and then build all images that have a matching `watchedAssets` configured in the gipgee.yaml. The pipeline will create and test a staging image, but not release it.

### Update Check
The update check pipeline parses your `gipgee.yaml` and then creates update check jobs for the corresponding images. All images defined in the `targetLocations` will be pulled and update checks will be performed. The update check consists of two jobs.
#### The Skopeo layer check
The skopeo layer check checks which layers the base image used for the given image has been configured. It downloads all layer infos (only the layer ids from the manifest, which is really lightweight) and check if the current image is still based on the base (by checking if the base image layer ids are the same as the first layer ids of the current image). 

#### The container image update check
The container image update check is a command you can implement on your own. It will be called and should write the update check result to a update check result file - the name of this file will be supplied.
The best thing the image update check command can do is in most cases:

* List the installed packages (e.g. by calling `dpkg -l | sort`), save the result to a file (e.g. `result-a.txt`)
* Perform package updates (e.g. by calling `apt-get update && apt-get -y upgrade`)
* List the installed packages (e.g. by calling `dpkg -l | sort`), save the result to a file (e.g. `result-b.txt`)
* Compare the list of installed packages. If they differ, yield that as status to the given update check result file.


### Update build pipeline
After the gipgee has processed the results of the update check pipeline, it will trigger an additional image rebuild pipeline which
only rebuilds the images that have updates. The goal here is to be as resource efficient as possible and not to swamp your with unnecessary image registry.

# Gitlab prequisites
## Image pull policy for gitlab runner job containers
One main goal of gipgee is to auto rebuild container images if there are updates. This images are normally
uploaded to the same location (same registry, repository and tag) as the images built in a earlier
build. Because the gipgee then downloads this images to
perform update checks, it's mandatory that the image pull policy for this images is set to `Always`. If
your Gitlab runner configuration uses default values, the `pull_policy` [is set to always by default](https://docs.gitlab.com/runner/configuration/advanced-configuration.html) and you have nothing to change.

# Troubleshooting
## Known kaniko problems
Kaniko is the tool used by gipgee to build container images, because it can run as normal container and is - compared to "docker in docker" - not a security nightmare.

### failed to write "security.capability" attribute
If you run your gitlab runner jobs without additional
capabilities, e.g. with RedHat UBI images there are problems when building with kaniko.

Error messages like
```
error building image: error building stage: failed to get filesystem from image: failed to write "security.capability" attribute to "/usr/bin/newuidmap": operation not permitted
``` 
may appear and break the build. If this happens,
you can add the `CAP_SETFCAP` capability [to the gitlab
runner job containers](https://docs.gitlab.com/runner/configuration/advanced-configuration.html).
