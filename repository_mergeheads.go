package git

/*
#include <git2.h>
#include <git2/sys/repository.h>
int mergeheads_callback(struct git_oid *, void *);
*/
import "C"
import (
	"runtime"
	"unsafe"
)

// MergeHeads returns the *Oids of the Heads involved in this merge.
// This does not include the *Oid of the HEAD of the current repo.
func (v *Repository) MergeHeads() ([]*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	arr := []*Oid{}
	C.git_repository_mergehead_foreach(v.ptr, (*[0]byte)(C.mergeheads_callback),
		unsafe.Pointer(&arr))
	return arr, nil
}

//export mergeheads_callback
func mergeheads_callback(pOid *C.struct_git_oid, pArr unsafe.Pointer) C.int {
	arr := (*[]*Oid)(pArr)
	oid := newOidFromC(pOid)
	*arr = append(*arr, oid)
	return C.int(0)
}
