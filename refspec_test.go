package git

import (
	"testing"
)

func TestRefspec(t *testing.T) {
	t.Parallel()

	const (
		input      = "+refs/heads/*:refs/remotes/origin/*"
		mainLocal  = "refs/heads/main"
		mainRemote = "refs/remotes/origin/main"
	)

	refspec, err := ParseRefspec(input, true)
	checkFatal(t, err)

	// Accessors

	s := refspec.String()
	if s != input {
		t.Errorf("expected string %q, got %q", input, s)
	}

	if d := refspec.Direction(); d != ConnectDirectionFetch {
		t.Errorf("expected fetch refspec, got direction %v", d)
	}

	if pat, expected := refspec.Src(), "refs/heads/*"; pat != expected {
		t.Errorf("expected refspec src %q, got %q", expected, pat)
	}

	if pat, expected := refspec.Dst(), "refs/remotes/origin/*"; pat != expected {
		t.Errorf("expected refspec dst %q, got %q", expected, pat)
	}

	if !refspec.Force() {
		t.Error("expected refspec force flag")
	}

	// SrcMatches

	if !refspec.SrcMatches(mainLocal) {
		t.Errorf("refspec source did not match %q", mainLocal)
	}

	if refspec.SrcMatches("refs/tags/v1.0") {
		t.Error("refspec source matched under refs/tags")
	}

	// DstMatches

	if !refspec.DstMatches(mainRemote) {
		t.Errorf("refspec destination did not match %q", mainRemote)
	}

	if refspec.DstMatches("refs/tags/v1.0") {
		t.Error("refspec destination matched under refs/tags")
	}

	// Transforms

	fromLocal, err := refspec.Transform(mainLocal)
	checkFatal(t, err)
	if fromLocal != mainRemote {
		t.Errorf("transform by refspec returned %s; expected %s", fromLocal, mainRemote)
	}

	fromRemote, err := refspec.Rtransform(mainRemote)
	checkFatal(t, err)
	if fromRemote != mainLocal {
		t.Errorf("rtransform by refspec returned %s; expected %s", fromRemote, mainLocal)
	}
}
