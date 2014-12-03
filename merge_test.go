package git

import (
	"os"
	"testing"
)

func TestMergeWithSelf(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)

	master, err := repo.LookupReference("refs/heads/master")
	checkFatal(t, err)

	mergeHead, err := repo.AnnotatedCommitFromRef(master)
	checkFatal(t, err)

	mergeHeads := make([]*AnnotatedCommit, 1)
	mergeHeads[0] = mergeHead
	err = repo.Merge(mergeHeads, nil, nil)
	checkFatal(t, err)
}

func TestMergeAnalysisWithSelf(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)

	master, err := repo.LookupReference("refs/heads/master")
	checkFatal(t, err)

	mergeHead, err := repo.AnnotatedCommitFromRef(master)
	checkFatal(t, err)

	mergeHeads := make([]*AnnotatedCommit, 1)
	mergeHeads[0] = mergeHead
	a, _, err := repo.MergeAnalysis(mergeHeads)
	checkFatal(t, err)

	if a != MergeAnalysisUpToDate {
		t.Fatalf("Expected up to date merge, not %v", a)
	}
}

func TestMergeSameFile(t *testing.T) {

	file := MergeFileInput{
		Path:     "test",
		Mode:     33188,
		Contents: []byte("hello world"),
	}

	result, err := MergeFile(file, file, file, nil)
	checkFatal(t, err)
	if !result.Automergeable {
		t.Fatal("expected automergeable")
	}
	if result.Path != file.Path {
		t.Fatal("path was incorrect")
	}
	if result.Mode != file.Mode {
		t.Fatal("mode was incorrect")
	}

	compareBytes(t, file.Contents, result.Contents)

}
func TestMergeTreesWithoutAncestor(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	_, originalTreeId := seedTestRepo(t, repo)
	originalTree, err := repo.LookupTree(originalTreeId)

	checkFatal(t, err)

	_, newTreeId := updateReadme(t, repo, "file changed\n")

	newTree, err := repo.LookupTree(newTreeId)
	checkFatal(t, err)
	index, err := repo.MergeTrees(nil, originalTree, newTree, nil)
	if !index.HasConflicts() {
		t.Fatal("expected conflicts in the index")
	}
	_, err = index.GetConflict("README")
	checkFatal(t, err)

}

func compareBytes(t *testing.T, expected, actual []byte) {
	for i, v := range expected {
		if actual[i] != v {
			t.Fatalf("Bad bytes")
		}
	}
}
