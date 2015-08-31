package git

import (
	"fmt"
	"testing"
)

func TestListRemotes(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, err := repo.Remotes.Create("test", "git://foo/bar")

	checkFatal(t, err)

	expected := []string{
		"test",
	}

	actual, err := repo.Remotes.List()
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
	defer cleanupTestRepo(t, repo)

	remote, err := repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	options := FetchOptions{
		RemoteCallbacks: RemoteCallbacks{
			CertificateCheckCallback: func(cert *Certificate, valid bool, hostname string) ErrorCode {
				return assertHostname(cert, valid, hostname, t)
			},
		},
	}

	err = remote.Fetch([]string{}, &options, "")
	checkFatal(t, err)
}

func TestRemoteConnect(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	remote, err := repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	err = remote.ConnectFetch(nil)
	checkFatal(t, err)
}

func TestRemoteLs(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	remote, err := repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	err = remote.ConnectFetch(nil)
	checkFatal(t, err)

	heads, err := remote.Ls()
	checkFatal(t, err)

	if len(heads) == 0 {
		t.Error("Expected remote heads")
	}
}

func TestRemoteLsFiltering(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	remote, err := repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	err = remote.ConnectFetch(nil)
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

func TestRemotePruneRefs(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	config, err := repo.Config()
	checkFatal(t, err)
	defer config.Free()

	err = config.SetBool("remote.origin.prune", true)
	checkFatal(t, err)

	_, err = repo.Remotes.Create("origin", "https://github.com/libgit2/TestGitRepository")
	checkFatal(t, err)

	remote, err := repo.Remotes.Lookup("origin")
	checkFatal(t, err)

	if !remote.PruneRefs() {
		t.Fatal("Expected remote to be configured to prune references")
	}
}

func TestRemotePrune(t *testing.T) {
	remoteRepo := createTestRepo(t)
	defer cleanupTestRepo(t, remoteRepo)

	head, _ := seedTestRepo(t, remoteRepo)
	commit, err := remoteRepo.LookupCommit(head)
	checkFatal(t, err)
	defer commit.Free()

	remoteRef, err := remoteRepo.CreateBranch("test-prune", commit, true)
	checkFatal(t, err)

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	config, err := repo.Config()
	checkFatal(t, err)
	defer config.Free()

	remoteUrl := fmt.Sprintf("file://%s", remoteRepo.Workdir())
	remote, err := repo.Remotes.Create("origin", remoteUrl)
	checkFatal(t, err)

	err = remote.Fetch([]string{"test-prune"}, nil, "")
	checkFatal(t, err)

	_, err = repo.References.Create("refs/remotes/origin/test-prune", head, true, "remote reference")
	checkFatal(t, err)

	err = remoteRef.Delete()
	checkFatal(t, err)

	err = config.SetBool("remote.origin.prune", true)
	checkFatal(t, err)

	rr, err := repo.Remotes.Lookup("origin")
	checkFatal(t, err)

	err = rr.ConnectFetch(nil)
	checkFatal(t, err)

	err = rr.Prune(nil)
	checkFatal(t, err)

	_, err = repo.References.Lookup("refs/remotes/origin/test-prune")
	if err == nil {
		t.Fatal("Expected error getting a pruned reference")
	}
}
