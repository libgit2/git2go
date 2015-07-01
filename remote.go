package git

/*
#include <git2.h>
#include <string.h>

extern void _go_git_setup_callbacks(git_remote_callbacks *callbacks);

*/
import "C"
import (
	"crypto/x509"
	"reflect"
	"runtime"
	"strings"
	"unsafe"
)

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
type ConnectDirection uint

const (
	RemoteCompletionDownload RemoteCompletion = C.GIT_REMOTE_COMPLETION_DOWNLOAD
	RemoteCompletionIndexing RemoteCompletion = C.GIT_REMOTE_COMPLETION_INDEXING
	RemoteCompletionError    RemoteCompletion = C.GIT_REMOTE_COMPLETION_ERROR

	ConnectDirectionFetch ConnectDirection = C.GIT_DIRECTION_FETCH
	ConnectDirectionPush  ConnectDirection = C.GIT_DIRECTION_PUSH
)

type TransportMessageCallback func(str string) ErrorCode
type CompletionCallback func(RemoteCompletion) ErrorCode
type CredentialsCallback func(url string, username_from_url string, allowed_types CredType) (ErrorCode, *Cred)
type TransferProgressCallback func(stats TransferProgress) ErrorCode
type UpdateTipsCallback func(refname string, a *Oid, b *Oid) ErrorCode
type CertificateCheckCallback func(cert *Certificate, valid bool, hostname string) ErrorCode
type PackbuilderProgressCallback func(stage int32, current, total uint32) ErrorCode
type PushTransferProgressCallback func(current, total uint32, bytes uint) ErrorCode
type PushUpdateReferenceCallback func(refname, status string) ErrorCode

type RemoteCallbacks struct {
	SidebandProgressCallback TransportMessageCallback
	CompletionCallback
	CredentialsCallback
	TransferProgressCallback
	UpdateTipsCallback
	CertificateCheckCallback
	PackProgressCallback PackbuilderProgressCallback
	PushTransferProgressCallback
	PushUpdateReferenceCallback
}

type Remote struct {
	ptr       *C.git_remote
	callbacks RemoteCallbacks
}

type CertificateKind uint

const (
	CertificateX509    CertificateKind = C.GIT_CERT_X509
	CertificateHostkey CertificateKind = C.GIT_CERT_HOSTKEY_LIBSSH2
)

// Certificate represents the two possible certificates which libgit2
// knows it might find. If Kind is CertficateX509 then the X509 field
// will be filled. If Kind is CertificateHostkey then the Hostkey
// field will be fille.d
type Certificate struct {
	Kind    CertificateKind
	X509    *x509.Certificate
	Hostkey HostkeyCertificate
}

type HostkeyKind uint

const (
	HostkeyMD5  HostkeyKind = C.GIT_CERT_SSH_MD5
	HostkeySHA1 HostkeyKind = C.GIT_CERT_SSH_SHA1
)

// Server host key information. If Kind is HostkeyMD5 the MD5 field
// will be filled. If Kind is HostkeySHA1, then HashSHA1 will be
// filled.
type HostkeyCertificate struct {
	Kind     HostkeyKind
	HashMD5  [16]byte
	HashSHA1 [20]byte
}

type PushOptions struct {
	PbParallelism uint
}

type RemoteHead struct {
	Id   *Oid
	Name string
}

func newRemoteHeadFromC(ptr *C.git_remote_head) RemoteHead {
	return RemoteHead{
		Id:   newOidFromC(&ptr.oid),
		Name: C.GoString(ptr.name),
	}
}

func populateRemoteCallbacks(ptr *C.git_remote_callbacks, callbacks *RemoteCallbacks) {
	C.git_remote_init_callbacks(ptr, C.GIT_REMOTE_CALLBACKS_VERSION)
	if callbacks == nil {
		return
	}
	C._go_git_setup_callbacks(ptr)
	ptr.payload = pointerHandles.Track(callbacks)
}

//export sidebandProgressCallback
func sidebandProgressCallback(_str *C.char, _len C.int, data unsafe.Pointer) int {
	callbacks := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.SidebandProgressCallback == nil {
		return 0
	}
	str := C.GoStringN(_str, _len)
	return int(callbacks.SidebandProgressCallback(str))
}

