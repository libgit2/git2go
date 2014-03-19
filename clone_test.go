package git

import (
	"io/ioutil"
	"testing"
)

func TestClone(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)

	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)

	_, err = Clone(repo.Path(), path, &CloneOptions{Bare: true})

	checkFatal(t, err)
}
