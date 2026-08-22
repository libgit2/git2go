package git

import (
	"testing"
)

func TestRemotePush(t *testing.T) {
	t.Parallel()
	repo := createBareTestRepo(t)
	defer cleanupTestRepo(t, repo)

	localRepo := createTestRepo(t)
	defer cleanupTestRepo(t, localRepo)

	remote, err := localRepo.Remotes.Create("test_push", repo.Path())
	checkFatal(t, err)
	defer remote.Free()

	seedTestRepo(t, localRepo)

	branchName := defaultBranchName(t, localRepo)
	err = remote.Push([]string{"refs/heads/" + branchName}, nil)
	checkFatal(t, err)

	ref, err := localRepo.References.Lookup("refs/remotes/test_push/" + branchName)
	checkFatal(t, err)
	defer ref.Free()

	ref, err = repo.References.Lookup("refs/heads/" + branchName)
	checkFatal(t, err)
	defer ref.Free()
}
