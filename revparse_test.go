package git

import (
	"testing"
)

func TestRevparse(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, _ := seedTestRepo(t, repo)

	revSpec, err := repo.Revparse("HEAD")
	checkFatal(t, err)

	checkObject(t, revSpec.From(), commitID)
}

func TestRevparseSingle(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, _ := seedTestRepo(t, repo)

	obj, err := repo.RevparseSingle("HEAD")
	checkFatal(t, err)

	checkObject(t, obj, commitID)
}

func TestRevparseExt(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, treeID := seedTestRepo(t, repo)

	ref, err := repo.CreateReference("refs/heads/master", treeID, true, nil, "")
	checkFatal(t, err)

	obj, ref, err := repo.RevparseExt("master")
	checkFatal(t, err)

	checkObject(t, obj, treeID)
	if ref == nil {
		t.Fatalf("bad reference")
	}
}

func checkObject(t *testing.T, obj Object, id *Oid) {
	if obj == nil {
		t.Fatalf("bad object")
	}

	if !obj.Id().Equal(id) {
		t.Fatalf("bad object, expected %s, got %s", id.String(), obj.Id().String())
	}
}
