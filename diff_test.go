package git

import (
	"errors"
	"os"
	"testing"
)

func TestFindSimilar(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	originalTree, newTree := createTestTrees(t, repo)

	diffOpt, _ := DefaultDiffOptions()

	diff, err := repo.DiffTreeToTree(originalTree, newTree, &diffOpt)
	checkFatal(t, err)
	if diff == nil {
		t.Fatal("no diff returned")
	}

	findOpts, err := DefaultDiffFindOptions()
	checkFatal(t, err)
	findOpts.Flags = DiffFindBreakRewrites

	err = diff.FindSimilar(&findOpts)
	checkFatal(t, err)

	numDiffs := 0
	numAdded := 0
	numDeleted := 0

	err = diff.ForEach(func(file DiffDelta, progress float64) (DiffForEachHunkCallback, error) {
		numDiffs++

		switch file.Status {
		case DeltaAdded:
			numAdded++
		case DeltaDeleted:
			numDeleted++
		}

		return func(hunk DiffHunk) (DiffForEachLineCallback, error) {
			return func(line DiffLine) error {
				return nil
			}, nil
		}, nil
	}, DiffDetailLines)

	if numDiffs != 2 {
		t.Fatal("Incorrect number of files in diff")
	}
	if numAdded != 1 {
		t.Fatal("Incorrect number of new files in diff")
	}
	if numDeleted != 1 {
		t.Fatal("Incorrect number of deleted files in diff")
	}

}

func TestDiffTreeToTree(t *testing.T) {

	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	originalTree, newTree := createTestTrees(t, repo)

	callbackInvoked := false
	opts := DiffOptions{
		NotifyCallback: func(diffSoFar *Diff, delta DiffDelta, matchedPathSpec string) error {
			callbackInvoked = true
			return nil
		},
	}

	diff, err := repo.DiffTreeToTree(originalTree, newTree, &opts)
	checkFatal(t, err)
	if !callbackInvoked {
		t.Fatal("callback not invoked")
	}

	if diff == nil {
		t.Fatal("no diff returned")
	}

	files := make([]string, 0)
	hunks := make([]DiffHunk, 0)
	lines := make([]DiffLine, 0)
	err = diff.ForEach(func(file DiffDelta, progress float64) (DiffForEachHunkCallback, error) {
		files = append(files, file.OldFile.Path)
		return func(hunk DiffHunk) (DiffForEachLineCallback, error) {
			hunks = append(hunks, hunk)
			return func(line DiffLine) error {
				lines = append(lines, line)
				return nil
			}, nil
		}, nil
	}, DiffDetailLines)

	checkFatal(t, err)

	if len(files) != 1 {
		t.Fatal("Incorrect number of files in diff")
	}

	if files[0] != "README" {
		t.Fatal("File in diff was expected to be README")
	}

	if len(hunks) != 1 {
		t.Fatal("Incorrect number of hunks in diff")
	}

	if hunks[0].OldStart != 1 || hunks[0].NewStart != 1 {
		t.Fatal("Incorrect hunk")
	}

	if len(lines) != 2 {
		t.Fatal("Incorrect number of lines in diff")
	}

	if lines[0].Content != "foo\n" {
		t.Fatal("Incorrect lines in diff")
	}

	if lines[1].Content != "file changed\n" {
		t.Fatal("Incorrect lines in diff")
	}

	stats, err := diff.Stats()
	checkFatal(t, err)

	if stats.Insertions() != 1 {
		t.Fatal("Incorrect number of insertions in diff")
	}
	if stats.Deletions() != 1 {
		t.Fatal("Incorrect number of deletions in diff")
	}
	if stats.FilesChanged() != 1 {
		t.Fatal("Incorrect number of changed files in diff")
	}

	errTest := errors.New("test error")

	err = diff.ForEach(func(file DiffDelta, progress float64) (DiffForEachHunkCallback, error) {
		return nil, errTest
	}, DiffDetailLines)

	if err != errTest {
		t.Fatal("Expected custom error to be returned")
	}

}

func createTestTrees(t *testing.T, repo *Repository) (originalTree *Tree, newTree *Tree) {
	var err error
	_, originalTreeId := seedTestRepo(t, repo)
	originalTree, err = repo.LookupTree(originalTreeId)

	checkFatal(t, err)

	_, newTreeId := updateReadme(t, repo, "file changed\n")

	newTree, err = repo.LookupTree(newTreeId)
	checkFatal(t, err)

	return originalTree, newTree
}
