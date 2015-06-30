package git

import (
	"io/ioutil"
	"testing"
)

func TestClone(t *testing.T) {

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)

	ref, err := repo.References.Lookup("refs/heads/master")
	checkFatal(t, err)

	repo2, err := Clone(repo.Path(), path, &CloneOptions{Bare: true})
	defer cleanupTestRepo(t, repo2)

	checkFatal(t, err)

	ref2, err := repo2.References.Lookup("refs/heads/master")
	checkFatal(t, err)

	if ref.Cmp(ref2) != 0 {
		t.Fatal("reference in clone does not match original ref")
	}
}
