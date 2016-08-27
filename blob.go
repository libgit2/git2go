package git

/*
#include <git2.h>
#include <string.h>

extern int _go_git_blob_create_fromchunks(git_oid *id,
	git_repository *repo,
	const char *hintpath,
	void *payload);

*/
import "C"
import (
	"io"
	"runtime"
	"unsafe"
)

type Blob struct {
	Object
	cast_ptr *C.git_blob
}

func (v *Blob) Size() int64 {
	return int64(C.git_blob_rawsize(v.cast_ptr))
}

func (v *Blob) Contents() []byte {
	size := C.int(C.git_blob_rawsize(v.cast_ptr))
	buffer := unsafe.Pointer(C.git_blob_rawcontent(v.cast_ptr))
	return C.GoBytes(buffer, size)
}

func (repo *Repository) CreateBlobFromBuffer(data []byte) (*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	// data may violate cgo 1.6's rules about not having pointers to
	// pointers passed to C, so make a copy where we know that won't be a
	// problem:
	safeData := make([]byte, len(data))
	copy(safeData, data)

	var id C.git_oid
	var ptr unsafe.Pointer

	if len(data) > 0 {
		ptr = unsafe.Pointer(&safeData[0])
	} else {
		ptr = unsafe.Pointer(nil)
	}

	ecode := C.git_blob_create_frombuffer(&id, repo.ptr, ptr, C.size_t(len(data)))
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	return newOidFromC(&id), nil
}

type BlobChunkCallback func(maxLen int) ([]byte, error)

type BlobCallbackData struct {
	Callback BlobChunkCallback
	Error    error
}

//export blobChunkCb
func blobChunkCb(buffer *C.char, maxLen C.size_t, handle unsafe.Pointer) int {
	payload := pointerHandles.Get(handle)
	data, ok := payload.(*BlobCallbackData)
	if !ok {
		panic("could not retrieve blob callback data")
	}

	goBuf, err := data.Callback(int(maxLen))
	if err == io.EOF {
		return 0
	} else if err != nil {
		data.Error = err
		return -1
	}
	C.memcpy(unsafe.Pointer(buffer), unsafe.Pointer(&goBuf[0]), C.size_t(len(goBuf)))
	return len(goBuf)
}

func (repo *Repository) CreateBlobFromChunks(hintPath string, callback BlobChunkCallback) (*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var chintPath *C.char = nil
	if len(hintPath) > 0 {
		chintPath = C.CString(hintPath)
		defer C.free(unsafe.Pointer(chintPath))
	}
	oid := C.git_oid{}

	payload := &BlobCallbackData{Callback: callback}
	handle := pointerHandles.Track(payload)
	defer pointerHandles.Untrack(handle)

	ecode := C._go_git_blob_create_fromchunks(&oid, repo.ptr, chintPath, handle)
	if payload.Error != nil {
		return nil, payload.Error
	}
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	return newOidFromC(&oid), nil
}
