package git

import (
	"strings"
	"testing"
)

func TestPatch(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	//defer os.RemoveAll(repo.Workdir())

	_, originalTreeId := seedTestRepo(t, repo)
	originalTree, err := repo.LookupTree(originalTreeId)

	checkFatal(t, err)

	_, newTreeId := updateReadme(t, repo, "file changed\n")

	newTree, err := repo.LookupTree(newTreeId)
	checkFatal(t, err)

	opts := &DiffOptions{
		OldPrefix: "a",
		NewPrefix: "b",
	}
	diff, err := repo.DiffTreeToTree(originalTree, newTree, opts)
	checkFatal(t, err)

	patch, err := diff.Patch(0)
	checkFatal(t, err)

	patchStr, err := patch.String()
	checkFatal(t, err)
	if strings.Index(patchStr, "diff --git a/README b/README\nindex 257cc56..820734a 100644\n--- a/README\n+++ b/README\n@@ -1 +1 @@\n-foo\n+file changed") == -1 {
		t.Fatalf("patch was bad")
	}
}
