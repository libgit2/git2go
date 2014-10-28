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

func assertHostname(cert *Certificate, valid bool, hostname string, t *testing.T) ErrorCode {
	if hostname != "github.com" {
		t.Fatal("Hostname does not match")
		return ErrUser
	}

	return 0
}

func TestCertificateCheck(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
	defer repo.Free()

	remote, err := repo.CreateRemote("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	callbacks := RemoteCallbacks{
		CertificateCheckCallback: func(cert *Certificate, valid bool, hostname string) ErrorCode {
			return assertHostname(cert, valid, hostname, t)
		},
	}

	err = remote.SetCallbacks(&callbacks)
	checkFatal(t, err)
	err = remote.Fetch([]string{}, nil, "")
	checkFatal(t, err)
}

func TestRemoteConnect(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
	defer repo.Free()

	remote, err := repo.CreateRemote("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	err = remote.ConnectFetch()
	checkFatal(t, err)
}

func TestRemoteLs(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
	defer repo.Free()

	remote, err := repo.CreateRemote("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	err = remote.ConnectFetch()
	checkFatal(t, err)

	heads, err := remote.Ls()
	checkFatal(t, err)

	if len(heads) == 0 {
		t.Error("Expected remote heads")
	}
}

func TestRemoteLsFiltering(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
	defer repo.Free()

	remote, err := repo.CreateRemote("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	err = remote.ConnectFetch()
	checkFatal(t, err)

	heads, err := remote.Ls("master")
	checkFatal(t, err)

	if len(heads) != 1 {
		t.Fatalf("Expected one head for master but I got %d", len(heads))
	}

	if heads[0].Id == nil {
		t.Fatalf("Expected head to have an Id, but it's nil")
	}

	if heads[0].Name == "" {
		t.Fatalf("Expected head to have a name, but it's empty")
	}
}
