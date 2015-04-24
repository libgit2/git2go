package git

import (
	"testing"
)

func TestBranchIterator(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

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

func TestBranchIteratorEach(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	seedTestRepo(t, repo)

	i, err := repo.NewBranchIterator(BranchLocal)
	checkFatal(t, err)

	var names []string
	f := func(b *Branch, t BranchType) error {
		name, err := b.Name()
		if err != nil {
			return err
		}

		names = append(names, name)
		return nil
	}

	err = i.ForEach(f)
	if err != nil && !IsErrorCode(err, ErrIterOver) {
		t.Fatal(err)
	}

	if len(names) != 1 {
		t.Fatalf("expect 1 branch, but it was %d\n", len(names))
	}

	if names[0] != "master" {
		t.Fatalf("expect branch master, but it was %s\n", names[0])
	}
}
