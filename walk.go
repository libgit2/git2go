package git

/*
#include <git2.h>
#include <git2/errors.h>
*/
import "C"

// RevWalk

type SortType uint
const (
	SortNone SortType = C.GIT_SORT_NONE
	SortTopological   = C.GIT_SORT_TOPOLOGICAL
	SortTime          = C.GIT_SORT_TIME
	SortReverse       = C.GIT_SORT_REVERSE
)

type RevWalk struct {
	ptr  *C.git_revwalk
	repo *Repository
}

func (v *RevWalk) Reset() {
	C.git_revwalk_reset(v.ptr)
}

func (v *RevWalk) Push(id *Oid) {
	C.git_revwalk_push(v.ptr, id.toC())
}

func (v *RevWalk) PushHead() (err error) {
	ecode := C.git_revwalk_push_head(v.ptr)
	if ecode < 0 {
		err = makeError(ecode)
	}

	return
}

func (v *RevWalk) Next(oid *Oid) (err error) {
	return makeError(C.git_revwalk_next(oid.toC(), v.ptr))
}

type RevWalkIterator func(commit *Commit) bool

func (v *RevWalk) Iterate(fun RevWalkIterator) (err error) {
	oid := new(Oid)
	for {
		err = v.Next(oid)
		if err == ErrIterOver {
			return nil
		}
		if err != nil {
			return err
		}

		commit, err := v.repo.LookupCommit(oid)
		if err != nil {
			return err
		}

		cont := fun(commit)
		if !cont {
			break
		}
	}

	return nil
}

func (v *RevWalk) Sorting(sm SortType) {
	C.git_revwalk_sorting(v.ptr, C.uint(sm))
}

func freeRevWalk(walk *RevWalk) {
	C.git_revwalk_free(walk.ptr)
}
