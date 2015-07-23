package git

import (
	"path"
	"runtime"
	"strings"
	"testing"
)

func TestDescribeCommit(t *testing.T) {
	repo := createTestRepo(t)
	defer cleanupTestRepo(t, repo)

	describeOpts, err := DefaultDescribeOptions()
	checkFatal(t, err)

	formatOpts, err := DefaultDescribeFormatOptions()
	checkFatal(t, err)

	commitID, _ := seedTestRepo(t, repo)

	commit, err := repo.LookupCommit(commitID)
	checkFatal(t, err)

	// No annotated tags can be used to describe master
	_, err = commit.Describe(&describeOpts)
	checkDescribeNoRefsFound(t, err)

	// Fallback
	fallback := describeOpts
	fallback.ShowCommitOidAsFallback = true
	result, err := commit.Describe(&fallback)
	checkFatal(t, err)
	resultStr, err := result.Format(&formatOpts)
	checkFatal(t, err)
	compareStrings(t, "473bf77", resultStr)

	// Abbreviated
	abbreviated := formatOpts
	abbreviated.AbbreviatedSize = 2
	result, err = commit.Describe(&fallback)
	checkFatal(t, err)
	resultStr, err = result.Format(&abbreviated)
	checkFatal(t, err)
	compareStrings(t, "473b", resultStr)

	createTestTag(t, repo, commit)

	// Exact tag
	patternOpts := describeOpts
	patternOpts.Pattern = "v[0-9]*"
	result, err = commit.Describe(&patternOpts)
	checkFatal(t, err)
	resultStr, err = result.Format(&formatOpts)
	checkFatal(t, err)
	compareStrings(t, "v0.0.0", resultStr)

	// Pattern no match
	patternOpts.Pattern = "v[1-9]*"
	result, err = commit.Describe(&patternOpts)
	checkDescribeNoRefsFound(t, err)

	commitID, _ = updateReadme(t, repo, "update1")
	commit, err = repo.LookupCommit(commitID)
	checkFatal(t, err)

	// Tag-1
	result, err = commit.Describe(&describeOpts)
	checkFatal(t, err)
	resultStr, err = result.Format(&formatOpts)
	checkFatal(t, err)
	compareStrings(t, "v0.0.0-1-gd88ef8d", resultStr)

	// Strategy: All
	describeOpts.Strategy = DescribeAll
	result, err = commit.Describe(&describeOpts)
	checkFatal(t, err)
	resultStr, err = result.Format(&formatOpts)
	checkFatal(t, err)
	compareStrings(t, "heads/master", resultStr)

	repo.CreateBranch("hotfix", commit, false)

	// Workdir (branch)
	result, err = repo.DescribeWorkdir(&describeOpts)
	checkFatal(t, err)
	resultStr, err = result.Format(&formatOpts)
	checkFatal(t, err)
	compareStrings(t, "heads/hotfix", resultStr)
}

func checkDescribeNoRefsFound(t *testing.T, err error) {
	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatalf("Unable to get caller")
	}
	if err == nil || !strings.Contains(err.Error(), "No reference found, cannot describe anything") {
		t.Fatalf(
			"%s:%v: was expecting error 'No reference found, cannot describe anything', got %v",
			path.Base(file),
			line,
			err,
		)
	}
}
