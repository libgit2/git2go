package git

import (
	"os"
	"testing"
)

func TestObjectPoymorphism(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
	commitId, treeId := seedTestRepo(t, repo)

	var obj Object

	commit, err := repo.LookupCommit(commitId)
	checkFatal(t, err)

	obj = commit
	if obj.Type() != OBJ_COMMIT {
		t.Fatalf("Wrong object type, expected commit, have %v", obj.Type())
	}

	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)

	obj = tree
	if obj.Type() != OBJ_TREE {
		t.Fatalf("Wrong object type, expected tree, have %v", obj.Type())
	}

	tree2, ok := obj.(*Tree)
	if !ok {
		t.Fatalf("Converting back to *Tree is not ok")
	}

	if tree2.EntryByName("README") == nil {
		t.Fatalf("Tree did not have expected \"README\" entry")
	}

	_, ok = obj.(*Commit)
	if ok {
		t.Fatalf("*Tree is somehow the same as *Commit")
	}

	obj, err = repo.Lookup(tree.Id())
	checkFatal(t, err)

	_, ok = obj.(*Tree)
	if !ok {
		t.Fatalf("Lookup creates the wrong type")
	}

	if obj.Type() != OBJ_TREE {
		t.Fatalf("Type() doesn't agree with dynamic type")
	}

	obj, err = repo.RevparseSingle("HEAD")
	checkFatal(t, err)
	if obj.Type() != OBJ_COMMIT || obj.Id().String() != commit.Id().String() {
		t.Fatalf("Failed to parse the right revision")
	}

	obj, err = repo.RevparseSingle("HEAD^{tree}")
	checkFatal(t, err)
	if obj.Type() != OBJ_TREE || obj.Id().String() != tree.Id().String() {
		t.Fatalf("Failed to parse the right revision")
	}
}
