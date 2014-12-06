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

	wantHunk1 := BlameHunk{
		LinesInHunk:          1,
		FinalCommitId:        commitId1,
		FinalStartLineNumber: 1,
		OrigCommitId:         commitId1,
		OrigPath:             "README",
		OrigStartLineNumber:  1,
		Boundary:             true,
	}
	wantHunk2 := BlameHunk{
		LinesInHunk:          2,
		FinalCommitId:        commitId2,
		FinalStartLineNumber: 2,
		OrigCommitId:         commitId2,
		OrigPath:             "README",
		OrigStartLineNumber:  2,
		Boundary:             false,
	}

	hunk1, err := blame.HunkByIndex(0)
	checkFatal(t, err)
	checkHunk(t, "index 0", hunk1, wantHunk1)

	hunk2, err := blame.HunkByIndex(1)
	checkFatal(t, err)
	checkHunk(t, "index 1", hunk2, wantHunk2)

	hunkLine1, err := blame.HunkByLine(1)
	checkFatal(t, err)
	checkHunk(t, "line 1", hunkLine1, wantHunk1)

	hunkLine2, err := blame.HunkByLine(3)
	checkFatal(t, err)
	checkHunk(t, "line 2", hunkLine2, wantHunk2)
}

func checkHunk(t *testing.T, label string, hunk, want BlameHunk) {
	hunk.FinalSignature = nil
	want.FinalSignature = nil
	hunk.OrigSignature = nil
	want.OrigSignature = nil
	if !reflect.DeepEqual(hunk, want) {
		t.Fatalf("%s: got hunk %+v, want %+v", label, hunk, want)
	}
}
