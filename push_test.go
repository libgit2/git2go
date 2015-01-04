package git

import (
	"os"
	"testing"
)

func TestRemotePush(t *testing.T) {
	repo := createBareTestRepo(t)
	defer os.RemoveAll(repo.Path())
	localRepo := createTestRepo(t)
	defer os.RemoveAll(localRepo.Workdir())

	remote, err := localRepo.CreateRemote("test_push", repo.Path())
	checkFatal(t, err)

	seedTestRepo(t, localRepo)

	err = remote.Push([]string{"refs/heads/master"}, nil, nil, "")
	checkFatal(t, err)

	_, err = localRepo.LookupReference("refs/remotes/test_push/master")
	checkFatal(t, err)

	_, err = repo.LookupReference("refs/heads/master")
	checkFatal(t, err)
}
