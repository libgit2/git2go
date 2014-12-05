package git

import (
	"errors"
	"os"
	"path"
	"testing"
	"time"
)

func TestDiffTreeToTree(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	_, originalTreeId := seedTestRepo(t, repo)
	originalTree, err := repo.LookupTree(originalTreeId)

	checkFatal(t, err)

	_, newTreeId := updateReadme(t, repo, "file changed\n")

	newTree, err := repo.LookupTree(newTreeId)
	checkFatal(t, err)

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

	errTest := errors.New("test error")

	err = diff.ForEach(func(file DiffDelta, progress float64) (DiffForEachHunkCallback, error) {
		return nil, errTest
	}, DiffDetailLines)

	if err != errTest {
		t.Fatal("Expected custom error to be returned")
	}

}

func TestDiffFindSimilar(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	_, originalTreeId := seedTestRepo(t, repo)
	originalTree, err := repo.LookupTree(originalTreeId)
	checkFatal(t, err)

	// rename README -> README.RENAMED
	wd := repo.Workdir()
	orig := "README"
	renamed := "README.RENAMED"
	err = os.Rename(path.Join(wd, orig), path.Join(wd, renamed))
	checkFatal(t, err)

	// add changes to index
	idx, err := repo.Index()
	checkFatal(t, err)
	//err = idx.AddAll([]string{"*"}, IndexAddDefault, func(a, b string) int {
	//t.Logf("IndexMatchedPathCB: %s %s\n", a, b)
	//return 0
	//})
	idx.RemoveByPath(orig)
	checkFatal(t, err)
	idx.AddByPath(renamed)
	checkFatal(t, err)
	newTreeId, err := idx.WriteTree()
	checkFatal(t, err)

	currentBranch, err := repo.Head()
	checkFatal(t, err)
	currentTip, err := repo.LookupCommit(currentBranch.Target())
	checkFatal(t, err)

	// commit changes
	newTree, err := repo.LookupTree(newTreeId)
	checkFatal(t, err)
	_, err = repo.CreateCommit("HEAD", sig, sig, "Rename\n", newTree, currentTip)
	checkFatal(t, err)

	// Find Similarity
	diff, err := repo.DiffTreeToTree(originalTree, newTree, nil)
	checkFatal(t, err)
	diff.FindSimilar(nil)

	// Check results
	diff.ForEach(func(delta DiffDelta, progress float64) (DiffForEachHunkCallback, error) {
		if delta.Status != DeltaRenamed {
			t.Fatal("Expected status to be 'DeltaRenamed'")
		}
		if delta.Similarity != 100 {
			t.Fatalf("Expected similarity to be 100, was %d\n", delta.Similarity)
		}
		return nil, nil
	}, DiffDetailFiles)
}
