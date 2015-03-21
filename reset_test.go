package git

import (
	"io/ioutil"
	"testing"
)

func TestResetToCommit(t *testing.T) {
	repo := createTestRepo(t)
	seedTestRepo(t, repo)
	// create commit to reset to
	commitId, _ := updateReadme(t, repo, "testing reset")
	// create commit to reset from
	nextCommitId, _ := updateReadme(t, repo, "will be reset")

	// confirm that we wrote "will be reset" to the readme
	newBytes, err := ioutil.ReadFile(pathInRepo(repo, "README"))
	checkFatal(t, err)
	if string(newBytes) != "will be reset" {
		t.Fatalf("expected %s to equal 'will be reset'", string(newBytes))
	}

	// confirm that the head of the repo is the next commit id
	head, err := repo.Head()
	checkFatal(t, err)
	if head.Target().String() != nextCommitId.String() {
		t.Fatalf(
			"expected to be at latest commit %s, but was %s",
			nextCommitId.String(),
			head.Target().String(),
		)
	}

	commitToResetTo, err := repo.LookupCommit(commitId)
	checkFatal(t, err)

	repo.ResetToCommit(commitToResetTo, ResetHard, &CheckoutOpts{})

	// check that the file now reads "testing reset" like it did before
	bytes, err := ioutil.ReadFile(pathInRepo(repo, "README"))
	checkFatal(t, err)
	if string(bytes) != "testing reset" {
		t.Fatalf("expected %s to equal 'testing reset'", string(bytes))
	}
}
