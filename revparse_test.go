package git

import (
	"os"
	"testing"
)

func TestRevParseSingle(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())

	commitId, _ := seedTestRepo(t, repo)

	revSpec, err := repo.RevParse("HEAD")
	checkFatal(t, err)

	checkObject(t, revSpec.From(), commitId)
}

func checkObject(t *testing.T, obj Object, id *Oid) {
	if obj == nil {
		t.Fatalf("bad object")
	}

	if !obj.Id().Equal(id) {
		t.Fatalf("bad object, expected %s, got %s", id.String(), obj.Id().String())
	}
}
