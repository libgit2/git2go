package git

/*
#include <git2.h>
#include <git2/errors.h>

extern void _go_git_setup_callbacks(git_remote_callbacks *callbacks);

*/
import "C"
import "unsafe"
import "runtime"

type TransferProgress struct {
	TotalObjects    uint
	IndexedObjects  uint
	ReceivedObjects uint
	LocalObjects    uint
	TotalDeltas     uint
	ReceivedBytes   uint
}

func newTransferProgressFromC(c *C.git_transfer_progress) TransferProgress {
	return TransferProgress{
		TotalObjects:    uint(c.total_objects),
		IndexedObjects:  uint(c.indexed_objects),
		ReceivedObjects: uint(c.received_objects),
		LocalObjects:    uint(c.local_objects),
		TotalDeltas:     uint(c.total_deltas),
		ReceivedBytes:   uint(c.received_bytes)}
}

type RemoteCompletion uint

const (
	RemoteCompletionDownload RemoteCompletion = C.GIT_REMOTE_COMPLETION_DOWNLOAD
	RemoteCompletionIndexing                  = C.GIT_REMOTE_COMPLETION_INDEXING
	RemoteCompletionError                     = C.GIT_REMOTE_COMPLETION_ERROR
)

type TransportMessageCallback func(str string) int
type CompletionCallback func(RemoteCompletion) int
type CredentialsCallback func(url string, username_from_url string, allowed_types CredType) (int, *Cred)
type TransferProgressCallback func(stats TransferProgress) int
type UpdateTipsCallback func(refname string, a *Oid, b *Oid) int

type RemoteCallbacks struct {
	SidebandProgressCallback TransportMessageCallback
	CompletionCallback
	CredentialsCallback
	TransferProgressCallback
	UpdateTipsCallback
}

type Remote struct {
	ptr *C.git_remote
}

func populateRemoteCallbacks(ptr *C.git_remote_callbacks, callbacks *RemoteCallbacks) {
	C.git_remote_init_callbacks(ptr, C.GIT_REMOTE_CALLBACKS_VERSION)
	if callbacks == nil {
		return
	}
	C._go_git_setup_callbacks(ptr)
	ptr.payload = unsafe.Pointer(callbacks)
}

//export sidebandProgressCallback
func sidebandProgressCallback(_str *C.char, _len C.int, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.SidebandProgressCallback == nil {
		return 0
	}
	str := C.GoStringN(_str, _len)
	return callbacks.SidebandProgressCallback(str)
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
	*_cred = cred.ptr
	return ret
}

