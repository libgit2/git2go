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

//export indexerTransferProgress
func indexerTransferProgress(_stats *C.git_transfer_progress, ptr unsafe.Pointer) C.int {
	callback, ok := pointerHandles.Get(ptr).(TransferProgressCallback)
	if !ok {
		return 0
	}
	return C.int(callback(newTransferProgressFromC(_stats)))
}

type Indexer struct {
	ptr      *C.git_indexer
	stats    C.git_transfer_progress
	callback unsafe.Pointer
}

func NewIndexer(packfilePath string, odb *Odb, callback TransferProgressCallback) (indexer *Indexer, err error) {
	indexer = new(Indexer)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var odbPtr *C.git_odb = nil
	if odb != nil {
		odbPtr = odb.ptr
	}

	if callback != nil {
		indexer.callback = pointerHandles.Track(callback)
	}

	ret := C._go_git_indexer_new(&indexer.ptr, C.CString(packfilePath), 0, odbPtr, indexer.callback)
	if ret < 0 {
		if indexer.callback != nil {
			pointerHandles.Untrack(indexer.callback)
		}
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(indexer, (*Indexer).Free)
	return indexer, nil
}

func (indexer *Indexer) Write(data []byte) (int, error) {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ptr := unsafe.Pointer(header.Data)
	size := C.size_t(header.Len)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_indexer_append(indexer.ptr, ptr, size, &indexer.stats)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}

	return len(data), nil
}

func (indexer *Indexer) Commit() (*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_indexer_commit(indexer.ptr, &indexer.stats)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newOidFromC(C.git_indexer_hash(indexer.ptr)), nil
}

func (indexer *Indexer) Free() {
	if indexer.callback != nil {
		pointerHandles.Untrack(indexer.callback)
	}
	C.git_indexer_free(indexer.ptr)
}
