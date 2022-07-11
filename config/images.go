package config

import (
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

var GolangImage = pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "golang", Tag: "1.18.3"}
var AlpineImage = pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "alpine", Tag: "latest"}
var LinterImage = pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "golangci/golangci-lint", Tag: "v1.46.2"}
var SecurityScannerImage = pm.ContainerImageCoordinates{Registry: "docker.io", Repository: "securego/gosec", Tag: "2.12.0"}
var KanikoImage = pm.ContainerImageCoordinates{Registry: "gcr.io", Repository: "kaniko-project/executor", Tag: "v1.8.1-debug"}
var SkopeoImage = pm.ContainerImageCoordinates{Registry: "quay.io", Repository: "skopeo/stable", Tag: "v1.8.0"}
