package git

import (
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
	"testing"
)

func TestFindSimilar(t *testing.T) {
	t.Parallel()
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
	t.Parallel()
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

	files := make([]string, 0)
	hunks := make([]DiffHunk, 0)
	lines := make([]DiffLine, 0)
	patches := make([]string, 0)
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
	_, originalTreeId := seedTestRepo(t, repo)
	originalTree, err = repo.LookupTree(originalTreeId)

	checkFatal(t, err)

	_, newTreeId := updateReadme(t, repo, "file changed\n")

	newTree, err = repo.LookupTree(newTreeId)
	checkFatal(t, err)

	return originalTree, newTree
}

func TestDiffBlobs(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	odb, err := repo.Odb()
	checkFatal(t, err)

	id1, err := odb.Write([]byte("hello\nhello\n"), ObjectBlob)
	checkFatal(t, err)

	id2, err := odb.Write([]byte("hallo\nhallo\n"), ObjectBlob)
	checkFatal(t, err)

	blob1, err := repo.LookupBlob(id1)
	checkFatal(t, err)

	blob2, err := repo.LookupBlob(id2)
	checkFatal(t, err)

	var files, hunks, lines int
	err = DiffBlobs(blob1, "hi", blob2, "hi", nil,
		func(delta DiffDelta, progress float64) (DiffForEachHunkCallback, error) {
			files++
			return func(hunk DiffHunk) (DiffForEachLineCallback, error) {
				hunks++
				return func(line DiffLine) error {
					lines++
					return nil
				}, nil
			}, nil
		},
		DiffDetailLines)

	if files != 1 {
		t.Fatal("Bad number of files iterated")
	}

	if hunks != 1 {
		t.Fatal("Bad number of hunks iterated")
	}

	// two removals, two additions
	if lines != 4 {
		t.Fatalf("Bad number of lines iterated")
	}
}

func Test_ApplyDiff_Addfile(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	addFirstFileCommit, addFileTree := addAndGetTree(t, repo, "file1", `hello`)
	addSecondFileCommit, addSecondFileTree := addAndGetTree(t, repo, "file2", `hello2`)

	diff, err := repo.DiffTreeToTree(addFileTree, addSecondFileTree, nil)
	checkFatal(t, err)

	opts := StatusOptions{
		Show:  StatusShowIndexAndWorkdir,
		Flags: StatusOptIncludeUntracked,
	}

	t.Run("check does not apply to current tree because file exists", func(t *testing.T) {
		err = repo.ResetToCommit(addSecondFileCommit, ResetHard, &CheckoutOpts{})
		checkFatal(t, err)

		err = repo.ApplyDiff(diff, GitApplyLocationBoth, nil)
		if err == nil {
			t.Error("expecting applying patch to current repo to fail")
		}
	})

	t.Run("check apply to correct commit", func(t *testing.T) {
		err = repo.ResetToCommit(addFirstFileCommit, ResetHard, &CheckoutOpts{})
		checkFatal(t, err)

		err = repo.ApplyDiff(diff, GitApplyLocationBoth, nil)
		checkFatal(t, err)

		t.Run("Check that diff only changed one file", func(t *testing.T) {
			statuses, err := repo.StatusList(&opts)
			checkFatal(t, err)

			count, err := statuses.EntryCount()
			checkFatal(t, err)

			if count != 1 {
				t.Error("diff should affect exactly one file")
			}
			if count == 0 {
				t.Fatal("no statuses, cannot continue test")
			}

			entry, err := statuses.ByIndex(0)
			checkFatal(t, err)

			if entry.Status != StatusIndexNew {
				t.Error("status should be 'new' as file has been added between commits")
			}

			if entry.HeadToIndex.NewFile.Path != "file2" {
				t.Error("new file should be 'file2")
			}

			index, err := repo.Index()
			checkFatal(t, err)
			defer index.Free()

			newTreeOID, err := index.WriteTreeTo(repo)
			checkFatal(t, err)

			newTree, err := repo.LookupTree(newTreeOID)
			checkFatal(t, err)
			defer newTree.Free()

			_, err = repo.CreateCommit("HEAD", signature(), signature(), fmt.Sprintf("patch apply"), newTree, addFirstFileCommit)
			checkFatal(t, err)
		})

		t.Run("test applying patch produced the same diff", func(t *testing.T) {
			head, err := repo.Head()
			checkFatal(t, err)

			commit, err := repo.LookupCommit(head.Target())
			checkFatal(t, err)

			tree, err := commit.Tree()
			checkFatal(t, err)

			newDiff, err := repo.DiffTreeToTree(addFileTree, tree, nil)
			checkFatal(t, err)

			raw1b, err := diff.ToBuf(DiffFormatPatch)
			checkFatal(t, err)
			raw2b, err := newDiff.ToBuf(DiffFormatPatch)
			checkFatal(t, err)

			raw1 := string(raw1b)
			raw2 := string(raw2b)

			if raw1 != raw2 {
				t.Error("diffs should be the same")
			}
		})
	})

	t.Run("check convert to raw buffer and apply", func(t *testing.T) {
		err = repo.ResetToCommit(addFirstFileCommit, ResetHard, &CheckoutOpts{})
		checkFatal(t, err)

		raw, err := diff.ToBuf(DiffFormatPatch)
		checkFatal(t, err)

		if len(raw) == 0 {
			t.Error("empty diff created")
		}

		diff2, err := DiffFromBuffer(raw, repo)
		checkFatal(t, err)

		err = repo.ApplyDiff(diff2, GitApplyLocationBoth, nil)
		checkFatal(t, err)
	})
}

func addAndGetTree(t *testing.T, repo *Repository, filename string, content string) (*Commit, *Tree) {
	headCommit, err := headCommit(repo)
	checkFatal(t, err)
	defer headCommit.Free()

	p := repo.Path()
	p = strings.TrimSuffix(p, ".git")
	p = strings.TrimSuffix(p, ".git/")

	err = ioutil.WriteFile(path.Join(p, filename), []byte((content)), 0777)
	checkFatal(t, err)

	index, err := repo.Index()
	checkFatal(t, err)
	defer index.Free()

	err = index.AddByPath(filename)
	checkFatal(t, err)

	newTreeOID, err := index.WriteTreeTo(repo)
	checkFatal(t, err)

	newTree, err := repo.LookupTree(newTreeOID)
	checkFatal(t, err)
	defer newTree.Free()

	commitId, err := repo.CreateCommit("HEAD", signature(), signature(), fmt.Sprintf("add %s", filename), newTree, headCommit)
	checkFatal(t, err)

	commit, err := repo.LookupCommit(commitId)
	checkFatal(t, err)

	tree, err := commit.Tree()
	checkFatal(t, err)

	return commit, tree
}
