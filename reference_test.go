package git

import (
	"runtime"
	"sort"
	"testing"
	"time"
)

func TestRefModification(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, treeId := seedTestRepo(t, repo)

	_, err := repo.References.Create("refs/tags/tree", treeId, true, "testTreeTag")
	checkFatal(t, err)

	tag, err := repo.References.Lookup("refs/tags/tree")
	checkFatal(t, err)
	checkRefType(t, tag, ReferenceOid)

	ref, err := repo.References.Lookup("HEAD")
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

	_, err = tag.Rename("refs/tags/renamed", false, "")
	checkFatal(t, err)
	tag, err = repo.References.Lookup("refs/tags/renamed")
	checkFatal(t, err)
	checkRefType(t, ref, ReferenceOid)

}

func TestReferenceIterator(t *testing.T) {
	t.Parallel()
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

	_, err = repo.References.Create("refs/heads/one", commitId, true, "headOne")
	checkFatal(t, err)

	_, err = repo.References.Create("refs/heads/two", commitId, true, "headTwo")
	checkFatal(t, err)

	_, err = repo.References.Create("refs/heads/three", commitId, true, "headThree")
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
	if !IsErrorCode(err, ErrorCodeIterOver) {
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
	if !IsErrorCode(err, ErrorCodeIterOver) {
		t.Fatal("Iteration not over")
	}

	if count != 4 {
		t.Fatalf("Wrong number of references returned %v", count)
	}

}

func TestReferenceOwner(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, _ := seedTestRepo(t, repo)

	ref, err := repo.References.Create("refs/heads/foo", commitId, true, "")
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
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitId, _ := seedTestRepo(t, repo)

	ref, err := repo.References.Create("refs/heads/foo", commitId, true, "")
	checkFatal(t, err)

	ref2, err := repo.References.Dwim("foo")
	checkFatal(t, err)

	if ref.Cmp(ref2) != 0 {
		t.Fatalf("foo didn't dwim to the right thing")
	}

	if ref.Shorthand() != "foo" {
		t.Fatalf("refs/heads/foo has no foo shorthand")
	}

	hasLog, err := repo.References.HasLog("refs/heads/foo")
	checkFatal(t, err)
	if !hasLog {
		t.Fatalf("branches have logs by default")
	}
}

func TestIsNote(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	commitID, _ := seedTestRepo(t, repo)

	sig := &Signature{
		Name:  "Rand Om Hacker",
		Email: "random@hacker.com",
		When:  time.Now(),
	}

	refname, err := repo.Notes.DefaultRef()
	checkFatal(t, err)

	_, err = repo.Notes.Create(refname, sig, sig, commitID, "This is a note", false)
	checkFatal(t, err)

	ref, err := repo.References.Lookup(refname)
	checkFatal(t, err)

	if !ref.IsNote() {
		t.Fatalf("%s should be a note", ref.Name())
	}

	ref, err = repo.References.Create("refs/heads/foo", commitID, true, "")
	checkFatal(t, err)

	if ref.IsNote() {
		t.Fatalf("%s should not be a note", ref.Name())
	}
}

func TestReferenceNameIsValid(t *testing.T) {
	t.Parallel()
	valid, err := ReferenceNameIsValid("HEAD")
	checkFatal(t, err)
	if !valid {
		t.Errorf("HEAD should be a valid reference name")
	}
	valid, err = ReferenceNameIsValid("HEAD1")
	checkFatal(t, err)
	if valid {
		t.Errorf("HEAD1 should not be a valid reference name")
	}
}

func TestReferenceNormalizeName(t *testing.T) {
	t.Parallel()

	ref, err := ReferenceNormalizeName("refs/heads//master", ReferenceFormatNormal)
	checkFatal(t, err)

	if ref != "refs/heads/master" {
		t.Errorf("ReferenceNormalizeName(%q) = %q; want %q", "refs/heads//master", ref, "refs/heads/master")
	}

	ref, err = ReferenceNormalizeName("master", ReferenceFormatAllowOnelevel|ReferenceFormatRefspecShorthand)
	checkFatal(t, err)

	if ref != "master" {
		t.Errorf("ReferenceNormalizeName(%q) = %q; want %q", "master", ref, "master")
	}

	ref, err = ReferenceNormalizeName("foo^", ReferenceFormatNormal)
	if !IsErrorCode(err, ErrorCodeInvalidSpec) {
		t.Errorf("foo^ should be invalid")
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
