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
	Exec        ExecCmd                    `cmd:""`
}

func main() {
	ctx := kong.Parse(&cli)
	err := ctx.Run()
	ctx.FatalIfErrorf(err)
}

// Below are the exec commands. The gipgee uses this commands to execute commands like the
// update check command, defined in the image config / defaults. This is safer than trying to
// render commands to the yaml file (you have to handle escaping and quoting for the shell in yaml then)
// because you can just run a process and pass the args as slice to the command instead of serializing
// them to one string in the yaml file.
type ExecCmd struct {
	UpdateCheck      updatecheck.ExecUpdateCheckCmd     `cmd:""`
	StagingImageTest imagebuild.ExecStagingImageTestCmd `cmd:""`
}
