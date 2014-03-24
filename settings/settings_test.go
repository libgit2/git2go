package settings

import (
	"testing"
	"runtime"
	"github.com/libgit2/git2go"
)

type pathPair struct {
	Level git.ConfigLevel
	Path  string
}

func TestSearchPath(t *testing.T) {
	paths := []pathPair{
		pathPair{git.ConfigLevelSystem, "/tmp/system"},
		pathPair{git.ConfigLevelGlobal, "/tmp/global"},
		pathPair{git.ConfigLevelXDG, "/tmp/xdg"},
	}

	for _, pair := range paths {
		err := SetSearchPath(pair.Level, pair.Path)
		checkFatal(t, err)

		actual, err := SearchPath(pair.Level)
		checkFatal(t, err)

		if pair.Path != actual {
			t.Fatal("Search paths don't match")
		}
	}
}

func TestMmapSizes(t *testing.T) {
	size := 42 * 1024

	err := SetMwindowSize(size)
	checkFatal(t, err)

	actual, err := MwindowSize()
	if size != actual {
		t.Fatal("Sizes don't match")
	}

	err = SetMwindowMappedLimit(size)
	checkFatal(t, err)

	actual, err = MwindowMappedLimit()
	if size != actual {
		t.Fatal("Sizes don't match")
	}
}

func checkFatal(t *testing.T, err error) {
	if err == nil {
		return
	}

	// The failure happens at wherever we were called, not here
	_, file, line, ok := runtime.Caller(1)
	if !ok {
		t.Fatal()
	}

	t.Fatalf("Fail at %v:%v; %v", file, line, err)
}
