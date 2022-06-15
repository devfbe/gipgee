package main

import (
	"github.com/alecthomas/kong"
	"github.com/devfbe/gipgee/imagebuild"
	"github.com/devfbe/gipgee/initialize"
	"github.com/devfbe/gipgee/selfrelease"
	"github.com/devfbe/gipgee/updatecheck"
)

var cli struct {
	Initialize  initialize.InitCmd         `cmd:""`
	SelfRelease selfrelease.SelfReleaseCmd `cmd:""`
	UpdateCheck updatecheck.UpdateCheckCmd `cmd:""`
	ImageBuild  imagebuild.ImageBuildCmd   `cmd:""`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}
