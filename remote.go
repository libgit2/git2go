package git

/*
#include <git2.h>
#include <git2/errors.h>

extern void _go_git_setup_callbacks(git_remote_callbacks *callbacks);
extern git_remote_callbacks _go_git_remote_callbacks_init();
extern void _go_git_set_strarray_n(git_strarray *array, char *str, size_t n);
extern char *_go_git_get_strarray_n(git_strarray *array, size_t n);

*/
import "C"
import "unsafe"
import "runtime"

type TransferProgress struct {
	ptr *C.git_transfer_progress
}

type RemoteCompletion uint

const (
	RemoteCompletionDownload RemoteCompletion = C.GIT_REMOTE_COMPLETION_DOWNLOAD
	RemoteCompletionIndexing                  = C.GIT_REMOTE_COMPLETION_INDEXING
	RemoteCompletionError                     = C.GIT_REMOTE_COMPLETION_ERROR
)

type ProgressCallback func(str string) int
type CompletionCallback func(RemoteCompletion) int
type CredentialsCallback func(url string, username_from_url string, allowed_types CredType) (int, Cred)
type TransferProgressCallback func(stats TransferProgress) int
type UpdateTipsCallback func(refname string, a *Oid, b *Oid) int

type RemoteCallbacks struct {
	ProgressCallback
	CompletionCallback
	CredentialsCallback
	TransferProgressCallback
	UpdateTipsCallback
}

type Remote interface {
	Save() error
	Owner() Repository
	Name() string
	Url() string
	PushUrl() string

	SetUrl(url string) error
	SetPushUrl(url string) error

	AddFetch(refspec string) error
	GetFetchRefspecs() ([]string, error)
	SetFetchRefspecs(refspecs []string) error
	AddPush(refspec string) error
	GetPushRefspecs() ([]string, error)
	SetPushRefspecs(refspecs []string) error
	ClearRefspecs()
	RefspecCount() uint
}

type gitRemote struct {
	ptr *C.git_remote
}

func populateRemoteCallbacks(ptr *C.git_remote_callbacks, callbacks *RemoteCallbacks) {
	*ptr = C._go_git_remote_callbacks_init()
	if callbacks == nil {
		return
	}
	C._go_git_setup_callbacks(ptr)
	ptr.payload = unsafe.Pointer(callbacks)
}

//export progressCallback
func progressCallback(_str *C.char, _len C.int, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.ProgressCallback == nil {
		return 0
	}
	str := C.GoStringN(_str, _len)
	return callbacks.ProgressCallback(str)
}

//export completionCallback
func completionCallback(completion_type C.git_remote_completion_type, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.CompletionCallback == nil {
		return 0
	}
	return callbacks.CompletionCallback((RemoteCompletion)(completion_type))
}

//export credentialsCallback
func credentialsCallback(_cred **C.git_cred, _url *C.char, _username_from_url *C.char, allowed_types uint, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.CredentialsCallback == nil {
		return 0
	}
	url := C.GoString(_url)
	username_from_url := C.GoString(_username_from_url)
	ret, cred := callbacks.CredentialsCallback(url, username_from_url, (CredType)(allowed_types))
	if gcred, ok := cred.(gitCred); ok {
		*_cred = gcred.ptr
	}
	return ret
}

//export transferProgressCallback
func transferProgressCallback(stats *C.git_transfer_progress, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.TransferProgressCallback == nil {
		return 0
	}
	return callbacks.TransferProgressCallback(TransferProgress{stats})
}

//export updateTipsCallback
func updateTipsCallback(_refname *C.char, _a *C.git_oid, _b *C.git_oid, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.UpdateTipsCallback == nil {
		return 0
	}
	refname := C.GoString(_refname)
	a := newOidFromC(_a)
	b := newOidFromC(_b)
	return callbacks.UpdateTipsCallback(refname, a, b)
}

func (o TransferProgress) TotalObjects() uint {
	return uint(o.ptr.total_objects)
}

func (o TransferProgress) IndexedObjects() uint {
	return uint(o.ptr.indexed_objects)
}

func (o TransferProgress) ReceivedObjects() uint {
	return uint(o.ptr.received_objects)
}

func (o TransferProgress) LocalObjects() uint {
	return uint(o.ptr.local_objects)
}

func (o TransferProgress) TotalDeltas() uint {
	return uint(o.ptr.total_deltas)
}

func (o TransferProgress) ReceivedBytes() uint {
	return uint(o.ptr.received_bytes)
}

func RemoteIsValidName(name string) bool {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	if C.git_remote_is_valid_name(cname) == 1 {
		return true
	}
	return false
}

func freeRemote(o *gitRemote) {
	C.git_remote_free(o.ptr)
}

