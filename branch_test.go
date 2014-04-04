package git

import (
	"testing"
)

func TestBranchIterator(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)

	i, err := repo.NewBranchIterator(BranchLocal)
	checkFatal(t, err)

	b, bt, err := i.Next()
	checkFatal(t, err)
	if name, _ := b.Name(); name != "master" {
		t.Fatalf("expected master")
	} else if bt != BranchLocal {
		t.Fatalf("expected BranchLocal, not %v", t)
	}
	b, bt, err = i.Next()
	if !IsErrorCode(err, ErrIterOver) {
		t.Fatal("expected iterover")
	}
}
