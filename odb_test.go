package git

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
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

	odb, err := repo.Odb()
	checkFatal(t, err)

	str := "hello, world!"

	writeStream, err := odb.NewWriteStream(int64(len(str)), ObjectBlob)
	checkFatal(t, err)
	n, err := io.WriteString(writeStream, str)
	checkFatal(t, err)
	if n != len(str) {
		t.Fatalf("Bad write length %v != %v", n, len(str))
	}

	err = writeStream.Close()
	checkFatal(t, err)

	expectedId, err := NewOid("30f51a3fba5274d53522d0f19748456974647b4f")
	checkFatal(t, err)
	if writeStream.Id.Cmp(expectedId) != 0 {
		t.Fatal("Wrong data written")
	}

	readStream, err := odb.NewReadStream(&writeStream.Id)
	checkFatal(t, err)
	data, err := ioutil.ReadAll(readStream)
	if str != string(data) {
		t.Fatalf("Wrong data read %v != %v", str, string(data))
	}
}

func TestOdbHash(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, _ = seedTestRepo(t, repo)

	odb, err := repo.Odb()
	checkFatal(t, err)

	str := `tree 115fcae49287c82eb55bb275cbbd4556fbed72b7
parent 66e1c476199ebcd3e304659992233132c5a52c6c
author John Doe <john@doe.com> 1390682018 +0000
committer John Doe <john@doe.com> 1390682018 +0000

Initial commit.`

	for _, data := range [][]byte{[]byte(str), doublePointerBytes()} {
		oid, err := odb.Hash(data, ObjectCommit)
		checkFatal(t, err)

		coid, err := odb.Write(data, ObjectCommit)
		checkFatal(t, err)

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

func TestOdbWritepack(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, _ = seedTestRepo(t, repo)

	odb, err := repo.Odb()
	checkFatal(t, err)

	var finalStats TransferProgress
	writepack, err := odb.NewWritePack(func(stats TransferProgress) error {
		finalStats = stats
		return nil
	})
	checkFatal(t, err)
	defer writepack.Free()

	_, err = writepack.Write(outOfOrderPack)
	checkFatal(t, err)
	err = writepack.Commit()
	checkFatal(t, err)

	if finalStats.TotalObjects != 3 {
		t.Errorf("mismatched transferred objects, expected 3, got %v", finalStats.TotalObjects)
	}
	if finalStats.ReceivedObjects != 3 {
		t.Errorf("mismatched received objects, expected 3, got %v", finalStats.ReceivedObjects)
	}
	if finalStats.IndexedObjects != 3 {
		t.Errorf("mismatched indexed objects, expected 3, got %v", finalStats.IndexedObjects)
	}
}

func TestOdbBackendLoose(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, _ = seedTestRepo(t, repo)

	odb, err := repo.Odb()
	checkFatal(t, err)

	looseObjectsDir, err := ioutil.TempDir("", fmt.Sprintf("loose_objects_%s", path.Base(repo.Path())))
	checkFatal(t, err)
	defer os.RemoveAll(looseObjectsDir)

	looseObjectsBackend, err := NewOdbBackendLoose(looseObjectsDir, -1, false, 0, 0)
	checkFatal(t, err)
	if err := odb.AddBackend(looseObjectsBackend, 999); err != nil {
		looseObjectsBackend.Free()
		checkFatal(t, err)
	}

	str := "hello, world!"

	writeStream, err := odb.NewWriteStream(int64(len(str)), ObjectBlob)
	checkFatal(t, err)
	n, err := io.WriteString(writeStream, str)
	checkFatal(t, err)
	if n != len(str) {
		t.Fatalf("Bad write length %v != %v", n, len(str))
	}

	err = writeStream.Close()
	checkFatal(t, err)

	expectedId, err := NewOid("30f51a3fba5274d53522d0f19748456974647b4f")
	checkFatal(t, err)
	if !writeStream.Id.Equal(expectedId) {
		t.Fatalf("writeStream.id = %v; want %v", writeStream.Id, expectedId)
	}

	_, err = os.Stat(path.Join(looseObjectsDir, expectedId.String()[:2], expectedId.String()[2:]))
	checkFatal(t, err)
}
