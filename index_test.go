package git

import (
	"io/ioutil"
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

func TestIndexAddAndWriteTreeTo(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	odb, err := repo.Odb()
	checkFatal(t, err)

	blobID, err := odb.Write([]byte("foo\n"), ObjectBlob)
	checkFatal(t, err)

	idx, err := NewIndex()
	checkFatal(t, err)

	entry := IndexEntry{
		Path: "README",
		Id:   blobID,
		Mode: FilemodeBlob,
	}

	err = idx.Add(&entry)
	checkFatal(t, err)

	treeId, err := idx.WriteTreeTo(repo)
	checkFatal(t, err)

	if treeId.String() != "b7119b11e8ef7a1a5a34d3ac87f5b075228ac81e" {
		t.Fatalf("%v", treeId.String())
	}
}

func TestIndexAddAllNoCallback(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	err := ioutil.WriteFile(repo.Workdir()+"/README", []byte("foo\n"), 0644)
	checkFatal(t, err)

	idx, err := repo.Index()
	checkFatal(t, err)

	err = idx.AddAll([]string{}, IndexAddDefault, nil)
	checkFatal(t, err)

	treeId, err := idx.WriteTreeTo(repo)
	checkFatal(t, err)

	if treeId.String() != "b7119b11e8ef7a1a5a34d3ac87f5b075228ac81e" {
		t.Fatalf("%v", treeId.String())
	}
}

func TestIndexAddAllCallback(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	err := ioutil.WriteFile(repo.Workdir()+"/README", []byte("foo\n"), 0644)
	checkFatal(t, err)

	idx, err := repo.Index()
	checkFatal(t, err)

	cbPath := ""
	err = idx.AddAll([]string{}, IndexAddDefault, func(p, mP string) int {
		cbPath = p
		return 0
	})
	checkFatal(t, err)
	if cbPath != "README" {
		t.Fatalf("%v", cbPath)
	}

	treeId, err := idx.WriteTreeTo(repo)
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
