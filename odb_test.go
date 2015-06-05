package git

import (
	"errors"
	"io"
	"testing"
)

func TestOdbStream(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, _ = seedTestRepo(t, repo)

	odb, error := repo.Odb()
	checkFatal(t, error)

	str := "hello, world!"

	stream, error := odb.NewWriteStream(len(str), ObjectBlob)
	checkFatal(t, error)
	n, error := io.WriteString(stream, str)
	checkFatal(t, error)
	if n != len(str) {
		t.Fatalf("Bad write length %v != %v", n, len(str))
	}

	error = stream.Close()
	checkFatal(t, error)

	expectedId, error := NewOid("30f51a3fba5274d53522d0f19748456974647b4f")
	checkFatal(t, error)
	if stream.Id.Cmp(expectedId) != 0 {
		t.Fatal("Wrong data written")
	}
}

func TestOdbHash(t *testing.T) {

	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, _ = seedTestRepo(t, repo)

	odb, error := repo.Odb()
	checkFatal(t, error)

	str := `tree 115fcae49287c82eb55bb275cbbd4556fbed72b7
parent 66e1c476199ebcd3e304659992233132c5a52c6c
author John Doe <john@doe.com> 1390682018 +0000
committer John Doe <john@doe.com> 1390682018 +0000

Initial commit.`

	oid, error := odb.Hash([]byte(str), ObjectCommit)
	checkFatal(t, error)

	coid, error := odb.Write([]byte(str), ObjectCommit)
	checkFatal(t, error)

	if oid.Cmp(coid) != 0 {
		t.Fatal("Hash and write Oids are different")
	}
}

func TestOdbForeach(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, _ = seedTestRepo(t, repo)

	odb, err := repo.Odb()
	checkFatal(t, err)

	expect := 3
	count := 0
	err = odb.ForEach(func(id *Oid) error {
		count++
		return nil
	})

	checkFatal(t, err)
	if count != expect {
		t.Fatalf("Expected %v objects, got %v", expect, count)
	}

	expect = 1
	count = 0
	to_return := errors.New("not really an error")
	err = odb.ForEach(func(id *Oid) error {
		count++
		return to_return
	})

	if err != to_return {
		t.Fatalf("Odb.ForEach() did not return the expected error, got %v", err)
	}
}
