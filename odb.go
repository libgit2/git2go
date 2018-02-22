package git

/*
#include <git2.h>

extern int _go_git_odb_foreach(git_odb *db, void *payload);
extern void _go_git_odb_backend_free(git_odb_backend *backend);
*/
import "C"
import (
	"reflect"
	"runtime"
	"unsafe"
)

type Odb struct {
	ptr *C.git_odb
}

type OdbBackend struct {
	ptr *C.git_odb_backend
}

func NewOdb() (odb *Odb, err error) {
	odb = new(Odb)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_new(&odb.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(odb, (*Odb).Free)
	return odb, nil
}

func NewOdbBackendFromC(ptr unsafe.Pointer) (backend *OdbBackend) {
	backend = &OdbBackend{(*C.git_odb_backend)(ptr)}
	return backend
}

func (v *Odb) AddBackend(backend *OdbBackend, priority int) (err error) {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_add_backend(v.ptr, backend.ptr, C.int(priority))
	runtime.KeepAlive(v)
	if ret < 0 {
		backend.Free()
		return MakeGitError(ret)
	}
	return nil
}

func (v *Odb) ReadHeader(oid *Oid) (uint64, ObjectType, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var sz C.size_t
	var cotype C.git_otype

	ret := C.git_odb_read_header(&sz, &cotype, v.ptr, oid.toC())
	runtime.KeepAlive(v)
	if ret < 0 {
		return 0, C.GIT_OBJ_BAD, MakeGitError(ret)
	}

	return uint64(sz), ObjectType(cotype), nil
}

func (v *Odb) Exists(oid *Oid) bool {
	ret := C.git_odb_exists(v.ptr, oid.toC())
	runtime.KeepAlive(v)
	runtime.KeepAlive(oid)
	return ret != 0
}

func (v *Odb) Write(data []byte, otype ObjectType) (oid *Oid, err error) {
	oid = new(Oid)
	var cptr unsafe.Pointer
	if len(data) > 0 {
		cptr = unsafe.Pointer(&data[0])
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_write(oid.toC(), v.ptr, cptr, C.size_t(len(data)), C.git_otype(otype))
	runtime.KeepAlive(v)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return oid, nil
}

func (v *Odb) Read(oid *Oid) (obj *OdbObject, err error) {
	obj = new(OdbObject)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_read(&obj.ptr, v.ptr, oid.toC())
	runtime.KeepAlive(v)
	runtime.KeepAlive(oid)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(obj, (*OdbObject).Free)
	return obj, nil
}

type OdbForEachCallback func(id *Oid) error

type foreachData struct {
	callback OdbForEachCallback
	err      error
}

//export odbForEachCb
func odbForEachCb(id *C.git_oid, handle unsafe.Pointer) int {
	data, ok := pointerHandles.Get(handle).(*foreachData)

	if !ok {
		panic("could not retrieve handle")
	}

	err := data.callback(newOidFromC(id))
	if err != nil {
		data.err = err
		return C.GIT_EUSER
	}

	return 0
}

func (v *Odb) ForEach(callback OdbForEachCallback) error {
	data := foreachData{
		callback: callback,
		err:      nil,
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	handle := pointerHandles.Track(&data)
	defer pointerHandles.Untrack(handle)

	ret := C._go_git_odb_foreach(v.ptr, handle)
	runtime.KeepAlive(v)
	if ret == C.GIT_EUSER {
		return data.err
	} else if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// Hash determines the object-ID (sha1) of a data buffer.
func (v *Odb) Hash(data []byte, otype ObjectType) (oid *Oid, err error) {
	oid = new(Oid)
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ptr := unsafe.Pointer(header.Data)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_hash(oid.toC(), ptr, C.size_t(header.Len), C.git_otype(otype))
	runtime.KeepAlive(data)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return oid, nil
}

// NewReadStream opens a read stream from the ODB. Reading from it will give you the
// contents of the object.
func (v *Odb) NewReadStream(id *Oid) (*OdbReadStream, error) {
	stream := new(OdbReadStream)
	var ctype C.git_otype
	var csize C.size_t

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_open_rstream(&stream.ptr, &csize, &ctype, v.ptr, id.toC())
	runtime.KeepAlive(v)
	runtime.KeepAlive(id)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	stream.Size = uint64(csize)
	stream.Type = ObjectType(ctype)
	runtime.SetFinalizer(stream, (*OdbReadStream).Free)
	return stream, nil
}

// NewWriteStream opens a write stream to the ODB, which allows you to
// create a new object in the database. The size and type must be
// known in advance
func (v *Odb) NewWriteStream(size int64, otype ObjectType) (*OdbWriteStream, error) {
	stream := new(OdbWriteStream)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_open_wstream(&stream.ptr, v.ptr, C.git_off_t(size), C.git_otype(otype))
	runtime.KeepAlive(v)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	runtime.SetFinalizer(stream, (*OdbWriteStream).Free)
	return stream, nil
}

func (v *OdbBackend) Free() {
	C._go_git_odb_backend_free(v.ptr)
}

type OdbObject struct {
	ptr *C.git_odb_object
}

func (v *OdbObject) Free() {
	runtime.SetFinalizer(v, nil)
	C.git_odb_object_free(v.ptr)
}

func (object *OdbObject) Id() (oid *Oid) {
	ret := newOidFromC(C.git_odb_object_id(object.ptr))
	runtime.KeepAlive(object)
	return ret
}

func (object *OdbObject) Len() (len uint64) {
	ret := uint64(C.git_odb_object_size(object.ptr))
	runtime.KeepAlive(object)
	return ret
}

func (object *OdbObject) Type() ObjectType {
	ret := ObjectType(C.git_odb_object_type(object.ptr))
	runtime.KeepAlive(object)
	return ret
}

// Data returns a slice pointing to the unmanaged object memory. You must make
// sure the object is referenced for at least as long as the slice is used.
func (object *OdbObject) Data() (data []byte) {
	var c_blob unsafe.Pointer = C.git_odb_object_data(object.ptr)
	var blob []byte

	len := int(C.git_odb_object_size(object.ptr))

	sliceHeader := (*reflect.SliceHeader)((unsafe.Pointer(&blob)))
	sliceHeader.Cap = len
	sliceHeader.Len = len
	sliceHeader.Data = uintptr(c_blob)

	return blob
}

type OdbReadStream struct {
	ptr  *C.git_odb_stream
	Size uint64
	Type ObjectType
}

// Read reads from the stream
func (stream *OdbReadStream) Read(data []byte) (int, error) {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ptr := (*C.char)(unsafe.Pointer(header.Data))
	size := C.size_t(header.Cap)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_stream_read(stream.ptr, ptr, size)
	runtime.KeepAlive(stream)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}

	header.Len = int(ret)

	return len(data), nil
}

// Close is a dummy function in order to implement the Closer and
// ReadCloser interfaces
func (stream *OdbReadStream) Close() error {
	return nil
}

func (stream *OdbReadStream) Free() {
	runtime.SetFinalizer(stream, nil)
	C.git_odb_stream_free(stream.ptr)
}

type OdbWriteStream struct {
	ptr *C.git_odb_stream
	Id  Oid
}

// Write writes to the stream
func (stream *OdbWriteStream) Write(data []byte) (int, error) {
	header := (*reflect.SliceHeader)(unsafe.Pointer(&data))
	ptr := (*C.char)(unsafe.Pointer(header.Data))
	size := C.size_t(header.Len)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_stream_write(stream.ptr, ptr, size)
	runtime.KeepAlive(stream)
	if ret < 0 {
		return 0, MakeGitError(ret)
	}

	return len(data), nil
}

// Close signals that all the data has been written and stores the
// resulting object id in the stream's Id field.
func (stream *OdbWriteStream) Close() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_odb_stream_finalize_write(stream.Id.toC(), stream.ptr)
	runtime.KeepAlive(stream)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (stream *OdbWriteStream) Free() {
	runtime.SetFinalizer(stream, nil)
	C.git_odb_stream_free(stream.ptr)
}
