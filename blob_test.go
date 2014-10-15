package git

import (
	"os"
	"testing"
)

func TestCreateBlobFromBuffer(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	id, err := repo.CreateBlobFromBuffer(make([]byte, 0))
	checkFatal(t, err)

	if id.String() != "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391" {
		t.Fatal("Empty buffer did not deliver empty blob id")
	}
}
