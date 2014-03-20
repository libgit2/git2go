package git

import (
	"testing"
)

func TestBranchIterator(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)

	i, err := repo.NewBranchIterator(BranchLocal)
	checkFatal(t, err)

	b, err := i.NextBranch()
	checkFatal(t, err)
	if name, _ := b.Branch.Name(); name != "master" {
		t.Fatalf("expected master")
	}
	if b.Type != BranchLocal {
		t.Fatalf("expected BranchLocal, not %v", t)
	}
	b, err = i.NextBranch()
	if err != ErrIterOver {
		t.Fatal("expected iterover")
	}

	// test channel iterator

	i, err = repo.NewBranchIterator(BranchLocal)
	checkFatal(t, err)

	list := make([]string, 0)
	for ref := range NameIteratorChannel(i) {
		list = append(list, ref)
	}

	if len(list) != 1 {
		t.Fatal("expected single match")
	}

	if list[0] != "master" {
		t.Fatal("expected master")
	}

}
