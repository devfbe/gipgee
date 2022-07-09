package git

import (
	"fmt"
	"os"
	"path/filepath"

	git5 "github.com/go-git/go-git/v5"
)

func GetCurrentGitRevisionHex(workDir string) string {

	var currentDir string
	var err error
	if workDir == "" {
		currentDir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	} else {
		currentDir = workDir
	}

	gitDir := filepath.Join(currentDir, ".git")

	if _, err := os.Stat(gitDir); os.IsNotExist(err) {
		parentDir := filepath.Dir(currentDir)
		if parentDir == currentDir {
			panic(fmt.Errorf("reached dir '%s' while searching for .git but no .git found", parentDir))
		}
		return GetCurrentGitRevisionHex(parentDir)
	}

	repo, err := git5.PlainOpen(gitDir)
	if err != nil {
		panic(err)
	}
	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}
	return ref.Hash().String()
}
