package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	zglob "github.com/mattn/go-zglob"
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
	currentRevisionHexFromFunction := GetCurrentGitRevisionHex()
	if currentRevisionHexFromGitExecutable != currentRevisionHexFromFunction {
		t.Errorf("Expected commit hex '%s' doesn't match '%s'", currentRevisionHexFromGitExecutable, currentRevisionHexFromFunction)
	}
}

func execWithPanicOnFail(workDir string, executable string, params []string, t *testing.T) string {
	gitPath, err := exec.LookPath(executable)
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(gitPath, params...)
	cmd.Dir = workDir
	t.Logf("Executing command '%s' with params '%s' in '%s'\n", executable, strings.Join(params, " "), workDir)

	output, err := cmd.Output()
	if err != nil {
		t.Fatal(err)
	}

	return string(output)
}

func TestGetChangedFiles(t *testing.T) {
	tempGitDir, err := os.MkdirTemp("", "gipgee-git-unittest")
	//defer os.RemoveAll(tempGitDir)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Created temp dir '%s'\n", tempGitDir)

	wf := func(fileName, content string) {
		filePath := filepath.Join(tempGitDir, fileName)
		parentDir := filepath.Dir(filePath)

		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			t.Logf("Creating directory '%s'\n", parentDir)
			err := os.MkdirAll(parentDir, 0700)
			if err != nil {
				t.Fatal(err)
			}
		}
		t.Logf("Writing file '%s'", filePath)
		err := os.WriteFile(filePath, []byte(content+"\n"), 0600)
		if err != nil {
			t.Fatal(err)
		}

	}

	// it might be a little bit dirty to rely on an existing git installation
	// but this is much easier to test
	execWithPanicOnFail(tempGitDir, "git", []string{"init"}, t)
	execWithPanicOnFail(tempGitDir, "git", []string{"config", "user.name", "unittest"}, t)
	execWithPanicOnFail(tempGitDir, "git", []string{"config", "user.email", "unittest@localhost"}, t)

	wf("README.md", "# This is a test readme")     // this file remains unchanged
	wf("COPYING", "GPL, only GPL!")                // this file will be changed
	wf("test.txt", "this will be deleted content") // this file will be deleted

	wf("a/README.md", "# This is a test readme")     // this file remains unchanged
	wf("a/COPYING", "GPL, only GPL!")                // this file will be changed
	wf("a/test.txt", "this will be deleted content") // this file will be deleted

	wf("a/b/README.md", "# This is a test readme")     // this file remains unchanged
	wf("a/b/COPYING", "GPL, only GPL!")                // this file will be changed
	wf("a/b/test.txt", "this will be deleted content") // this file will be deleted

	execWithPanicOnFail(tempGitDir, "git", []string{"add", "."}, t)
	execWithPanicOnFail(tempGitDir, "git", []string{"commit", "-m", "test"}, t)
	execWithPanicOnFail(tempGitDir, "git", []string{"update-ref", "refs/remotes/origin/test123", "HEAD"}, t)

	execWithPanicOnFail(tempGitDir, "git", []string{"rm", "test.txt", "a/test.txt", "a/b/test.txt"}, t)
	execWithPanicOnFail(tempGitDir, "git", []string{"commit", "-m", "test.txt deleted from all dirs"}, t)

	wf("COPYING", "GPL, only GPL! Join us now and share the software!")
	wf("a/COPYING", "GPL, only GPL! Join us now and share the software!")
	wf("a/b/COPYING", "GPL, only GPL! Join us now and share the software!")

	execWithPanicOnFail(tempGitDir, "git", []string{"commit", "-a", "-m", "changed copying file"}, t)

	currentCommitSha := execWithPanicOnFail(tempGitDir, "git", []string{"rev-parse", "--verify", "HEAD"}, t)
	t.Logf("Current commit sha is '%s'\n", strings.TrimSpace(currentCommitSha))

	changedFiles := getChangedFiles(tempGitDir, currentCommitSha, "test123")
	t.Logf("%v\n", changedFiles)
	expected := []string{"COPYING", "test.txt", "a/COPYING", "a/test.txt", "a/b/COPYING", "a/b/test.txt"}
	less := func(a, b string) bool { return a < b }
	equals := cmp.Equal(changedFiles, expected, cmpopts.SortSlices(less))
	if !equals {
		t.Errorf("slice '%v' does not match expected '%v'\n", changedFiles, expected)
	}
}

func TestFilePathMatch(t *testing.T) {
	fps := []string{
		"foo.txt",
		"bar/x.txt",
		"barfoo/foo/xyz.txt",
	}

	for _, v := range fps {
		m, err := zglob.Match("**/*", v)
		if err != nil {
			t.Error(err)
		}
		t.Logf("%v matches: %v", v, m)
	}

	t.Error("bazinga")
}
