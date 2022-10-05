package git

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/packet"
)

func TestCreateTag(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

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

func TestCreateTagWithSignature(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	tagContent := repo.Tags.CreateTagBuffer("v0.1.0", commit, &Signature{
		Name:  "Alice",
		Email: "alice@example.com",
		When:  time.Now(),
	}, "This is a tag")

	entity, err := openpgp.NewEntity("Alice", "test comment", "alice@example.com", nil)
	checkFatal(t, err)

	pgpSignatureBuf := new(bytes.Buffer)
	err = openpgp.ArmoredDetachSignText(pgpSignatureBuf, entity, strings.NewReader(string(tagContent)), &packet.Config{})
	checkFatal(t, err)

	tagID, err := repo.Tags.CreateTagWithSignature(string(tagContent), pgpSignatureBuf.String(), false)
	checkFatal(t, err)

	tag, err := repo.LookupTag(tagID)
	checkFatal(t, err)

	compareStrings(t, "v0.1.0", tag.Name())
	compareStrings(t, commitID.String(), tag.TargetId().String())
}

func TestCreateTagLightweight(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	tagID, err := repo.Tags.CreateLightweight("v0.1.0", commit, false)
	checkFatal(t, err)

	_, err = repo.Tags.CreateLightweight("v0.1.0", commit, true)
	checkFatal(t, err)

	ref, err := repo.References.Lookup("refs/tags/v0.1.0")
	checkFatal(t, err)

	compareStrings(t, "refs/tags/v0.1.0", ref.Name())
	compareStrings(t, "v0.1.0", ref.Shorthand())
	compareStrings(t, tagID.String(), commitID.String())
	compareStrings(t, commitID.String(), ref.Target().String())
}

func TestListTags(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	createTag(t, repo, commit, "v1.0.1", "Release v1.0.1")

	commitID, _ = updateReadme(t, repo, "Release version 2")

	commit, err = repo.LookupCommit(commitID)
	checkFatal(t, err)

	createTag(t, repo, commit, "v2.0.0", "Release v2.0.0")

	expected := []string{
		"v1.0.1",
		"v2.0.0",
	}

	actual, err := repo.Tags.List()
	checkFatal(t, err)

	compareStringList(t, expected, actual)
}

func TestListTagsWithMatch(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	createTag(t, repo, commit, "v1.0.1", "Release v1.0.1")

	commitID, _ = updateReadme(t, repo, "Release version 2")

	commit, err = repo.LookupCommit(commitID)
	checkFatal(t, err)

	createTag(t, repo, commit, "v2.0.0", "Release v2.0.0")

	expected := []string{
		"v2.0.0",
	}

	actual, err := repo.Tags.ListWithMatch("v2*")
	checkFatal(t, err)

	compareStringList(t, expected, actual)

	expected = []string{
		"v1.0.1",
	}

	actual, err = repo.Tags.ListWithMatch("v1*")
	checkFatal(t, err)

	compareStringList(t, expected, actual)
}

func TestTagForeach(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	tag1 := createTag(t, repo, commit, "v1.0.1", "Release v1.0.1")

	commitID, _ = updateReadme(t, repo, "Release version 2")

	commit, err = repo.LookupCommit(commitID)
	checkFatal(t, err)

	tag2 := createTag(t, repo, commit, "v2.0.0", "Release v2.0.0")

	expectedNames := []string{
		"refs/tags/v1.0.1",
		"refs/tags/v2.0.0",
	}
	actualNames := []string{}
	expectedOids := []string{
		tag1.String(),
		tag2.String(),
	}
	actualOids := []string{}

	err = repo.Tags.Foreach(func(name string, id *Oid) error {
		actualNames = append(actualNames, name)
		actualOids = append(actualOids, id.String())
		return nil
	})
	checkFatal(t, err)

	compareStringList(t, expectedNames, actualNames)
	compareStringList(t, expectedOids, actualOids)

	fakeErr := errors.New("fake error")

	err = repo.Tags.Foreach(func(name string, id *Oid) error {
		return fakeErr
	})

	if err != fakeErr {
		t.Fatalf("Tags.Foreach() did not return the expected error, got %v", err)
	}
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

	tagId, err := repo.Tags.Create("v0.0.0", commit, sig, "This is a tag")
	checkFatal(t, err)
	return tagId
}

func createTag(t *testing.T, repo *Repository, commit *Commit, name, message string) *Oid {
	loc, err := time.LoadLocation("Europe/Bucharest")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	tagId, err := repo.Tags.Create(name, commit, sig, message)
	checkFatal(t, err)
	return tagId
}
