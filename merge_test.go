package git

import (
	"testing"
	"time"
)

func TestMergeWithSelf(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

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
	defer cleanupTestRepo(t, repo)

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
	defer cleanupTestRepo(t, repo)

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

func appendCommit(t *testing.T, repo *Repository) (*Oid, *Oid) {
	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is another commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)

	ref, err := repo.LookupReference("HEAD")
	checkFatal(t, err)

	parent, err := ref.Peel(ObjectCommit)
	checkFatal(t, err)

	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree, parent.(*Commit))
	checkFatal(t, err)

	return commitId, treeId
}

func TestMergeBase(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitAId, _ := seedTestRepo(t, repo)
	commitBId, _ := appendCommit(t, repo)

	mergeBase, err := repo.MergeBase(commitAId, commitBId)
	checkFatal(t, err)

	if mergeBase.Cmp(commitAId) != 0 {
		t.Fatalf("unexpected merge base")
	}

	mergeBases, err := repo.MergeBases(commitAId, commitBId)
	checkFatal(t, err)

	if len(mergeBases) != 1 {
		t.Fatalf("expected merge bases len to be 1, got %v", len(mergeBases))
	}

	if mergeBases[0].Cmp(commitAId) != 0 {
		t.Fatalf("unexpected merge base")
	}
}

func compareBytes(t *testing.T, expected, actual []byte) {
	for i, v := range expected {
		if actual[i] != v {
			t.Fatalf("Bad bytes")
		}
	}
}
