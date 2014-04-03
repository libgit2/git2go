package git

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestClone(t *testing.T) {

	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	seedTestRepo(t, repo)

	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)

	_, err = Clone(repo.Path(), path, &CloneOptions{Bare: true})
	defer os.RemoveAll(path)

	checkFatal(t, err)
}
