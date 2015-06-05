package git

import (
	"fmt"
	"reflect"
	"testing"
	"time"
)

func TestCreateNote(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitId)
	checkFatal(t, err)

	note, noteId := createTestNote(t, repo, commit)

	compareStrings(t, "I am a note\n", note.Message())
	compareStrings(t, noteId.String(), note.Id().String())
	compareStrings(t, "alice", note.Author().Name)
	compareStrings(t, "alice@example.com", note.Author().Email)
	compareStrings(t, "alice", note.Committer().Name)
	compareStrings(t, "alice@example.com", note.Committer().Email)
}

func TestNoteIterator(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	notes := make([]*Note, 5)
	for i := range notes {
		commitId, _ := updateReadme(t, repo, fmt.Sprintf("README v%d\n", i+1))
		commit, err := repo.LookupCommit(commitId)
		checkFatal(t, err)

		note, _ := createTestNote(t, repo, commit)
		notes[i] = note
	}

	iter, err := repo.NewNoteIterator("")
	checkFatal(t, err)
	for {
		noteId, commitId, err := iter.Next()
		if err != nil {
			if !IsErrorCode(err, ErrIterOver) {
				checkFatal(t, err)
			}
			break
		}

		note, err := repo.ReadNote("", commitId)
		checkFatal(t, err)

		if !reflect.DeepEqual(note.Id(), noteId) {
			t.Errorf("expected note oid '%v', actual '%v'", note.Id(), noteId)
		}
	}
}

func TestRemoveNote(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitId)
	checkFatal(t, err)

	note, _ := createTestNote(t, repo, commit)

	_, err = repo.ReadNote("", commit.Id())
	checkFatal(t, err)

	err = repo.RemoveNote("", note.Author(), note.Committer(), commitId)
	checkFatal(t, err)

	_, err = repo.ReadNote("", commit.Id())
	if err == nil {
		t.Fatal("note remove failed")
	}
}

func TestDefaultNoteRef(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	ref, err := repo.DefaultNoteRef()
	checkFatal(t, err)

	compareStrings(t, "refs/notes/commits", ref)
}

func createTestNote(t *testing.T, repo *Repository, commit *Commit) (*Note, *Oid) {
	loc, err := time.LoadLocation("Europe/Berlin")
	sig := &Signature{
		Name:  "alice",
		Email: "alice@example.com",
		When:  time.Date(2015, 01, 05, 13, 0, 0, 0, loc),
	}

	noteId, err := repo.CreateNote("", sig, sig, commit.Id(), "I am a note\n", false)
	checkFatal(t, err)

	note, err := repo.ReadNote("", commit.Id())
	checkFatal(t, err)

	return note, noteId
}