//export completionCallback
func completionCallback(completion_type C.git_remote_completion_type, data unsafe.Pointer) int {
	callbacks := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.CompletionCallback == nil {
		return 0
	}
	return int(callbacks.CompletionCallback(RemoteCompletion(completion_type)))
}

//export credentialsCallback
func credentialsCallback(_cred **C.git_cred, _url *C.char, _username_from_url *C.char, allowed_types uint, data unsafe.Pointer) int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.CredentialsCallback == nil {
		return 0
	}
	url := C.GoString(_url)
	username_from_url := C.GoString(_username_from_url)
	ret, cred := callbacks.CredentialsCallback(url, username_from_url, (CredType)(allowed_types))
	*_cred = cred.ptr
	return int(ret)
}

//export transferProgressCallback
func transferProgressCallback(stats *C.git_transfer_progress, data unsafe.Pointer) int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.TransferProgressCallback == nil {
		return 0
	}
	return int(callbacks.TransferProgressCallback(newTransferProgressFromC(stats)))
}

//export updateTipsCallback
func updateTipsCallback(_refname *C.char, _a *C.git_oid, _b *C.git_oid, data unsafe.Pointer) int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.UpdateTipsCallback == nil {
		return 0
	}
	refname := C.GoString(_refname)
	a := newOidFromC(_a)
	b := newOidFromC(_b)
	return int(callbacks.UpdateTipsCallback(refname, a, b))
}

//export certificateCheckCallback
func certificateCheckCallback(_cert *C.git_cert, _valid C.int, _host *C.char, data unsafe.Pointer) int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	// if there's no callback set, we need to make sure we fail if the library didn't consider this cert valid
	if callbacks.CertificateCheckCallback == nil {
		if _valid == 1 {
			return 0
		} else {
			return C.GIT_ECERTIFICATE
		}
	}
	host := C.GoString(_host)
	valid := _valid != 0

	var cert Certificate
	if _cert.cert_type == C.GIT_CERT_X509 {
		cert.Kind = CertificateX509
		ccert := (*C.git_cert_x509)(unsafe.Pointer(_cert))
		x509_certs, err := x509.ParseCertificates(C.GoBytes(ccert.data, C.int(ccert.len)))
		if err != nil {
			return C.GIT_EUSER
		}

		// we assume there's only one, which should hold true for any web server we want to talk to
		cert.X509 = x509_certs[0]
	} else if _cert.cert_type == C.GIT_CERT_HOSTKEY_LIBSSH2 {
		cert.Kind = CertificateHostkey
		ccert := (*C.git_cert_hostkey)(unsafe.Pointer(_cert))
		cert.Hostkey.Kind = HostkeyKind(ccert._type)
		C.memcpy(unsafe.Pointer(&cert.Hostkey.HashMD5[0]), unsafe.Pointer(&ccert.hash_md5[0]), C.size_t(len(cert.Hostkey.HashMD5)))
		C.memcpy(unsafe.Pointer(&cert.Hostkey.HashSHA1[0]), unsafe.Pointer(&ccert.hash_sha1[0]), C.size_t(len(cert.Hostkey.HashSHA1)))
	} else {
		cstr := C.CString("Unsupported certificate type")
		C.giterr_set_str(C.GITERR_NET, cstr)
		C.free(unsafe.Pointer(cstr))
		return -1 // we don't support anything else atm
	}

	return int(callbacks.CertificateCheckCallback(&cert, valid, host))
}

//export packProgressCallback
func packProgressCallback(stage C.int, current, total C.uint, data unsafe.Pointer) int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)

	if callbacks.PackProgressCallback == nil {
		return 0
	}

	return int(callbacks.PackProgressCallback(int32(stage), uint32(current), uint32(total)))
}

//export pushTransferProgressCallback
func pushTransferProgressCallback(current, total C.uint, bytes C.size_t, data unsafe.Pointer) int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.PushTransferProgressCallback == nil {
		return 0
	}

	return int(callbacks.PushTransferProgressCallback(uint32(current), uint32(total), uint(bytes)))
}

//export pushUpdateReferenceCallback
func pushUpdateReferenceCallback(refname, status *C.char, data unsafe.Pointer) int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)

	if callbacks.PushUpdateReferenceCallback == nil {
		return 0
	}

	return int(callbacks.PushUpdateReferenceCallback(C.GoString(refname), C.GoString(status)))
}

