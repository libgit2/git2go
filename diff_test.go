package git

import (
	"errors"
	"testing"
)

func TestDiffTreeToTree(t *testing.T) {
	repo := createTestRepo(t)
	defer repo.Free()
	//defer os.RemoveAll(repo.Workdir())

	_, originalTreeId := seedTestRepo(t, repo)
	originalTree, err := repo.LookupTree(originalTreeId)

	checkFatal(t, err)

	_, newTreeId := updateReadme(t, repo, "file changed\n")

	newTree, err := repo.LookupTree(newTreeId)
	checkFatal(t, err)

	diff, err := repo.DiffTreeToTree(originalTree, newTree)
	checkFatal(t, err)

	if diff == nil {
		t.Fatal("no diff returned")
	}

	files := make([]string, 0)

	err = diff.ForEachFile(func(file *DiffDelta) error {
		files = append(files, file.OldFile.Path)
		return nil
	})

	checkFatal(t, err)

	if len(files) != 1 {
		t.Fatal("Incorrect number of files in diff")
	}

	if files[0] != "README" {
		t.Fatal("File in diff was expected to be README")
	}

	errTest := errors.New("test error")

	err = diff.ForEachFile(func(file *DiffDelta) error {
		return errTest
	})

	if err != errTest {
		t.Fatal("Expected custom error to be returned")
	}

}
