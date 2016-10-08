package git

import (
	"bytes"
	"testing"
)

type bufWrapper struct {
	buf     [64]byte
	pointer []byte
}

func doublePointerBytes() []byte {
	o := &bufWrapper{}
	o.pointer = o.buf[0:10]
	return o.pointer[0:1]
}

func TestCreateBlobFromBuffer(t *testing.T) {
	t.Parallel()
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	id, err := repo.CreateBlobFromBuffer(make([]byte, 0))
	checkFatal(t, err)

	if id.String() != "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391" {
		t.Fatal("Empty buffer did not deliver empty blob id")
	}

	for _, data := range []([]byte){[]byte("hello there"), doublePointerBytes()} {
		id, err = repo.CreateBlobFromBuffer(data)
		checkFatal(t, err)

		blob, err := repo.LookupBlob(id)
		checkFatal(t, err)
		if !bytes.Equal(blob.Contents(), data) {
			t.Fatal("Loaded bytes don't match original bytes:",
				blob.Contents(), "!=", data)
		}
	}
}
