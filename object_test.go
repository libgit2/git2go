package git

import (
	"testing"
)

func TestObjectPoymorphism(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, treeId := seedTestRepo(t, repo)

	var obj *Object

	commit, err := repo.LookupCommit(commitId)
	checkFatal(t, err)

	obj = &commit.Object
	if obj.Type() != ObjectCommit {
		t.Fatalf("Wrong object type, expected commit, have %v", obj.Type())
	}

	commitTree, err := commit.Tree()
	checkFatal(t, err)
	commitTree.EntryCount()

	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)

	obj = &tree.Object
	if obj.Type() != ObjectTree {
		t.Fatalf("Wrong object type, expected tree, have %v", obj.Type())
	}

	tree2, err := obj.AsTree()
	if err != nil {
		t.Fatalf("Converting back to *Tree is not ok")
	}

	entry := tree2.EntryByName("README")
	if entry == nil {
		t.Fatalf("Tree did not have expected \"README\" entry")
	}

	if entry.Filemode != FilemodeBlob {
		t.Fatal("Wrong filemode for \"README\"")
	}

	_, err = obj.AsCommit()
	if err == nil {
		t.Fatalf("*Tree is somehow the same as *Commit")
	}

	obj, err = repo.Lookup(tree.Id())
	checkFatal(t, err)

	_, err = obj.AsTree()
	if err != nil {
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

	checkOwner(t, repo, commit.Object)
	checkOwner(t, repo, tree.Object)
}

func TestObjectPeel(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, treeID := seedTestRepo(t, repo)

	var obj *Object

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	obj, err = commit.Peel(ObjectAny)
	checkFatal(t, err)

	if obj.Type() != ObjectTree {
		t.Fatalf("Wrong object type when peeling a commit, expected tree, have %v", obj.Type())
	}

	obj, err = commit.Peel(ObjectTag)

	if !IsErrorCode(err, ErrInvalidSpec) {
		t.Fatalf("Wrong error when peeling a commit to a tag, expected ErrInvalidSpec, have %v", err)
	}

	tree, err := repo.LookupTree(treeID)
	checkFatal(t, err)

	obj, err = tree.Peel(ObjectAny)

	if !IsErrorCode(err, ErrInvalidSpec) {
		t.Fatalf("Wrong error when peeling a tree, expected ErrInvalidSpec, have %v", err)
	}

	entry := tree.EntryByName("README")

	blob, err := repo.LookupBlob(entry.Id)
	checkFatal(t, err)

	obj, err = blob.Peel(ObjectAny)

	if !IsErrorCode(err, ErrInvalidSpec) {
		t.Fatalf("Wrong error when peeling a blob, expected ErrInvalidSpec, have %v", err)
	}

	tagID := createTestTag(t, repo, commit)

	tag, err := repo.LookupTag(tagID)
	checkFatal(t, err)

	obj, err = tag.Peel(ObjectAny)
	checkFatal(t, err)

	if obj.Type() != ObjectCommit {
		t.Fatalf("Wrong object type when peeling a tag, expected commit, have %v", obj.Type())
	}

	// TODO: Should test a tag that annotates a different object than a commit
	// but it's impossible at the moment to tag such an object.
}
