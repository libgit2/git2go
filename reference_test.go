package git

import (
	"runtime"
	"sort"
	"testing"
	"time"
)

func TestRefModification(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, treeId := seedTestRepo(t, repo)

	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}
	_, err = repo.CreateReference("refs/tags/tree", treeId, true, sig, "testTreeTag")
	checkFatal(t, err)

	tag, err := repo.LookupReference("refs/tags/tree")
	checkFatal(t, err)
	checkRefType(t, tag, ReferenceOid)

	ref, err := repo.LookupReference("HEAD")
	checkFatal(t, err)
	checkRefType(t, ref, ReferenceSymbolic)

	if target := ref.Target(); target != nil {
		t.Fatalf("Expected nil *Oid, got %v", target)
	}

	ref, err = ref.Resolve()
	checkFatal(t, err)
	checkRefType(t, ref, ReferenceOid)

	if target := ref.Target(); target == nil {
		t.Fatalf("Expected valid target got nil")
	}

	if target := ref.SymbolicTarget(); target != "" {
		t.Fatalf("Expected empty string, got %v", target)
	}

	if commitId.String() != ref.Target().String() {
		t.Fatalf("Wrong ref target")
	}

	_, err = tag.Rename("refs/tags/renamed", false, nil, "")
	checkFatal(t, err)
	tag, err = repo.LookupReference("refs/tags/renamed")
	checkFatal(t, err)
	checkRefType(t, ref, ReferenceOid)

}

func TestReferenceIterator(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	loc, err := time.LoadLocation("Europe/Berlin")
	checkFatal(t, err)
	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Date(2013, 03, 06, 14, 30, 0, 0, loc),
	}

	idx, err := repo.Index()
	checkFatal(t, err)
	err = idx.AddByPath("README")
	checkFatal(t, err)
	treeId, err := idx.WriteTree()
	checkFatal(t, err)

	message := "This is a commit\n"
	tree, err := repo.LookupTree(treeId)
	checkFatal(t, err)
	commitId, err := repo.CreateCommit("HEAD", sig, sig, message, tree)
	checkFatal(t, err)

	_, err = repo.CreateReference("refs/heads/one", commitId, true, sig, "headOne")
	checkFatal(t, err)

	_, err = repo.CreateReference("refs/heads/two", commitId, true, sig, "headTwo")
	checkFatal(t, err)

	_, err = repo.CreateReference("refs/heads/three", commitId, true, sig, "headThree")
	checkFatal(t, err)

	iter, err := repo.NewReferenceIterator()
	checkFatal(t, err)

	var list []string
	expected := []string{
		"refs/heads/master",
		"refs/heads/one",
		"refs/heads/three",
		"refs/heads/two",
	}

	// test some manual iteration
	nameIter := iter.Names()
	name, err := nameIter.Next()
	for err == nil {
		list = append(list, name)
		name, err = nameIter.Next()
	}
	if !IsErrorCode(err, ErrIterOver) {
		t.Fatal("Iteration not over")
	}

	sort.Strings(list)
	compareStringList(t, expected, list)

	// test the iterator for full refs, rather than just names
	iter, err = repo.NewReferenceIterator()
	checkFatal(t, err)
	count := 0
	_, err = iter.Next()
	for err == nil {
		count++
		_, err = iter.Next()
	}
	if !IsErrorCode(err, ErrIterOver) {
		t.Fatal("Iteration not over")
	}

	if count != 4 {
		t.Fatalf("Wrong number of references returned %v", count)
	}

}

func TestReferenceOwner(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, _ := seedTestRepo(t, repo)

	ref, err := repo.CreateReference("refs/heads/foo", commitId, true, nil, "")
	checkFatal(t, err)

	owner := ref.Owner()
	if owner == nil {
		t.Fatal("nil owner")
	}

	if owner.ptr != repo.ptr {
		t.Fatalf("bad ptr, expected %v have %v\n", repo.ptr, owner.ptr)
	}
}

func TestUtil(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, _ := seedTestRepo(t, repo)

	ref, err := repo.CreateReference("refs/heads/foo", commitId, true, nil, "")
	checkFatal(t, err)

	ref2, err := repo.DwimReference("foo")
	checkFatal(t, err)

	if ref.Cmp(ref2) != 0 {
		t.Fatalf("foo didn't dwim to the right thing")
	}

	if ref.Shorthand() != "foo" {
		t.Fatalf("refs/heads/foo has no foo shorthand")
	}

	hasLog, err := repo.HasLog("refs/heads/foo")
	checkFatal(t, err)
	if !hasLog {
		t.Fatalf("branches have logs by default")
	}
}

func compareStringList(t *testing.T, expected, actual []string) {
	for i, v := range expected {
		if actual[i] != v {
			t.Fatalf("Bad list")
		}
	}
}

func checkRefType(t *testing.T, ref *Reference, kind ReferenceType) {
	if ref.Type() == kind {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("Unable to get caller")
	}
	t.Fatalf("Wrong ref type at %v:%v; have %v, expected %v", file, line, ref.Type(), kind)
}
