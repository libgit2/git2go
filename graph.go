package git

/*
#include <git2.h>
*/
import "C"
import (
	"runtime"
)

func (repo *Repository) DescendantOf(commit, ancestor *Oid) (bool, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_graph_descendant_of(repo.ptr, commit.toC(), ancestor.toC())
	runtime.KeepAlive(repo)
	runtime.KeepAlive(commit)
	runtime.KeepAlive(ancestor)
	if ret < 0 {
		return false, MakeGitError(ret)
	}

	return (ret > 0), nil
}

func (repo *Repository) AheadBehind(local, upstream *Oid) (ahead, behind int, err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var aheadT C.size_t
	var behindT C.size_t

	ret := C.git_graph_ahead_behind(&aheadT, &behindT, repo.ptr, local.toC(), upstream.toC())
	runtime.KeepAlive(repo)
	runtime.KeepAlive(local)
	runtime.KeepAlive(upstream)
	if ret < 0 {
		return 0, 0, MakeGitError(ret)
	}

	return int(aheadT), int(behindT), nil
}

// ReachableFromAny returns whether a commit is reachable from any of a list of
// commits by following parent edges.
func (repo *Repository) ReachableFromAny(commit *Oid, descendants []*Oid) (bool, error) {
	if len(descendants) == 0 {
		return false, nil
	}

	coids := make([]C.git_oid, len(descendants))
	for i := 0; i < len(descendants); i++ {
		coids[i] = *descendants[i].toC()
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_graph_reachable_from_any(repo.ptr, commit.toC(), &coids[0], C.size_t(len(descendants)))
	runtime.KeepAlive(repo)
	runtime.KeepAlive(commit)
	runtime.KeepAlive(coids)
	runtime.KeepAlive(descendants)
	if ret < 0 {
		return false, MakeGitError(ret)
	}

	return (ret > 0), nil
}
