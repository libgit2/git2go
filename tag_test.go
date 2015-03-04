package git

import (
	"os"
	"testing"
	"time"
)

func TestCreateTag(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
	commitId, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitId)
	checkFatal(t, err)

	tagId := createTestTag(t, repo, commit)

	tag, err := repo.LookupTag(tagId)
	checkFatal(t, err)

	compareStrings(t, "v0.0.0", tag.Name())
	compareStrings(t, "This is a tag", tag.Message())
	compareStrings(t, commitId.String(), tag.TargetId().String())
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

	tagId, err := repo.CreateTag("v0.0.0", commit, sig, "This is a tag")
	checkFatal(t, err)
	return tagId
}
