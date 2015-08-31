package git

import (
	"io/ioutil"
	"testing"
)

const (
	REMOTENAME = "testremote"
)

func TestClone(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)

	repo2, err := Clone(repo.Path(), path, &CloneOptions{Bare: true})
	defer cleanupTestRepo(t, repo2)

	checkFatal(t, err)
}

func TestCloneWithCallback(t *testing.T) {
	testPayload := 0

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	path, err := ioutil.TempDir("", "git2go")
	checkFatal(t, err)

	opts := CloneOptions{
		Bare: true,
		RemoteCreateCallback: func(r Repository, name, url string) (*Remote, ErrorCode) {
			testPayload += 1

			remote, err := r.CreateRemote(REMOTENAME, url)
			if err != nil {
				return nil, ErrGeneric
			}

			return remote, ErrOk
		},
	}

	repo2, err := Clone(repo.Path(), path, &opts)
	defer cleanupTestRepo(t, repo2)

	checkFatal(t, err)

	if testPayload != 1 {
		t.Fatal("Payload's value has not been changed")
	}

	remote, err := repo2.LookupRemote(REMOTENAME)
	if err != nil || remote == nil {
		t.Fatal("Remote was not created properly")
	}
}
