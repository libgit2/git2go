package git

import (
	"os"
	"reflect"
	"testing"
)

func TestBlame(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	commitId1, _ := seedTestRepo(t, repo)
	commitId2, _ := updateReadme(t, repo, "foo\nbar\nbaz\n")

	opts := BlameOptions{
		NewestCommit: commitId2,
		OldestCommit: nil,
		MinLine:      1,
		MaxLine:      3,
	}
	blame, err := repo.BlameFile("README", &opts)
	checkFatal(t, err)
	defer blame.Free()
	if blame.HunkCount() != 2 {
		t.Errorf("got hunk count %d, want 2", blame.HunkCount())
	}

	hunk1, err := blame.HunkByIndex(0)
	checkFatal(t, err)
	checkHunk(t, hunk1, BlameHunk{
		LinesInHunk:          1,
		FinalCommitId:        commitId1,
		FinalStartLineNumber: 1,
		OrigCommitId:         commitId1,
		OrigPath:             "README",
		OrigStartLineNumber:  1,
		Boundary:             true,
	})

	hunk2, err := blame.HunkByIndex(1)
	checkFatal(t, err)
	checkHunk(t, hunk2, BlameHunk{
		LinesInHunk:          2,
		FinalCommitId:        commitId2,
		FinalStartLineNumber: 2,
		OrigCommitId:         commitId2,
		OrigPath:             "README",
		OrigStartLineNumber:  2,
		Boundary:             false,
	})
}

func checkHunk(t *testing.T, hunk, want BlameHunk) {
	hunk.FinalSignature = nil
	want.FinalSignature = nil
	hunk.OrigSignature = nil
	want.OrigSignature = nil
	if !reflect.DeepEqual(hunk, want) {
		t.Fatalf("got hunk %+v, want %+v", hunk, want)
	}
}