func CreateRemote(repo *Repository, name string, url string) (Remote, error) {
	remote := &gitRemote{}
	runtime.SetFinalizer(remote, freeRemote)

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_create(&remote.ptr, repo.ptr, cname, curl)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return remote, nil
}

func CreateRemoteWithFetchspec(repo *Repository, name string, url string, fetch string) (Remote, error) {
	remote := &gitRemote{}
	runtime.SetFinalizer(remote, freeRemote)

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	cfetch := C.CString(fetch)
	defer C.free(unsafe.Pointer(cfetch))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_create_with_fetchspec(&remote.ptr, repo.ptr, cname, curl, cfetch)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return remote, nil
}

func CreateRemoteInMemory(repo *Repository, fetch string, url string) (Remote, error) {
	remote := &gitRemote{}
	runtime.SetFinalizer(remote, freeRemote)

	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	cfetch := C.CString(fetch)
	defer C.free(unsafe.Pointer(cfetch))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_create_inmemory(&remote.ptr, repo.ptr, cfetch, curl)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return remote, nil
}

func LoadRemote(repo *Repository, name string) (Remote, error) {
	remote := &gitRemote{}
	runtime.SetFinalizer(remote, freeRemote)

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_load(&remote.ptr, repo.ptr, cname)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return remote, nil
}

func (o *gitRemote) Save() error {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_save(o.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *gitRemote) Owner() Repository {
	return Repository{C.git_remote_owner(o.ptr)}
}

func (o *gitRemote) Name() string {
	return C.GoString(C.git_remote_name(o.ptr))
}

func (o *gitRemote) Url() string {
	return C.GoString(C.git_remote_url(o.ptr))
}

func (o *gitRemote) PushUrl() string {
	return C.GoString(C.git_remote_pushurl(o.ptr))
}

func (o *gitRemote) SetUrl(url string) error {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_set_url(o.ptr, curl)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *gitRemote) SetPushUrl(url string) error {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_set_pushurl(o.ptr, curl)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *gitRemote) AddFetch(refspec string) error {
	crefspec := C.CString(refspec)
	defer C.free(unsafe.Pointer(crefspec))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_add_fetch(o.ptr, crefspec)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *gitRemote) GetFetchRefspecs() ([]string, error) {
	crefspecs := C.git_strarray{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_get_fetch_refspecs(&crefspecs, o.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	defer C.git_strarray_free(&crefspecs)
	refspecs := make([]string, crefspecs.count)

	for i := 0; i < int(crefspecs.count); i++ {
		refspecs[i] = C.GoString(C._go_git_get_strarray_n(&crefspecs, C.size_t(i)))
	}
	return refspecs, nil
}

func (o *gitRemote) SetFetchRefspecs(refspecs []string) error {
	crefspecs := C.git_strarray{}
	crefspecs.count = C.size_t(len(refspecs))
	crefspecs.strings = (**C.char)(C.malloc(C.size_t(unsafe.Sizeof(unsafe.Pointer(nil)) * uintptr(crefspecs.count))))
	for i, refspec := range refspecs {
		C._go_git_set_strarray_n(&crefspecs, C.CString(refspec), C.size_t(i))
	}
	defer C.git_strarray_free(&crefspecs)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_set_fetch_refspecs(o.ptr, &crefspecs)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *gitRemote) AddPush(refspec string) error {
	crefspec := C.CString(refspec)
	defer C.free(unsafe.Pointer(crefspec))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_add_push(o.ptr, crefspec)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *gitRemote) GetPushRefspecs() ([]string, error) {
	crefspecs := C.git_strarray{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_get_push_refspecs(&crefspecs, o.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	defer C.git_strarray_free(&crefspecs)
	refspecs := make([]string, crefspecs.count)

	for i := 0; i < int(crefspecs.count); i++ {
		refspecs[i] = C.GoString(C._go_git_get_strarray_n(&crefspecs, C.size_t(i)))
	}
	return refspecs, nil
}

func (o *gitRemote) SetPushRefspecs(refspecs []string) error {
	crefspecs := C.git_strarray{}
	crefspecs.count = C.size_t(len(refspecs))
	crefspecs.strings = (**C.char)(C.malloc(C.size_t(unsafe.Sizeof(unsafe.Pointer(nil)) * uintptr(crefspecs.count))))
	for i, refspec := range refspecs {
		C._go_git_set_strarray_n(&crefspecs, C.CString(refspec), C.size_t(i))
	}
	defer C.git_strarray_free(&crefspecs)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_set_push_refspecs(o.ptr, &crefspecs)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *gitRemote) ClearRefspecs() {
	C.git_remote_clear_refspecs(o.ptr)
}

func (o *gitRemote) RefspecCount() uint {
	return uint(C.git_remote_refspec_count(o.ptr))
}
