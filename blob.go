package git

/*
#include <git2.h>
#include <string.h>

int _go_git_writestream_write(git_writestream *stream, const char *buffer, size_t len);
void _go_git_writestream_free(git_writestream *stream);
*/
import "C"
import (
	"reflect"
	"runtime"
	"unsafe"
)

type Blob struct {
	doNotCompare
	Object
	cast_ptr *C.git_blob
}

func (b *Blob) AsObject() *Object {
	return &b.Object
}

func (v *Blob) Size() int64 {
	ret := int64(C.git_blob_rawsize(v.cast_ptr))
	runtime.KeepAlive(v)
	return ret
}

func (v *Blob) Contents() []byte {
	size := C.int(C.git_blob_rawsize(v.cast_ptr))
	buffer := unsafe.Pointer(C.git_blob_rawcontent(v.cast_ptr))

	goBytes := C.GoBytes(buffer, size)
	runtime.KeepAlive(v)

	return goBytes
}

func (v *Blob) IsBinary() bool {
	ret := C.git_blob_is_binary(v.cast_ptr) == 1
	runtime.KeepAlive(v)
	return ret
}

func (repo *Repository) CreateBlobFromBuffer(data []byte) (*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var id C.git_oid
	var size C.size_t

	// Go 1.6 added some increased checking of passing pointer to
	// C, but its check depends on its expectations of what we
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

	ecode := C.git_blob_create_from_buffer(&id, repo.ptr, unsafe.Pointer(&data[0]), size)
	runtime.KeepAlive(repo)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	return newOidFromC(&id), nil
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

	ecode := C.git_blob_create_from_stream(&stream, repo.ptr, chintPath)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newBlobWriteStreamFromC(stream, repo), nil
}

type BlobWriteStream struct {
	doNotCompare
	ptr  *C.git_writestream
	repo *Repository
}

func newBlobWriteStreamFromC(ptr *C.git_writestream, repo *Repository) *BlobWriteStream {
	stream := &BlobWriteStream{
		ptr:  ptr,
		repo: repo,
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
	runtime.KeepAlive(stream)
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

	ecode := C.git_blob_create_from_stream_commit(&oid, stream.ptr)
	runtime.KeepAlive(stream)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newOidFromC(&oid), nil
}
