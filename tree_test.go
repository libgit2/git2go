package git

import "testing"

func TestTreeEntryById(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	_, treeID := seedTestRepo(t, repo)

	tree, err := repo.LookupTree(treeID)
	checkFatal(t, err)

	id, err := NewOid("257cc5642cb1a054f08cc83f2d943e56fd3ebe99")
	checkFatal(t, err)

	entry := tree.EntryById(id)

	if entry == nil {
		t.Fatalf("entry id %v was not found", id)
	}
}
