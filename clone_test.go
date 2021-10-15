package git

import (
	"fmt"
	"io/ioutil"
	"net"
	"testing"
	"time"
)

const (
	REMOTENAME = "testremote"
)

func TestClone(t *testing.T) {
	t.Parallel()
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

func TestCloneWithCallback(t *testing.T) {
	t.Parallel()
	testPayload := 0

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)

	opts := CloneOptions{
		Bare: true,
		RemoteCreateCallback: func(r *Repository, name, url string) (*Remote, error) {
			testPayload += 1
			return r.Remotes.Create(REMOTENAME, url)
		},
	}

	repo2, err := Clone(repo.Path(), path, &opts)
	defer cleanupTestRepo(t, repo2)

	checkFatal(t, err)

	if testPayload != 1 {
		t.Fatal("Payload's value has not been changed")
	}

	remote, err := repo2.Remotes.Lookup(REMOTENAME)
	if err != nil || remote == nil {
		t.Fatal("Remote was not created properly")
	}
	defer remote.Free()
}

// TestCloneWithExternalHTTPSUrl warning: this test might be flaky
func TestCloneWithExternalHTTPSUrl(t *testing.T) {
	// fail the test, if the URL is not reachable
	timeout := 1 * time.Second
	_, err := net.DialTimeout("tcp", "github.com:443", timeout)
	if err != nil {
		t.Fatal("Site unreachable, retry later, error: ", err)
	}

	url := "https://github.com/libgit2/git2go.git"
	fmt.Println(url)
	path, err := ioutil.TempDir("", "git2go")
	_, err = Clone(url, path, &CloneOptions{})
	if err != nil {
		t.Fatal("cannot clone remote repo via https, error: ", err)
	}
}