//export transferProgressCallback
func transferProgressCallback(stats *C.git_transfer_progress, data unsafe.Pointer) int {
	callbacks := (*RemoteCallbacks)(data)
	if callbacks.TransferProgressCallback == nil {
		return 0
	}
	return callbacks.TransferProgressCallback(newTransferProgressFromC(stats))
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

func RemoteIsValidName(name string) bool {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	if C.git_remote_is_valid_name(cname) == 1 {
		return true
	}
	return false
}

func (r *Remote) SetCheckCert(check bool) {
	C.git_remote_check_cert(r.ptr, cbool(check))
}

func (r *Remote) SetCallbacks(callbacks *RemoteCallbacks) error {
	var ccallbacks C.git_remote_callbacks

	populateRemoteCallbacks(&ccallbacks, callbacks)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_remote_set_callbacks(r.ptr, &ccallbacks)
	if ecode < 0 {
		return MakeGitError(ecode)
	}

	return nil
}

func (r *Remote) Free() {
	runtime.SetFinalizer(r, nil)
	C.git_remote_free(r.ptr)
}

func (repo *Repository) ListRemotes() ([]string, error) {
	var r C.git_strarray
	ecode := C.git_remote_list(&r, repo.ptr)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	defer C.git_strarray_free(&r)

	remotes := makeStringsFromCStrings(r.strings, int(r.count))
	return remotes, nil
}

func (repo *Repository) CreateRemote(name string, url string) (*Remote, error) {
	remote := &Remote{}

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
	runtime.SetFinalizer(remote, (*Remote).Free)
	return remote, nil
}

func (repo *Repository) CreateRemoteWithFetchspec(name string, url string, fetch string) (*Remote, error) {
	remote := &Remote{}

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
	runtime.SetFinalizer(remote, (*Remote).Free)
	return remote, nil
}

func (repo *Repository) CreateAnonymousRemote(url, fetch string) (*Remote, error) {
	remote := &Remote{}

	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	cfetch := C.CString(fetch)
	defer C.free(unsafe.Pointer(cfetch))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_create_anonymous(&remote.ptr, repo.ptr, curl, cfetch)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(remote, (*Remote).Free)
	return remote, nil
}

func (repo *Repository) LoadRemote(name string) (*Remote, error) {
	remote := &Remote{}

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_load(&remote.ptr, repo.ptr, cname)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(remote, (*Remote).Free)
	return remote, nil
}

func (o *Remote) Save() error {

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_save(o.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *Remote) Owner() Repository {
	return Repository{C.git_remote_owner(o.ptr)}
}

func (o *Remote) Name() string {
	return C.GoString(C.git_remote_name(o.ptr))
}

func (o *Remote) Url() string {
	return C.GoString(C.git_remote_url(o.ptr))
}

func (o *Remote) PushUrl() string {
	return C.GoString(C.git_remote_pushurl(o.ptr))
}

func (o *Remote) SetUrl(url string) error {
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

func (o *Remote) SetPushUrl(url string) error {
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

func (o *Remote) AddFetch(refspec string) error {
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

func sptr(p uintptr) *C.char {
	return *(**C.char)(unsafe.Pointer(p))
}

func makeStringsFromCStrings(x **C.char, l int) []string {
	s := make([]string, l)
	i := 0
	for p := uintptr(unsafe.Pointer(x)); i < l; p += unsafe.Sizeof(uintptr(0)) {
		s[i] = C.GoString(sptr(p))
		i++
	}
	return s
}

func makeCStringsFromStrings(s []string) **C.char {
	l := len(s)
	x := (**C.char)(C.malloc(C.size_t(unsafe.Sizeof(unsafe.Pointer(nil)) * uintptr(l))))
	i := 0
	for p := uintptr(unsafe.Pointer(x)); i < l; p += unsafe.Sizeof(uintptr(0)) {
		*(**C.char)(unsafe.Pointer(p)) = C.CString(s[i])
		i++
	}
	return x
}

func freeStrarray(arr *C.git_strarray) {
	count := int(arr.count)
	size := unsafe.Sizeof(unsafe.Pointer(nil))

	i := 0
	for p := uintptr(unsafe.Pointer(arr.strings)); i < count; p += size {
		C.free(unsafe.Pointer(sptr(p)))
		i++
	}

	C.free(unsafe.Pointer(arr.strings))
}

func (o *Remote) FetchRefspecs() ([]string, error) {
	crefspecs := C.git_strarray{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_get_fetch_refspecs(&crefspecs, o.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	defer C.git_strarray_free(&crefspecs)

	refspecs := makeStringsFromCStrings(crefspecs.strings, int(crefspecs.count))
	return refspecs, nil
}

func (o *Remote) SetFetchRefspecs(refspecs []string) error {
	crefspecs := C.git_strarray{}
	crefspecs.count = C.size_t(len(refspecs))
	crefspecs.strings = makeCStringsFromStrings(refspecs)
	defer freeStrarray(&crefspecs)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_set_fetch_refspecs(o.ptr, &crefspecs)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *Remote) AddPush(refspec string) error {
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

func (o *Remote) PushRefspecs() ([]string, error) {
	crefspecs := C.git_strarray{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_get_push_refspecs(&crefspecs, o.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	defer C.git_strarray_free(&crefspecs)
	refspecs := makeStringsFromCStrings(crefspecs.strings, int(crefspecs.count))
	return refspecs, nil
}

func (o *Remote) SetPushRefspecs(refspecs []string) error {
	crefspecs := C.git_strarray{}
	crefspecs.count = C.size_t(len(refspecs))
	crefspecs.strings = makeCStringsFromStrings(refspecs)
	defer freeStrarray(&crefspecs)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_set_push_refspecs(o.ptr, &crefspecs)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *Remote) ClearRefspecs() {
	C.git_remote_clear_refspecs(o.ptr)
}

func (o *Remote) RefspecCount() uint {
	return uint(C.git_remote_refspec_count(o.ptr))
}

func (o *Remote) Fetch(sig *Signature, msg string) error {

	var csig *C.git_signature = nil
	if sig != nil {
		csig = sig.toC()
		defer C.free(unsafe.Pointer(csig))
	}

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}
	ret := C.git_remote_fetch(o.ptr, csig, cmsg)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}
