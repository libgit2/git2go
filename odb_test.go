package git

import (
	"io"
	"os"
	"testing"
)

func TestOdbStream(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
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

	expectedId, error := NewOidFromString("30f51a3fba5274d53522d0f19748456974647b4f")
	checkFatal(t, error)
	if stream.Id.Cmp(expectedId) != 0 {
		t.Fatal("Wrong data written")
	}
}