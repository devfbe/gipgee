# gipgee

The Gitlab Image Pipeline Generator (Enhanced Edition)

## Work in progress
The current version of Gipgee is only able to release itself, the project is currently under development.

## What will Gipgee be?
The Gipgee will be a tool for dynamically creating container image build and update pipelines for gitlab, based on a yaml configuration.

In most companies I worked for you often have the problem that your container (base) images are build by a pipeline in GitLab, pushed to the registry and then it remains there untouched and without updates. Often there is no time to maintain the images and apply proper quality checks for container images. The Gipgee tries to solve this problem for you, if you are using GitLab.

Gitlab has a feature called [dynamic child pipelines](https://docs.gitlab.com/ee/ci/pipelines/parent_child_pipelines.html#dynamic-child-pipelines). Dynamic child pipelines are pipelines that can be created by a job in your pipeline and then be triggered. Here, the Gipgee comes in. You can commit a gipgeee.yaml configuration file which describes how to build your container image - which base image has to be used, which target locations you want to upload the image to and which tests you want to perform in the image before rleassing it. After committing this yaml file and your Containerfile (for most people known as Dockerfile) you can run the Gipgee in your pipeline. The Gipgee will then genereate a child pipeline, based on the supplied yaml file which performs one of the following tasks:
* Rebuild the images
* Update Check
