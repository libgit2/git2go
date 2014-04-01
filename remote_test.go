package git

import (
	"os"
	"testing"
)

func TestRefspecs(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
	defer repo.Free()

	remote, err := repo.CreateAnonymousRemote("git://foo/bar", "refs/heads/*:refs/heads/*")
	checkFatal(t, err)

	expected := []string{
		"refs/heads/*:refs/remotes/origin/*",
		"refs/pull/*/head:refs/remotes/origin/*",
	}

	err = remote.SetFetchRefspecs(expected)
	checkFatal(t, err)

	actual, err := remote.FetchRefspecs()
	checkFatal(t, err)

	compareStringList(t, expected, actual)
}

func TestListRemotes(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
	defer repo.Free()

	_, err := repo.CreateRemote("test", "git://foo/bar")

	checkFatal(t, err)

	expected := []string{
		"test",
	}

	actual, err := repo.ListRemotes()
	checkFatal(t, err)

	compareStringList(t, expected, actual)
}
