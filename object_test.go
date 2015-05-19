package git

import (
	"testing"
)

func TestObjectPoymorphism(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, treeId := seedTestRepo(t, repo)

	var obj Object

	commit, err := repo.LookupCommit(commitId)
	checkFatal(t, err)

	obj = commit
	if obj.Type() != ObjectCommit {
		t.Fatalf("Wrong object type, expected commit, have %v", obj.Type())
	}

	commitTree, err := commit.Tree()
	checkFatal(t, err)
	commitTree.EntryCount()

	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)

	obj = tree
	if obj.Type() != ObjectTree {
		t.Fatalf("Wrong object type, expected tree, have %v", obj.Type())
	}

	tree2, ok := obj.(*Tree)
	if !ok {
		t.Fatalf("Converting back to *Tree is not ok")
	}

	entry := tree2.EntryByName("README")
	if entry == nil {
		t.Fatalf("Tree did not have expected \"README\" entry")
	}

	if entry.Filemode != FilemodeBlob {
		t.Fatal("Wrong filemode for \"README\"")
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

	if obj.Type() != ObjectTree {
		t.Fatalf("Type() doesn't agree with dynamic type")
	}

	obj, err = repo.RevparseSingle("HEAD")
	checkFatal(t, err)
	if obj.Type() != ObjectCommit || obj.Id().String() != commit.Id().String() {
		t.Fatalf("Failed to parse the right revision")
	}

	obj, err = repo.RevparseSingle("HEAD^{tree}")
	checkFatal(t, err)
	if obj.Type() != ObjectTree || obj.Id().String() != tree.Id().String() {
		t.Fatalf("Failed to parse the right revision")
	}
}

func checkOwner(t *testing.T, repo *Repository, obj Object) {
	owner := obj.Owner()
	if owner == nil {
		t.Fatal("bad owner")
	}

	if owner.ptr != repo.ptr {
		t.Fatalf("bad owner, got %v expected %v\n", owner.ptr, repo.ptr)
	}
}

func TestObjectOwner(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, treeId := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitId)
	checkFatal(t, err)

	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)

	checkOwner(t, repo, commit)
	checkOwner(t, repo, tree)
}
