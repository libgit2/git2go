package git

/*
#include <string.h>

#include <git2.h>
#include <git2/sys/cred.h>

extern void _go_git_populate_remote_callbacks(git_remote_callbacks *callbacks);

*/
import "C"
import (
	"crypto/x509"
	"errors"
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
type CredentialsCallback func(url string, username_from_url string, allowed_types CredentialType) (*Credential, error)
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

type FetchPrune uint

const (
	// Use the setting from the configuration
	FetchPruneUnspecified FetchPrune = C.GIT_FETCH_PRUNE_UNSPECIFIED
	// Force pruning on
	FetchPruneOn FetchPrune = C.GIT_FETCH_PRUNE
	// Force pruning off
	FetchNoPrune FetchPrune = C.GIT_FETCH_NO_PRUNE
)

type DownloadTags uint

const (

	// Use the setting from the configuration.
	DownloadTagsUnspecified DownloadTags = C.GIT_REMOTE_DOWNLOAD_TAGS_UNSPECIFIED
	// Ask the server for tags pointing to objects we're already
	// downloading.
	DownloadTagsAuto DownloadTags = C.GIT_REMOTE_DOWNLOAD_TAGS_AUTO

	// Don't ask for any tags beyond the refspecs.
	DownloadTagsNone DownloadTags = C.GIT_REMOTE_DOWNLOAD_TAGS_NONE

	// Ask for the all the tags.
	DownloadTagsAll DownloadTags = C.GIT_REMOTE_DOWNLOAD_TAGS_ALL
)

type FetchOptions struct {
	// Callbacks to use for this fetch operation
	RemoteCallbacks RemoteCallbacks
	// Whether to perform a prune after the fetch
	Prune FetchPrune
	// Whether to write the results to FETCH_HEAD. Defaults to
	// on. Leave this default in order to behave like git.
	UpdateFetchhead bool

	// Determines how to behave regarding tags on the remote, such
	// as auto-downloading tags for objects we're downloading or
	// downloading all of them.
	//
	// The default is to auto-follow tags.
	DownloadTags DownloadTags

	// Headers are extra headers for the fetch operation.
	Headers []string

	// Proxy options to use for this fetch operation
	ProxyOptions ProxyOptions
}

type ProxyType uint

const (
	// Do not attempt to connect through a proxy
	//
	// If built against lbicurl, it itself may attempt to connect
	// to a proxy if the environment variables specify it.
	ProxyTypeNone ProxyType = C.GIT_PROXY_NONE

	// Try to auto-detect the proxy from the git configuration.
	ProxyTypeAuto ProxyType = C.GIT_PROXY_AUTO

	// Connect via the URL given in the options
	ProxyTypeSpecified ProxyType = C.GIT_PROXY_SPECIFIED
)

type ProxyOptions struct {
	// The type of proxy to use (or none)
	Type ProxyType

	// The proxy's URL
	Url string
}

type Remote struct {
	ptr       *C.git_remote
	callbacks RemoteCallbacks
	repo      *Repository
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
	HostkeyMD5    HostkeyKind = C.GIT_CERT_SSH_MD5
	HostkeySHA1   HostkeyKind = C.GIT_CERT_SSH_SHA1
	HostkeySHA256 HostkeyKind = C.GIT_CERT_SSH_SHA256
)

// Server host key information. A bitmask containing the available fields.
// Check for combinations of: HostkeyMD5, HostkeySHA1, HostkeySHA256.
type HostkeyCertificate struct {
	Kind       HostkeyKind
	HashMD5    [16]byte
	HashSHA1   [20]byte
	HashSHA256 [32]byte
}

type PushOptions struct {
	// Callbacks to use for this push operation
	RemoteCallbacks RemoteCallbacks

	PbParallelism uint

	// Headers are extra headers for the push operation.
	Headers []string
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

func untrackCalbacksPayload(callbacks *C.git_remote_callbacks) {
	if callbacks != nil && callbacks.payload != nil {
		pointerHandles.Untrack(callbacks.payload)
	}
}

func populateRemoteCallbacks(ptr *C.git_remote_callbacks, callbacks *RemoteCallbacks) {
	C.git_remote_init_callbacks(ptr, C.GIT_REMOTE_CALLBACKS_VERSION)
	if callbacks == nil {
		return
	}
	C._go_git_populate_remote_callbacks(ptr)
	ptr.payload = pointerHandles.Track(callbacks)
}

//export sidebandProgressCallback
func sidebandProgressCallback(errorMessage **C.char, _str *C.char, _len C.int, data unsafe.Pointer) C.int {
	callbacks := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.SidebandProgressCallback == nil {
		return C.int(ErrorCodeOK)
	}
	str := C.GoStringN(_str, _len)
	ret := callbacks.SidebandProgressCallback(str)
	if ret < 0 {
		return setCallbackError(errorMessage, errors.New(ErrorCode(ret).String()))
	}
	return C.int(ErrorCodeOK)
}

//export completionCallback
func completionCallback(errorMessage **C.char, completion_type C.git_remote_completion_type, data unsafe.Pointer) C.int {
	callbacks := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.CompletionCallback == nil {
		return C.int(ErrorCodeOK)
	}
	ret := callbacks.CompletionCallback(RemoteCompletion(completion_type))
	if ret < 0 {
		return setCallbackError(errorMessage, errors.New(ErrorCode(ret).String()))
	}
	return C.int(ErrorCodeOK)
}

//export credentialsCallback
func credentialsCallback(
	errorMessage **C.char,
	_cred **C.git_credential,
	_url *C.char,
	_username_from_url *C.char,
	allowed_types uint,
	data unsafe.Pointer,
) C.int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.CredentialsCallback == nil {
		return C.int(ErrorCodePassthrough)
	}
	url := C.GoString(_url)
	username_from_url := C.GoString(_username_from_url)
	cred, err := callbacks.CredentialsCallback(url, username_from_url, (CredentialType)(allowed_types))
	if err != nil {
		return setCallbackError(errorMessage, err)
	}
	if cred != nil {
		*_cred = cred.ptr

		// have transferred ownership to libgit, 'forget' the native pointer
		cred.ptr = nil
		runtime.SetFinalizer(cred, nil)
	}
	return C.int(ErrorCodeOK)
}

//export transferProgressCallback
func transferProgressCallback(errorMessage **C.char, stats *C.git_transfer_progress, data unsafe.Pointer) C.int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.TransferProgressCallback == nil {
		return C.int(ErrorCodeOK)
	}
	ret := callbacks.TransferProgressCallback(newTransferProgressFromC(stats))
	if ret < 0 {
		return setCallbackError(errorMessage, errors.New(ErrorCode(ret).String()))
	}
	return C.int(ErrorCodeOK)
}

