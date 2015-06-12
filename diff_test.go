package git

import (
	"errors"
	"strings"
	"testing"
)

func TestFindSimilar(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

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
	defer cleanupTestRepo(t, repo)

	originalTree, newTree := createTestTrees(t, repo)

	callbackInvoked := false
	opts := DiffOptions{
		NotifyCallback: func(diffSoFar *Diff, delta DiffDelta, matchedPathSpec string) error {
			callbackInvoked = true
			return nil
		},
		OldPrefix: "x1/",
		NewPrefix: "y1/",
	}

	diff, err := repo.DiffTreeToTree(originalTree, newTree, &opts)
	checkFatal(t, err)
	if !callbackInvoked {
		t.Fatal("callback not invoked")
	}

	if diff == nil {
		t.Fatal("no diff returned")
	}

	var files []string
	var hunks []DiffHunk
	var lines []DiffLine
	var patches []string
	err = diff.ForEach(func(file DiffDelta, progress float64) (DiffForEachHunkCallback, error) {
		patch, err := diff.Patch(len(patches))
		if err != nil {
			return nil, err
		}
		defer patch.Free()
		patchStr, err := patch.String()
		if err != nil {
			return nil, err
		}
		patches = append(patches, patchStr)

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

	if want1, want2 := "x1/README", "y1/README"; !strings.Contains(patches[0], want1) || !strings.Contains(patches[0], want2) {
		t.Errorf("Diff patch doesn't contain %q or %q\n\n%s", want1, want2, patches[0])

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
	_, originalTreeID := seedTestRepo(t, repo)
	originalTree, err = repo.LookupTree(originalTreeID)

	checkFatal(t, err)

	_, newTreeID := updateReadme(t, repo, "file changed\n")

	newTree, err = repo.LookupTree(newTreeID)
	checkFatal(t, err)

	return originalTree, newTree
}
