package config

import (
	pm "github.com/devfbe/gipgee/pipelinemodel"
)

var GolangImage = pm.ContainerImageCoordinates{Registry: "containerregistry.afriserver.de:5000", Repository: "golang", Tag: "1.18.3"}
var AlpineImage = pm.ContainerImageCoordinates{Registry: "containerregistry.afriserver.de:5000", Repository: "alpine", Tag: "latest"}
var LinterImage = pm.ContainerImageCoordinates{Registry: "containerregistry.afriserver.de:5000", Repository: "golangci/golangci-lint", Tag: "v1.46.2"}
var SecurityScannerImage = pm.ContainerImageCoordinates{Registry: "containerregistry.afriserver.de:5000", Repository: "securego/gosec", Tag: "2.12.0"}
var KanikoImage = pm.ContainerImageCoordinates{Registry: "containerregistry.afriserver.de:5000", Repository: "kaniko-project/executor", Tag: "v1.13.0-debug"}
var SkopeoImage = pm.ContainerImageCoordinates{Registry: "containerregistry.afriserver.de:5000", Repository: "skopeo/stable", Tag: "v1.8.0"}
