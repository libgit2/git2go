package git

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"testing"
)

func TestOdbRead(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, _ = seedTestRepo(t, repo)
	odb, err := repo.Odb()
	if err != nil {
		t.Fatalf("Odb: %v", err)
	}
	data := []byte("hello")
	id, err := odb.Write(data, ObjectBlob)
	if err != nil {
		t.Fatalf("odb.Write: %v", err)
	}

	sz, typ, err := odb.ReadHeader(id)
	if err != nil {
		t.Fatalf("ReadHeader: %v", err)
	}

	if sz != uint64(len(data)) {
		t.Errorf("ReadHeader got size %d, want %d", sz, len(data))
	}
	if typ != ObjectBlob {
		t.Errorf("ReadHeader got object type %s", typ)
	}

	obj, err := odb.Read(id)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if !bytes.Equal(obj.Data(), data) {
		t.Errorf("Read got wrong data")
	}
	if sz := obj.Len(); sz != uint64(len(data)) {
		t.Errorf("Read got size %d, want %d", sz, len(data))
	}
	if typ := obj.Type(); typ != ObjectBlob {
		t.Errorf("Read got object type %s", typ)
	}
}

func TestOdbStream(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, _ = seedTestRepo(t, repo)

	odb, error := repo.Odb()
	checkFatal(t, error)

	str := "hello, world!"

	writeStream, error := odb.NewWriteStream(int64(len(str)), ObjectBlob)
	checkFatal(t, error)
	n, error := io.WriteString(writeStream, str)
	checkFatal(t, error)
	if n != len(str) {
		t.Fatalf("Bad write length %v != %v", n, len(str))
	}

	error = writeStream.Close()
	checkFatal(t, error)

	expectedId, error := NewOid("30f51a3fba5274d53522d0f19748456974647b4f")
	checkFatal(t, error)
	if writeStream.Id.Cmp(expectedId) != 0 {
		t.Fatal("Wrong data written")
	}

	readStream, error := odb.NewReadStream(&writeStream.Id)
	checkFatal(t, error)
	data, error := ioutil.ReadAll(readStream)
	if str != string(data) {
		t.Fatalf("Wrong data read %v != %v", str, string(data))
	}
}

func TestOdbHash(t *testing.T) {
	t.Parallel()
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

	for _, data := range [][]byte{[]byte(str), doublePointerBytes()} {
		oid, error := odb.Hash(data, ObjectCommit)
		checkFatal(t, error)

		coid, error := odb.Write(data, ObjectCommit)
		checkFatal(t, error)

		if oid.Cmp(coid) != 0 {
			t.Fatal("Hash and write Oids are different")
		}
	}
}

func TestOdbForeach(t *testing.T) {
	t.Parallel()
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
