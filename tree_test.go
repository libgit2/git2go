package git

import "testing"

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
	var entryName string = "testEntry"
	var expect int = 9001
	var found bool = false

	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, treeID := seedTestRepo(t, repo)

	tree, err := repo.LookupTree(treeID)
	checkFatal(t, err)

	treeBuilder, err := repo.TreeBuilderFromTree(tree)
	checkFatal(t, err)

	defer treeBuilder.Free()

	odb, err := repo.Odb()
	checkFatal(t, err)

	blobId, err := odb.Write([]byte("hello, walk."), ObjectBlob)
	checkFatal(t, err)

	err = treeBuilder.Insert(entryName, blobId, FilemodeBlobExecutable)
	checkFatal(t, err)

	newTreeId, err := treeBuilder.Write()
	checkFatal(t, err)

	tree, err = repo.LookupTree(newTreeId)
	checkFatal(t, err)

	callback := func(e string, te *TreeEntry, ctx interface{}) int {
		if ctx.(int) != expect {
			t.Fatalf("Expected %d in all callback payloads", expect)
		}
		if te.Name == entryName {
			found = true
		}
		return 0
	}

	tree.Walk(callback, expect)

	if !found {
		t.Fatalf("Expected `%s` in subTree", entryName)
	}
}