//export updateTipsCallback
func updateTipsCallback(
	errorMessage **C.char,
	_refname *C.char,
	_a *C.git_oid,
	_b *C.git_oid,
	data unsafe.Pointer,
) C.int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.UpdateTipsCallback == nil {
		return C.int(ErrorCodeOK)
	}
	refname := C.GoString(_refname)
	a := newOidFromC(_a)
	b := newOidFromC(_b)
	ret := callbacks.UpdateTipsCallback(refname, a, b)
	if ret < 0 {
		return setCallbackError(errorMessage, errors.New(ErrorCode(ret).String()))
	}
	return C.int(ErrorCodeOK)
}

//export certificateCheckCallback
func certificateCheckCallback(
	errorMessage **C.char,
	_cert *C.git_cert,
	_valid C.int,
	_host *C.char,
	data unsafe.Pointer,
) C.int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	// if there's no callback set, we need to make sure we fail if the library didn't consider this cert valid
	if callbacks.CertificateCheckCallback == nil {
		if _valid == 0 {
			return C.int(ErrorCodeCertificate)
		}
		return C.int(ErrorCodeOK)
	}

	host := C.GoString(_host)
	valid := _valid != 0

	var cert Certificate
	if _cert.cert_type == C.GIT_CERT_X509 {
		cert.Kind = CertificateX509
		ccert := (*C.git_cert_x509)(unsafe.Pointer(_cert))
		x509_certs, err := x509.ParseCertificates(C.GoBytes(ccert.data, C.int(ccert.len)))
		if err != nil {
			return setCallbackError(errorMessage, err)
		}
		if len(x509_certs) < 1 {
			return setCallbackError(errorMessage, errors.New("empty certificate list"))
		}

		// we assume there's only one, which should hold true for any web server we want to talk to
		cert.X509 = x509_certs[0]
	} else if _cert.cert_type == C.GIT_CERT_HOSTKEY_LIBSSH2 {
		cert.Kind = CertificateHostkey
		ccert := (*C.git_cert_hostkey)(unsafe.Pointer(_cert))
		cert.Hostkey.Kind = HostkeyKind(ccert._type)
		C.memcpy(unsafe.Pointer(&cert.Hostkey.HashMD5[0]), unsafe.Pointer(&ccert.hash_md5[0]), C.size_t(len(cert.Hostkey.HashMD5)))
		C.memcpy(unsafe.Pointer(&cert.Hostkey.HashSHA1[0]), unsafe.Pointer(&ccert.hash_sha1[0]), C.size_t(len(cert.Hostkey.HashSHA1)))
		C.memcpy(unsafe.Pointer(&cert.Hostkey.HashSHA256[0]), unsafe.Pointer(&ccert.hash_sha256[0]), C.size_t(len(cert.Hostkey.HashSHA256)))
	} else {
		return setCallbackError(errorMessage, errors.New("unsupported certificate type"))
	}

	ret := callbacks.CertificateCheckCallback(&cert, valid, host)
	if ret < 0 {
		return setCallbackError(errorMessage, errors.New(ErrorCode(ret).String()))
	}
	return C.int(ErrorCodeOK)
}