func RemoteIsValidName(name string) bool {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	if C.git_remote_is_valid_name(cname) == 1 {
		return true
	}
	return false
}

func (r *Remote) SetCallbacks(callbacks *RemoteCallbacks) error {
	r.callbacks = *callbacks

	var ccallbacks C.git_remote_callbacks
	populateRemoteCallbacks(&ccallbacks, &r.callbacks)

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

	callbacks := C.git_remote_get_callbacks(r.ptr)
	if callbacks != nil && callbacks.payload != nil {
		pointerHandles.Untrack(callbacks.payload)
	}

	C.git_remote_free(r.ptr)
}

func (repo *Repository) ListRemotes() ([]string, error) {
	var r C.git_strarray

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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

func (repo *Repository) DeleteRemote(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_delete(repo.ptr, cname)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
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

func (repo *Repository) LookupRemote(name string) (*Remote, error) {
	remote := &Remote{}

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_lookup(&remote.ptr, repo.ptr, cname)
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

func (o *Remote) SetUpdateFetchHead(val bool) {
	C.git_remote_set_update_fetchhead(o.ptr, cbool(val))
}

func (o *Remote) UpdateFetchHead() bool {
	return C.git_remote_update_fetchhead(o.ptr) > 0
}

// Fetch performs a fetch operation. refspecs specifies which refspecs
// to use for this fetch, use an empty list to use the refspecs from
// the configuration; sig and msg specify what to use for the reflog
// entries. Leave nil and "" to use defaults.
func (o *Remote) Fetch(refspecs []string, sig *Signature, msg string) error {

	var csig *C.git_signature = nil
	if sig != nil {
		csig, err := sig.toC()
		if err != nil {
			return err
		}
		defer C.git_signature_free(csig)
	}

	var cmsg *C.char = nil
	if msg != "" {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	crefspecs := C.git_strarray{}
	crefspecs.count = C.size_t(len(refspecs))
	crefspecs.strings = makeCStringsFromStrings(refspecs)
	defer freeStrarray(&crefspecs)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_fetch(o.ptr, &crefspecs, csig, cmsg)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *Remote) ConnectFetch() error {
	return o.Connect(ConnectDirectionFetch)
}

func (o *Remote) ConnectPush() error {
	return o.Connect(ConnectDirectionPush)
}

func (o *Remote) Connect(direction ConnectDirection) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_remote_connect(o.ptr, C.git_direction(direction)); ret != 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *Remote) Ls(filterRefs ...string) ([]RemoteHead, error) {

	var refs **C.git_remote_head
	var length C.size_t

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if ret := C.git_remote_ls(&refs, &length, o.ptr); ret != 0 {
		return nil, MakeGitError(ret)
	}

	size := int(length)

	if size == 0 {
		return make([]RemoteHead, 0), nil
	}

	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(refs)),
		Len:  size,
		Cap:  size,
	}

	goSlice := *(*[]*C.git_remote_head)(unsafe.Pointer(&hdr))

	var heads []RemoteHead

	for _, s := range goSlice {
		head := newRemoteHeadFromC(s)

		if len(filterRefs) > 0 {
			for _, r := range filterRefs {
				if strings.Contains(head.Name, r) {
					heads = append(heads, head)
					break
				}
			}
		} else {
			heads = append(heads, head)
		}
	}

	return heads, nil
}

func (o *Remote) Push(refspecs []string, opts *PushOptions, sig *Signature, msg string) error {
	var csig *C.git_signature = nil
	if sig != nil {
		csig, err := sig.toC()
		if err != nil {
			return err
		}
		defer C.git_signature_free(csig)
	}

	var cmsg *C.char
	if msg == "" {
		cmsg = nil
	} else {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	var copts C.git_push_options
	C.git_push_init_options(&copts, C.GIT_PUSH_OPTIONS_VERSION)
	if opts != nil {
		copts.pb_parallelism = C.uint(opts.PbParallelism)
	}

	crefspecs := C.git_strarray{}
	crefspecs.count = C.size_t(len(refspecs))
	crefspecs.strings = makeCStringsFromStrings(refspecs)
	defer freeStrarray(&crefspecs)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_push(o.ptr, &crefspecs, &copts, csig, cmsg)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *Remote) PruneRefs() bool {
	return C.git_remote_prune_refs(o.ptr) > 0
}

func (o *Remote) Prune() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_prune(o.ptr)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}
