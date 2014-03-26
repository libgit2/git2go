package git

import (
	"os"
	"runtime"
	"testing"
)

func TestCreateRepoAndStage(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	if treeId.String() != "b7119b11e8ef7a1a5a34d3ac87f5b075228ac81e" {
		t.Fatalf("%v", treeId.String())
	}
}

func TestIndexWriteTreeTo(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	repo2 := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	treeId, err := idx.WriteTreeTo(repo2)
	checkFatal(t, err)

	if treeId.String() != "b7119b11e8ef7a1a5a34d3ac87f5b075228ac81e" {
		t.Fatalf("%v", treeId.String())
	}
}

func checkFatal(t *testing.T, err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatal()
	}

	t.Fatalf("Fail at %v:%v; %v", file, line, err)
}
