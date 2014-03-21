package git

import (
	"testing"
)

func TestDiffTreeToTree(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	defer os.RemoveAll(repo.Workdir())

	_, originalTreeId := seedTestRepo(t, repo)
	originalTree, err := repo.LookupTree(originalTreeId)

	checkFatal(t, err)
	updateReadme(t, repo, "file changed\n")

	_, newTreeId := seedTestRepo(t, repo)
	newTree, err := repo.LookupTree(newTreeId)
	checkFatal(t, err)

	diff, err := repo.DiffTreeToTree(originalTreeId, newTreeId)
	checkFatal(t, err)

	files := make([]string, 0)

	err := diff.ForEachFile(func(file *DiffFile) error {
		files = append(files, file.Path)
		return nil
	})

	checkFatal(t, err)

	if len(files) != 0 {
		t.Fatal("Incorrect number of files in diff")
	}

	if files[0] != "README" {
		t.Fatal("File in diff was expected to be README")
	}
}