//export packProgressCallback
func packProgressCallback(errorMessage **C.char, stage C.int, current, total C.uint, data unsafe.Pointer) C.int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.PackProgressCallback == nil {
		return C.int(ErrorCodeOK)
	}

	ret := callbacks.PackProgressCallback(int32(stage), uint32(current), uint32(total))
	if ret < 0 {
		return setCallbackError(errorMessage, errors.New(ErrorCode(ret).String()))
	}
	return C.int(ErrorCodeOK)
}

//export pushTransferProgressCallback
func pushTransferProgressCallback(errorMessage **C.char, current, total C.uint, bytes C.size_t, data unsafe.Pointer) C.int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.PushTransferProgressCallback == nil {
		return C.int(ErrorCodeOK)
	}

	ret := callbacks.PushTransferProgressCallback(uint32(current), uint32(total), uint(bytes))
	if ret < 0 {
		return setCallbackError(errorMessage, errors.New(ErrorCode(ret).String()))
	}
	return C.int(ErrorCodeOK)
}

//export pushUpdateReferenceCallback
func pushUpdateReferenceCallback(errorMessage **C.char, refname, status *C.char, data unsafe.Pointer) C.int {
	callbacks, _ := pointerHandles.Get(data).(*RemoteCallbacks)
	if callbacks.PushUpdateReferenceCallback == nil {
		return C.int(ErrorCodeOK)
	}

	ret := callbacks.PushUpdateReferenceCallback(C.GoString(refname), C.GoString(status))
	if ret < 0 {
		return setCallbackError(errorMessage, errors.New(ErrorCode(ret).String()))
	}
	return C.int(ErrorCodeOK)
}

func populateProxyOptions(ptr *C.git_proxy_options, opts *ProxyOptions) {
	C.git_proxy_options_init(ptr, C.GIT_PROXY_OPTIONS_VERSION)
	if opts == nil {
		return
	}

	ptr._type = C.git_proxy_t(opts.Type)
	ptr.url = C.CString(opts.Url)
}

func freeProxyOptions(ptr *C.git_proxy_options) {
	if ptr == nil {
		return
	}

	C.free(unsafe.Pointer(ptr.url))
}

// RemoteIsValidName returns whether the remote name is well-formed.
func RemoteIsValidName(name string) bool {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	return C.git_remote_is_valid_name(cname) == 1
}

// Free releases the resources of the Remote.
func (r *Remote) Free() {
	runtime.SetFinalizer(r, nil)
	C.git_remote_free(r.ptr)
	r.ptr = nil
	r.repo = nil
}

type RemoteCollection struct {
	repo *Repository
}

func (c *RemoteCollection) List() ([]string, error) {
	var r C.git_strarray

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_remote_list(&r, c.repo.ptr)
	runtime.KeepAlive(c.repo)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	defer C.git_strarray_free(&r)

	remotes := makeStringsFromCStrings(r.strings, int(r.count))
	return remotes, nil
}

func (c *RemoteCollection) Create(name string, url string) (*Remote, error) {
	remote := &Remote{repo: c.repo}

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_create(&remote.ptr, c.repo.ptr, cname, curl)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(remote, (*Remote).Free)
	return remote, nil
}

