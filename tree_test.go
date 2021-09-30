package git

import (
	"errors"
	"testing"
)

func TestTreeEntryById(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, treeID := seedTestRepo(t, repo)

	tree, err := repo.LookupTree(treeID)
	checkFatal(t, err)

	id, err := NewOid("257cc5642cb1a054f08cc83f2d943e56fd3ebe99")
	checkFatal(t, err)

	entry := tree.EntryById(id)

	if entry == nil {
		t.Fatalf("entry id %v was not found", id)
	}
}

func TestTreeBuilderInsert(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	subTree, err := repo.TreeBuilder()
	if err != nil {
		t.Fatalf("TreeBuilder: %v", err)
	}
	defer subTree.Free()

	odb, err := repo.Odb()
	if err != nil {
		t.Fatalf("repo.Odb: %v", err)
	}
	blobId, err := odb.Write([]byte("hello"), ObjectBlob)
	if err != nil {
		t.Fatalf("odb.Write: %v", err)
	}
	if err = subTree.Insert("subfile", blobId, FilemodeBlobExecutable); err != nil {
		t.Fatalf("TreeBuilder.Insert: %v", err)
	}
	treeID, err := subTree.Write()
	if err != nil {
		t.Fatalf("TreeBuilder.Write: %v", err)
	}

	tree, err := repo.LookupTree(treeID)
	if err != nil {
		t.Fatalf("LookupTree: %v", err)
	}

	entry, err := tree.EntryByPath("subfile")
	if err != nil {
		t.Fatalf("tree.EntryByPath(%q): %v", "subfile", err)
	}

	if !entry.Id.Equal(blobId) {
		t.Fatalf("got oid %v, want %v", entry.Id, blobId)
	}
}

func TestTreeWalk(t *testing.T) {
	t.Parallel()
	repo, err := OpenRepository("testdata/TestGitRepository.git")
	checkFatal(t, err)
	treeID, err := NewOid("6020a3b8d5d636e549ccbd0c53e2764684bb3125")
	checkFatal(t, err)

	tree, err := repo.LookupTree(treeID)
	checkFatal(t, err)

	var callCount int
	err = tree.Walk(func(name string, entry *TreeEntry) error {
		callCount++

		return nil
	})
	checkFatal(t, err)
	if callCount != 11 {
		t.Fatalf("got called %v times, want %v", callCount, 11)
	}
}

func TestTreeWalkSkip(t *testing.T) {
	t.Parallel()
	repo, err := OpenRepository("testdata/TestGitRepository.git")
	checkFatal(t, err)
	treeID, err := NewOid("6020a3b8d5d636e549ccbd0c53e2764684bb3125")
	checkFatal(t, err)

	tree, err := repo.LookupTree(treeID)
	checkFatal(t, err)

	var callCount int
	err = tree.Walk(func(name string, entry *TreeEntry) error {
		callCount++

		return TreeWalkSkip
	})
	checkFatal(t, err)
	if callCount != 4 {
		t.Fatalf("got called %v times, want %v", callCount, 4)
	}
}

func TestTreeWalkStop(t *testing.T) {
	t.Parallel()
	repo, err := OpenRepository("testdata/TestGitRepository.git")
	checkFatal(t, err)
	treeID, err := NewOid("6020a3b8d5d636e549ccbd0c53e2764684bb3125")
	checkFatal(t, err)

	tree, err := repo.LookupTree(treeID)
	checkFatal(t, err)

	var callCount int
	stopError := errors.New("stop")
	err = tree.Walk(func(name string, entry *TreeEntry) error {
		callCount++

		return stopError
	})
	if err != stopError {
		t.Fatalf("got error %v, want %v", err, stopError)
	}
	if callCount != 1 {
		t.Fatalf("got called %v times, want %v", callCount, 1)
	}
}
