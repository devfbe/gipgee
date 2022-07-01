package git

import git5 "github.com/go-git/go-git/v5"

func GetCurrentGitRevisionHex() string {
	repo, err := git5.PlainOpen(".git")
	if err != nil {
		panic(err)
	}
	ref, err := repo.Head()
	if err != nil {
		panic(err)
	}
	return ref.Hash().String()
}
