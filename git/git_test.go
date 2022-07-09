package git

import (
	"os/exec"
	"strings"
	"testing"
)

func TestGetCurrentGitRevisionHex(t *testing.T) {
	pathToNativeGit, err := exec.LookPath("git")
	if err != nil {
		t.Fatal("Cannot find git command, cannot run this test then")
	}
	cmd := exec.Command(pathToNativeGit, "rev-parse", "HEAD")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Error(err)
	}
	currentRevisionHexFromGitExecutable := strings.Trim(string(output), " \t\n")
	currentRevisionHexFromFunction := GetCurrentGitRevisionHex("")
	if currentRevisionHexFromGitExecutable != currentRevisionHexFromFunction {
		t.Errorf("Expected commit hex '%s' doesn't match '%s'", currentRevisionHexFromGitExecutable, currentRevisionHexFromFunction)
	}
}
