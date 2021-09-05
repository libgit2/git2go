package git

/*
#include <git2.h>

extern const git_oid * git_indexer_hash(const git_indexer *idx);
extern int git_indexer_append(git_indexer *idx, const void *data, size_t size, git_transfer_progress *stats);
extern int git_indexer_commit(git_indexer *idx, git_transfer_progress *stats);
extern int _go_git_indexer_new(git_indexer **out, const char *path, unsigned int mode, git_odb *odb, void *progress_cb_payload);
extern void git_indexer_free(git_indexer *idx);
*/
import "C"
import (
	"reflect"
	"runtime"
	"unsafe"
)

// Indexer can post-process packfiles and create an .idx file for efficient
// lookup.
type Indexer struct {
	doNotCompare
	ptr        *C.git_indexer
	stats      C.git_transfer_progress
	ccallbacks C.git_remote_callbacks
}

// NewIndexer creates a new indexer instance.
func NewIndexer(packfilePath string, odb *Odb, callback TransferProgressCallback) (indexer *Indexer, err error) {
	var odbPtr *C.git_odb = nil
	if odb != nil {
		odbPtr = odb.ptr
	}

	indexer = new(Indexer)
	populateRemoteCallbacks(&indexer.ccallbacks, &RemoteCallbacks{TransferProgressCallback: callback}, nil)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cstr := C.CString(packfilePath)
	defer C.free(unsafe.Pointer(cstr))

	ret := C._go_git_indexer_new(&indexer.ptr, cstr, 0, odbPtr, indexer.ccallbacks.payload)
	runtime.KeepAlive(odb)
	if ret < 0 {
		untrackCallbacksPayload(&indexer.ccallbacks)
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(indexer, (*Indexer).Free)
	return indexer, nil
}

// Write adds data to the indexer.
func (indexer *Indexer) Write(data []byte) (int, error) {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ptr := unsafe.Pointer(header.Data)
	size := C.size_t(header.Len)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_indexer_append(indexer.ptr, ptr, size, &indexer.stats)
	runtime.KeepAlive(indexer)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}

	return len(data), nil
}

// Commit finalizes the pack and index. It resolves any pending deltas and
// writes out the index file.
//
// It also returns the packfile's hash. A packfile's name is derived from the
// sorted hashing of all object names.
func (indexer *Indexer) Commit() (*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_indexer_commit(indexer.ptr, &indexer.stats)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	id := newOidFromC(C.git_indexer_hash(indexer.ptr))
	runtime.KeepAlive(indexer)
	return id, nil
}

// Free frees the indexer and its resources.
func (indexer *Indexer) Free() {
	untrackCallbacksPayload(&indexer.ccallbacks)
	runtime.SetFinalizer(indexer, nil)
	C.git_indexer_free(indexer.ptr)
}
