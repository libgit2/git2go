package git

import (
	"testing"
)

func TestRemotePush(t *testing.T) {
	repo := createBareTestRepo(t)
	defer cleanupTestRepo(t, repo)

	localRepo := createTestRepo(t)
	defer cleanupTestRepo(t, localRepo)

	remote, err := localRepo.Remotes.Create("test_push", repo.Path())
	checkFatal(t, err)

	seedTestRepo(t, localRepo)

	err = remote.Push([]string{"refs/heads/master"}, nil)
	checkFatal(t, err)

	_, err = localRepo.References.Lookup("refs/remotes/test_push/master")
	checkFatal(t, err)

	_, err = repo.References.Lookup("refs/heads/master")
	checkFatal(t, err)
}
