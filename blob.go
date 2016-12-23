package git

/*
#include <git2.h>
#include <string.h>

int _go_git_writestream_write(git_writestream *stream, const char *buffer, size_t len);
void _go_git_writestream_free(git_writestream *stream);
*/
import "C"
import (
	"io"
	"reflect"
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

	var id C.git_oid
	var size C.size_t

	// Go 1.6 added some increased checking of passing pointer to
	// C, but its check depends on its expectations of waht we
	// pass to the C function, so unless we take the address of
	// its contents at the call site itself, it can fail when
	// 'data' is a slice of a slice.
	//
	// When we're given an empty slice, create a dummy one where 0
	// isn't out of bounds.
	if len(data) > 0 {
		size = C.size_t(len(data))
	} else {
		data = []byte{0}
		size = C.size_t(0)
	}

	ecode := C.git_blob_create_frombuffer(&id, repo.ptr, unsafe.Pointer(&data[0]), size)
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

func (repo *Repository) CreateFromStream(hintPath string) (*BlobWriteStream, error) {
	var chintPath *C.char = nil
	var stream *C.git_writestream

	if len(hintPath) > 0 {
		chintPath = C.CString(hintPath)
		defer C.free(unsafe.Pointer(chintPath))
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_blob_create_fromstream(&stream, repo.ptr, chintPath)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newBlobWriteStreamFromC(stream), nil
}

type BlobWriteStream struct {
	ptr *C.git_writestream
}

func newBlobWriteStreamFromC(ptr *C.git_writestream) *BlobWriteStream {
	stream := &BlobWriteStream{
		ptr: ptr,
	}

	runtime.SetFinalizer(stream, (*BlobWriteStream).Free)
	return stream
}

// Implement io.Writer
func (stream *BlobWriteStream) Write(p []byte) (int, error) {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&p))
	ptr := (*C.char)(unsafe.Pointer(header.Data))
	size := C.size_t(header.Len)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C._go_git_writestream_write(stream.ptr, ptr, size)
	if ecode < 0 {
		return 0, MakeGitError(ecode)
	}

	return len(p), nil
}

func (stream *BlobWriteStream) Free() {
	runtime.SetFinalizer(stream, nil)
	C._go_git_writestream_free(stream.ptr)
}

func (stream *BlobWriteStream) Commit() (*Oid, error) {
	oid := C.git_oid{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_blob_create_fromstream_commit(&oid, stream.ptr)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newOidFromC(&oid), nil
}
