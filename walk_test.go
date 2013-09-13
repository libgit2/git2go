package git

import (
	"os"
	"testing"
)

func TestWalk(t *testing.T) {
	repo := createTestRepo(t)
	defer os.RemoveAll(repo.Workdir())
	commitId, _ := seedTestRepo(t, repo)

	walk, err := repo.NewRevWalk()
	checkFatal(t, err)
	walk.Push(commitId)
	walk.Sorting(SortTime | SortReverse)
	var id Oid
	err = walk.Next(&id)
	checkFatal(t, err)
	if id.Cmp(commitId) != 0 {
		t.Fatal("Bad id returned")
	}
}