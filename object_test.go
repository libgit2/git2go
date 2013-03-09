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

	obj, err = repo.LookupObject(tree.Id())
	checkFatal(t, err)

	_, ok = obj.(*Tree)
	if !ok {
		t.Fatalf("LookupObject creates the wrong type")
	}

	if obj.Type() != OBJ_TREE {
		t.Fatalf("Type() doesn't agree with dynamic type")
	}
}
