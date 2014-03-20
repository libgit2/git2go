package git

import (
	"os"
	"testing"
)

func TestRefspecs(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	remote, err := repo.CreateRemoteInMemory("refs/heads/*:refs/heads/*", "git://foo/bar")
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