func (c *RemoteCollection) Delete(name string) error {
	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_delete(c.repo.ptr, cname)
	runtime.KeepAlive(c.repo)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (c *RemoteCollection) CreateWithFetchspec(name string, url string, fetch string) (*Remote, error) {
	remote := &Remote{repo: c.repo}

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	cfetch := C.CString(fetch)
	defer C.free(unsafe.Pointer(cfetch))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_create_with_fetchspec(&remote.ptr, c.repo.ptr, cname, curl, cfetch)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(remote, (*Remote).Free)
	return remote, nil
}

func (c *RemoteCollection) CreateAnonymous(url string) (*Remote, error) {
	remote := &Remote{repo: c.repo}

	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_create_anonymous(&remote.ptr, c.repo.ptr, curl)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(remote, (*Remote).Free)
	return remote, nil
}

func (c *RemoteCollection) Lookup(name string) (*Remote, error) {
	remote := &Remote{repo: c.repo}

	cname := C.CString(name)
	defer C.free(unsafe.Pointer(cname))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_lookup(&remote.ptr, c.repo.ptr, cname)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(remote, (*Remote).Free)
	return remote, nil
}

func (o *Remote) Name() string {
	s := C.git_remote_name(o.ptr)
	runtime.KeepAlive(o)
	return C.GoString(s)
}

func (o *Remote) Url() string {
	s := C.git_remote_url(o.ptr)
	runtime.KeepAlive(o)
	return C.GoString(s)
}

func (o *Remote) PushUrl() string {
	s := C.git_remote_pushurl(o.ptr)
	runtime.KeepAlive(o)
	return C.GoString(s)
}

func (c *RemoteCollection) Rename(remote, newname string) ([]string, error) {
	cproblems := C.git_strarray{}
	defer freeStrarray(&cproblems)
	cnewname := C.CString(newname)
	defer C.free(unsafe.Pointer(cnewname))
	cremote := C.CString(remote)
	defer C.free(unsafe.Pointer(cremote))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_rename(&cproblems, c.repo.ptr, cremote, cnewname)
	runtime.KeepAlive(c.repo)
	if ret < 0 {
		return []string{}, MakeGitError(ret)
	}

	problems := makeStringsFromCStrings(cproblems.strings, int(cproblems.count))
	return problems, nil
}

func (c *RemoteCollection) SetUrl(remote, url string) error {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	cremote := C.CString(remote)
	defer C.free(unsafe.Pointer(cremote))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_set_url(c.repo.ptr, cremote, curl)
	runtime.KeepAlive(c.repo)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (c *RemoteCollection) SetPushUrl(remote, url string) error {
	curl := C.CString(url)
	defer C.free(unsafe.Pointer(curl))
	cremote := C.CString(remote)
	defer C.free(unsafe.Pointer(cremote))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_set_pushurl(c.repo.ptr, cremote, curl)
	runtime.KeepAlive(c.repo)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (c *RemoteCollection) AddFetch(remote, refspec string) error {
	crefspec := C.CString(refspec)
	defer C.free(unsafe.Pointer(crefspec))
	cremote := C.CString(remote)
	defer C.free(unsafe.Pointer(cremote))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_add_fetch(c.repo.ptr, cremote, crefspec)
	runtime.KeepAlive(c.repo)
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
	runtime.KeepAlive(o)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	defer C.git_strarray_free(&crefspecs)

	refspecs := makeStringsFromCStrings(crefspecs.strings, int(crefspecs.count))
	return refspecs, nil
}

func (c *RemoteCollection) AddPush(remote, refspec string) error {
	crefspec := C.CString(refspec)
	defer C.free(unsafe.Pointer(crefspec))
	cremote := C.CString(remote)
	defer C.free(unsafe.Pointer(cremote))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_add_push(c.repo.ptr, cremote, crefspec)
	runtime.KeepAlive(c.repo)
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
	runtime.KeepAlive(o)

	refspecs := makeStringsFromCStrings(crefspecs.strings, int(crefspecs.count))
	return refspecs, nil
}

func (o *Remote) RefspecCount() uint {
	count := C.git_remote_refspec_count(o.ptr)
	runtime.KeepAlive(o)
	return uint(count)
}

func populateFetchOptions(options *C.git_fetch_options, opts *FetchOptions) {
	C.git_fetch_options_init(options, C.GIT_FETCH_OPTIONS_VERSION)
	if opts == nil {
		return
	}
	populateRemoteCallbacks(&options.callbacks, &opts.RemoteCallbacks)
	options.prune = C.git_fetch_prune_t(opts.Prune)
	options.update_fetchhead = cbool(opts.UpdateFetchhead)
	options.download_tags = C.git_remote_autotag_option_t(opts.DownloadTags)

	options.custom_headers = C.git_strarray{}
	options.custom_headers.count = C.size_t(len(opts.Headers))
	options.custom_headers.strings = makeCStringsFromStrings(opts.Headers)
	populateProxyOptions(&options.proxy_opts, &opts.ProxyOptions)
}

func populatePushOptions(options *C.git_push_options, opts *PushOptions) {
	C.git_push_options_init(options, C.GIT_PUSH_OPTIONS_VERSION)
	if opts == nil {
		return
	}

	options.pb_parallelism = C.uint(opts.PbParallelism)

	options.custom_headers = C.git_strarray{}
	options.custom_headers.count = C.size_t(len(opts.Headers))
	options.custom_headers.strings = makeCStringsFromStrings(opts.Headers)

	populateRemoteCallbacks(&options.callbacks, &opts.RemoteCallbacks)
}

// Fetch performs a fetch operation. refspecs specifies which refspecs
// to use for this fetch, use an empty list to use the refspecs from
// the configuration; msg specifies what to use for the reflog
// entries. Leave "" to use defaults.
func (o *Remote) Fetch(refspecs []string, opts *FetchOptions, msg string) error {
	var cmsg *C.char = nil
	if msg != "" {
		cmsg = C.CString(msg)
		defer C.free(unsafe.Pointer(cmsg))
	}

	crefspecs := C.git_strarray{}
	crefspecs.count = C.size_t(len(refspecs))
	crefspecs.strings = makeCStringsFromStrings(refspecs)
	defer freeStrarray(&crefspecs)

	coptions := (*C.git_fetch_options)(C.calloc(1, C.size_t(unsafe.Sizeof(C.git_fetch_options{}))))
	defer C.free(unsafe.Pointer(coptions))

	populateFetchOptions(coptions, opts)
	defer untrackCalbacksPayload(&coptions.callbacks)
	defer freeStrarray(&coptions.custom_headers)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_fetch(o.ptr, &crefspecs, coptions, cmsg)
	runtime.KeepAlive(o)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *Remote) ConnectFetch(callbacks *RemoteCallbacks, proxyOpts *ProxyOptions, headers []string) error {
	return o.Connect(ConnectDirectionFetch, callbacks, proxyOpts, headers)
}

func (o *Remote) ConnectPush(callbacks *RemoteCallbacks, proxyOpts *ProxyOptions, headers []string) error {
	return o.Connect(ConnectDirectionPush, callbacks, proxyOpts, headers)
}

// Connect opens a connection to a remote.
//
// The transport is selected based on the URL. The direction argument
// is due to a limitation of the git protocol (over TCP or SSH) which
// starts up a specific binary which can only do the one or the other.
//
// 'headers' are extra HTTP headers to use in this connection.
func (o *Remote) Connect(direction ConnectDirection, callbacks *RemoteCallbacks, proxyOpts *ProxyOptions, headers []string) error {
	var ccallbacks C.git_remote_callbacks
	populateRemoteCallbacks(&ccallbacks, callbacks)

	var cproxy C.git_proxy_options
	populateProxyOptions(&cproxy, proxyOpts)
	defer freeProxyOptions(&cproxy)

	cheaders := C.git_strarray{}
	cheaders.count = C.size_t(len(headers))
	cheaders.strings = makeCStringsFromStrings(headers)
	defer freeStrarray(&cheaders)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_connect(o.ptr, C.git_direction(direction), &ccallbacks, &cproxy, &cheaders)
	runtime.KeepAlive(o)
	if ret != 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *Remote) Disconnect() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	C.git_remote_disconnect(o.ptr)
	runtime.KeepAlive(o)
}

func (o *Remote) Ls(filterRefs ...string) ([]RemoteHead, error) {

	var refs **C.git_remote_head
	var length C.size_t

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_ls(&refs, &length, o.ptr)
	runtime.KeepAlive(o)
	if ret != 0 {
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

func (o *Remote) Push(refspecs []string, opts *PushOptions) error {
	crefspecs := C.git_strarray{}
	crefspecs.count = C.size_t(len(refspecs))
	crefspecs.strings = makeCStringsFromStrings(refspecs)
	defer freeStrarray(&crefspecs)

	coptions := (*C.git_push_options)(C.calloc(1, C.size_t(unsafe.Sizeof(C.git_push_options{}))))
	defer C.free(unsafe.Pointer(coptions))

	populatePushOptions(coptions, opts)
	defer untrackCalbacksPayload(&coptions.callbacks)
	defer freeStrarray(&coptions.custom_headers)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_push(o.ptr, &crefspecs, coptions)
	runtime.KeepAlive(o)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}

func (o *Remote) PruneRefs() bool {
	return C.git_remote_prune_refs(o.ptr) > 0
}

func (o *Remote) Prune(callbacks *RemoteCallbacks) error {
	var ccallbacks C.git_remote_callbacks
	populateRemoteCallbacks(&ccallbacks, callbacks)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_remote_prune(o.ptr, &ccallbacks)
	runtime.KeepAlive(o)
	if ret < 0 {
		return MakeGitError(ret)
	}
	return nil
}
