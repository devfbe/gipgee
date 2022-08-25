package git

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	git5 "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
)

func getGitRepository(workDir string) *git5.Repository {
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
		return getGitRepository(parentDir)
	}

	repo, err := git5.PlainOpen(gitDir)
	if err != nil {
		panic(err)
	}
	return repo
}

func GetCurrentGitRevisionHex() string {
	repo := getGitRepository("")
	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}
	return ref.Hash().String()
}

// This is a special function only designed for calculating
// diff between the current pipeline commit, passed by the CI_COMMIT_SHA env var in your
// gitlab job and a defined other branch (in the gipgee case normally the CI_DEFAULT_BRANCH)
// It assumes, that the originBranch exists as reference remotes/origin/<originBranch>
// To ensure this, you need to set the GIT_DEPTH variable to 0 in your .gitlab-ci.yml to avoid
// a shallow clone which does not contain the other branch
func GetChangedFiles(localCiCommitSHA string, originBranchName string) []string {
	return getChangedFiles("", localCiCommitSHA, originBranchName)
}

func getChangedFiles(workDir string, localCiCommitSHA string, originBranchName string) []string {
	repo := getGitRepository(workDir)

	commitHash := plumbing.NewHash(localCiCommitSHA)

	originRefName := fmt.Sprintf("refs/remotes/origin/%s", originBranchName)
	log.Printf("Looking up remote branch reference '%s'\n", originRefName)
	originBranch, err := repo.Storer.Reference(plumbing.ReferenceName(originRefName))
	if err != nil {
		panic(err)
	}

	localBranchCommit, err := repo.CommitObject(commitHash)
	if err != nil {
		panic(err)
	}
	log.Printf("Local branch commit id is '%s'\n", localBranchCommit.Hash.String())

	originBranchCommit, err := repo.CommitObject(originBranch.Hash())
	if err != nil {
		panic(err)
	}
	log.Printf("Remote branch commit id is '%s'\n", originBranchCommit.Hash.String())

	log.Printf("Calculating diff between '%s' and '%s'\n", originBranchCommit.Hash.String(), localBranchCommit.Hash.String())
	patch, err := localBranchCommit.Patch(originBranchCommit)
	if err != nil {
		panic(err)
	}

	filesChanged := make([]string, len(patch.Stats()))
	for idx, stat := range patch.Stats() {
		log.Printf("Detected file change: '%s'\n", stat.Name)
		filesChanged[idx] = stat.Name
	}

	return filesChanged
}
