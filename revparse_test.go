package git

import (
	"os"
	"testing"
)

func TestRevParse(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	commitId, _ := seedTestRepo(t, repo)

	revSpec, err := repo.RevParse("HEAD")
	checkFatal(t, err)

	checkObject(t, revSpec.From(), commitId)
}

func TestRevParseSingle(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	commitId, _ := seedTestRepo(t, repo)

	obj, err := repo.RevParseSingle("HEAD")
	checkFatal(t, err)

	checkObject(t, obj, commitId)
}

func TestRevParseExt(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	_, treeId := seedTestRepo(t, repo)

	ref, err := repo.CreateReference("refs/heads/master", treeId, true, nil, "")
	checkFatal(t, err)

	obj, ref, err := repo.RevParseExt("master")
	checkFatal(t, err)

	checkObject(t, obj, treeId)
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
