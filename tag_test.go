package git

import (
	"testing"
	"time"
)

func TestCreateTag(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	tagID := createTestTag(t, repo, commit)

	tag, err := repo.LookupTag(tagID)
	checkFatal(t, err)

	compareStrings(t, "v0.0.0", tag.Name())
	compareStrings(t, "This is a tag", tag.Message())
	compareStrings(t, commitID.String(), tag.TargetId().String())
}

func compareStrings(t *testing.T, expected, value string) {
	if value != expected {
		t.Fatalf("expected '%v', actual '%v'", expected, value)
	}
}

func createTestTag(t *testing.T, repo *Repository, commit *Commit) *Oid {
	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	tagID, err := repo.CreateTag("v0.0.0", commit, sig, "This is a tag")
	checkFatal(t, err)
	return tagID
}
