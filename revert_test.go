package git

import (
	"testing"
)

const (
	expectedRevertedReadmeContents = "foo\n"
)

func TestRevert(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)
	commitID, _ := updateReadme(t, repo, content)

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	revertOptions, err := DefaultRevertOptions()
	checkFatal(t, err)

	err = repo.Revert(commit, &revertOptions)
	checkFatal(t, err)

	actualReadmeContents := readReadme(t, repo)

	if actualReadmeContents != expectedRevertedReadmeContents {
		t.Fatalf(`README has incorrect contents after revert. Expected: "%v", Actual: "%v"`,
			expectedRevertedReadmeContents, actualReadmeContents)
	}

	state := repo.State()
	if state != RepositoryStateRevert {
		t.Fatalf("Incorrect repository state. Expected: %v, Actual: %v", RepositoryStateRevert, state)
	}

	err = repo.StateCleanup()
	checkFatal(t, err)

	state = repo.State()
	if state != RepositoryStateNone {
		t.Fatalf("Incorrect repository state. Expected: %v, Actual: %v", RepositoryStateNone, state)
	}
}

func TestRevertCommit(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)
	commitID, _ := updateReadme(t, repo, content)

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	revertOptions, err := DefaultRevertOptions()
	checkFatal(t, err)

	index, err := repo.RevertCommit(commit, commit, 0, &revertOptions.MergeOptions)
	checkFatal(t, err)
	defer index.Free()

	err = repo.CheckoutIndex(index, &revertOptions.CheckoutOptions)
	checkFatal(t, err)

	actualReadmeContents := readReadme(t, repo)

	if actualReadmeContents != expectedRevertedReadmeContents {
		t.Fatalf(`README has incorrect contents after revert. Expected: "%v", Actual: "%v"`,
			expectedRevertedReadmeContents, actualReadmeContents)
	}
}
