package git

import (
	"testing"
)

func TestMergeWithSelf(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)

	master, err := repo.LookupReference("refs/heads/master")
	checkFatal(t, err)

	mergeHead, err := repo.MergeHeadFromRef(master)
	checkFatal(t, err)

	options, _ := DefaultMergeOptions()
	mergeHeads := make([]*MergeHead, 1)
	mergeHeads[0] = mergeHead
	err = repo.Merge(mergeHeads, &options, nil)
	checkFatal(t, err)
}
