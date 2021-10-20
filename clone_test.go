package git

import (
	"github.com/sosedoff/gitkit"
	"io/ioutil"
	"net/http/httptest"
	"testing"
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

// StartHTTP starts a new HTTP git server with the current configuration.
func StartHTTP(repoDir string) (*httptest.Server, error) {
	service := gitkit.New(gitkit.Config{
		Dir:        repoDir,
		Auth:       false,
		Hooks:      &gitkit.HookScripts{},
		AutoCreate: false,
	})

	if err := service.Setup(); err != nil {
		return nil, err
	}
	server := httptest.NewServer(service)
	return server, nil
}

// TestCloneWithExternalHTTPUrWithLocalServer
func TestCloneWithExternalHTTPUrWithLocalServer(t *testing.T) {

	// create an empty repo
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	// initialize a git server at the path of the newly created repo
	serv, err := StartHTTP(repo.Workdir())
	checkFatal(t, err)

	// clone the repo
	url := serv.URL + "/.git"
	path, err := ioutil.TempDir("", "git2go")
	_, err = Clone(url, path, &CloneOptions{})
	if err != nil {
		t.Fatal("cannot clone remote repo via https, error: ", err)
	}
}
