package git

import (
	"testing"
)

func Test_Merge_With_Self(t *testing.T) {

	repo := createTestRepo(t)
	seedTestRepo(t, repo)

	master, err := repo.LookupReference("refs/heads/master")
	checkFatal(t, err)

	mergeHead, err := repo.MergeHeadFromRef(master)
	checkFatal(t, err)

	options := DefaultMergeOptions()
	mergeHeads := make([]*MergeHead, 1)
	mergeHeads[0] = mergeHead
	results, err := repo.Merge(mergeHeads, options)
	checkFatal(t, err)

	if !results.IsUpToDate() {
		t.Fatal("Expected up to date")
	}
}
