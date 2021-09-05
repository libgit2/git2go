package git

/*
#include <git2.h>
*/
import "C"
import (
	"unsafe"
)

// The constants, functions, and types in this files are slated for deprecation
// in the next major version.

// blob.go

// Deprecated: BlobChunkCallback is not used.
type BlobChunkCallback func(maxLen int) ([]byte, error)

// Deprecated: BlobCallbackData is not used.
type BlobCallbackData struct {
	Callback BlobChunkCallback
	Error    error
}

// checkout.go

// Deprecated: CheckoutOpts is a deprecated alias of CheckoutOptions.
type CheckoutOpts = CheckoutOptions

// credentials.go

// Deprecated: CredType is a deprecated alias of CredentialType
type CredType = CredentialType

const (
	CredTypeUserpassPlaintext = CredentialTypeUserpassPlaintext
	CredTypeSshKey            = CredentialTypeSSHKey
	CredTypeSshCustom         = CredentialTypeSSHCustom
	CredTypeDefault           = CredentialTypeDefault
)

// Deprecated: Cred is a deprecated alias of Credential
type Cred = Credential

// Deprecated: NewCredUsername is a deprecated alias of NewCredentialUsername.
func NewCredUsername(username string) (*Cred, error) {
	return NewCredentialUsername(username)
}

// Deprecated: NewCredUserpassPlaintext is a deprecated alias of NewCredentialUserpassPlaintext.
func NewCredUserpassPlaintext(username string, password string) (*Cred, error) {
	return NewCredentialUserpassPlaintext(username, password)
}

// Deprecated: NewCredSshKey is a deprecated alias of NewCredentialSshKey.
func NewCredSshKey(username string, publicKeyPath string, privateKeyPath string, passphrase string) (*Cred, error) {
	return NewCredentialSSHKey(username, publicKeyPath, privateKeyPath, passphrase)
}

// Deprecated: NewCredSshKeyFromMemory is a deprecated alias of NewCredentialSSHKeyFromMemory.
func NewCredSshKeyFromMemory(username string, publicKey string, privateKey string, passphrase string) (*Cred, error) {
	return NewCredentialSSHKeyFromMemory(username, publicKey, privateKey, passphrase)
}

// Deprecated: NewCredSshKeyFromAgent is a deprecated alias of NewCredentialSSHFromAgent.
func NewCredSshKeyFromAgent(username string) (*Cred, error) {
	return NewCredentialSSHKeyFromAgent(username)
}

// Deprecated: NewCredDefault is a deprecated alias fof NewCredentialDefault.
func NewCredDefault() (*Cred, error) {
	return NewCredentialDefault()
}

// diff.go

const (
	// Deprecated: DiffIgnoreWhitespaceEol is a deprecated alias of DiffIgnoreWhitespaceEOL.
	DiffIgnoreWitespaceEol = DiffIgnoreWhitespaceEOL
)

// features.go

const (
	// Deprecated: FeatureHttps is a deprecated alias of FeatureHTTPS.
	FeatureHttps = FeatureHTTPS

	// Deprecated: FeatureSsh is a deprecated alias of FeatureSSH.
	FeatureSsh = FeatureSSH
)

// git.go

const (
	// Deprecated: ErrClassNone is a deprecated alias of ErrorClassNone.
	ErrClassNone = ErrorClassNone
	// Deprecated: ErrClassNoMemory is a deprecated alias of ErrorClassNoMemory.
	ErrClassNoMemory = ErrorClassNoMemory
	// Deprecated: ErrClassOs is a deprecated alias of ErrorClassOS.
	ErrClassOs = ErrorClassOS
	// Deprecated: ErrClassInvalid is a deprecated alias of ErrorClassInvalid.
	ErrClassInvalid = ErrorClassInvalid
	// Deprecated: ErrClassReference is a deprecated alias of ErrorClassReference.
	ErrClassReference = ErrorClassReference
	// Deprecated: ErrClassZlib is a deprecated alias of ErrorClassZlib.
	ErrClassZlib = ErrorClassZlib
	// Deprecated: ErrClassRepository is a deprecated alias of ErrorClassRepository.
	ErrClassRepository = ErrorClassRepository
	// Deprecated: ErrClassConfig is a deprecated alias of ErrorClassConfig.
	ErrClassConfig = ErrorClassConfig
	// Deprecated: ErrClassRegex is a deprecated alias of ErrorClassRegex.
	ErrClassRegex = ErrorClassRegex
	// Deprecated: ErrClassOdb is a deprecated alias of ErrorClassOdb.
	ErrClassOdb = ErrorClassOdb
	// Deprecated: ErrClassIndex is a deprecated alias of ErrorClassIndex.
	ErrClassIndex = ErrorClassIndex
	// Deprecated: ErrClassObject is a deprecated alias of ErrorClassObject.
	ErrClassObject = ErrorClassObject
	// Deprecated: ErrClassNet is a deprecated alias of ErrorClassNet.
	ErrClassNet = ErrorClassNet
	// Deprecated: ErrClassTag is a deprecated alias of ErrorClassTag.
	ErrClassTag = ErrorClassTag
	// Deprecated: ErrClassTree is a deprecated alias of ErrorClassTree.
	ErrClassTree = ErrorClassTree
	// Deprecated: ErrClassIndexer is a deprecated alias of ErrorClassIndexer.
	ErrClassIndexer = ErrorClassIndexer
	// Deprecated: ErrClassSSL is a deprecated alias of ErrorClassSSL.
	ErrClassSSL = ErrorClassSSL
	// Deprecated: ErrClassSubmodule is a deprecated alias of ErrorClassSubmodule.
	ErrClassSubmodule = ErrorClassSubmodule
	// Deprecated: ErrClassThread is a deprecated alias of ErrorClassThread.
	ErrClassThread = ErrorClassThread
	// Deprecated: ErrClassStash is a deprecated alias of ErrorClassStash.
	ErrClassStash = ErrorClassStash
	// Deprecated: ErrClassCheckout is a deprecated alias of ErrorClassCheckout.
	ErrClassCheckout = ErrorClassCheckout
	// Deprecated: ErrClassFetchHead is a deprecated alias of ErrorClassFetchHead.
	ErrClassFetchHead = ErrorClassFetchHead
	// Deprecated: ErrClassMerge is a deprecated alias of ErrorClassMerge.
	ErrClassMerge = ErrorClassMerge
	// Deprecated: ErrClassSsh is a deprecated alias of ErrorClassSSH.
	ErrClassSsh = ErrorClassSSH
	// Deprecated: ErrClassFilter is a deprecated alias of ErrorClassFilter.
	ErrClassFilter = ErrorClassFilter
	// Deprecated: ErrClassRevert is a deprecated alias of ErrorClassRevert.
	ErrClassRevert = ErrorClassRevert
	// Deprecated: ErrClassCallback is a deprecated alias of ErrorClassCallback.
	ErrClassCallback = ErrorClassCallback
	// Deprecated: ErrClassRebase is a deprecated alias of ErrorClassRebase.
	ErrClassRebase = ErrorClassRebase
	// Deprecated: ErrClassPatch is a deprecated alias of ErrorClassPatch.
	ErrClassPatch = ErrorClassPatch
)

const (
	// Deprecated: ErrOk is a deprecated alias of ErrorCodeOK.
	ErrOk = ErrorCodeOK
	// Deprecated: ErrGeneric is a deprecated alias of ErrorCodeGeneric.
	ErrGeneric = ErrorCodeGeneric
	// Deprecated: ErrNotFound is a deprecated alias of ErrorCodeNotFound.
	ErrNotFound = ErrorCodeNotFound
	// Deprecated: ErrExists is a deprecated alias of ErrorCodeExists.
	ErrExists = ErrorCodeExists
	// Deprecated: ErrAmbiguous is a deprecated alias of ErrorCodeAmbiguous.
	ErrAmbiguous = ErrorCodeAmbiguous
	// Deprecated: ErrAmbigious is a deprecated alias of ErrorCodeAmbiguous.
	ErrAmbigious = ErrorCodeAmbiguous
	// Deprecated: ErrBuffs is a deprecated alias of ErrorCodeBuffs.
	ErrBuffs = ErrorCodeBuffs
	// Deprecated: ErrUser is a deprecated alias of ErrorCodeUser.
	ErrUser = ErrorCodeUser
	// Deprecated: ErrBareRepo is a deprecated alias of ErrorCodeBareRepo.
	ErrBareRepo = ErrorCodeBareRepo
	// Deprecated: ErrUnbornBranch is a deprecated alias of ErrorCodeUnbornBranch.
	ErrUnbornBranch = ErrorCodeUnbornBranch
	// Deprecated: ErrUnmerged is a deprecated alias of ErrorCodeUnmerged.
	ErrUnmerged = ErrorCodeUnmerged
	// Deprecated: ErrNonFastForward is a deprecated alias of ErrorCodeNonFastForward.
	ErrNonFastForward = ErrorCodeNonFastForward
	// Deprecated: ErrInvalidSpec is a deprecated alias of ErrorCodeInvalidSpec.
	ErrInvalidSpec = ErrorCodeInvalidSpec
	// Deprecated: ErrConflict is a deprecated alias of ErrorCodeConflict.
	ErrConflict = ErrorCodeConflict
	// Deprecated: ErrLocked is a deprecated alias of ErrorCodeLocked.
	ErrLocked = ErrorCodeLocked
	// Deprecated: ErrModified is a deprecated alias of ErrorCodeModified.
	ErrModified = ErrorCodeModified
	// Deprecated: ErrAuth is a deprecated alias of ErrorCodeAuth.
	ErrAuth = ErrorCodeAuth
	// Deprecated: ErrCertificate is a deprecated alias of ErrorCodeCertificate.
	ErrCertificate = ErrorCodeCertificate
	// Deprecated: ErrApplied is a deprecated alias of ErrorCodeApplied.
	ErrApplied = ErrorCodeApplied
	// Deprecated: ErrPeel is a deprecated alias of ErrorCodePeel.
	ErrPeel = ErrorCodePeel
	// Deprecated: ErrEOF is a deprecated alias of ErrorCodeEOF.
	ErrEOF = ErrorCodeEOF
	// Deprecated: ErrUncommitted is a deprecated alias of ErrorCodeUncommitted.
	ErrUncommitted = ErrorCodeUncommitted
	// Deprecated: ErrDirectory is a deprecated alias of ErrorCodeDirectory.
	ErrDirectory = ErrorCodeDirectory
	// Deprecated: ErrMergeConflict is a deprecated alias of ErrorCodeMergeConflict.
	ErrMergeConflict = ErrorCodeMergeConflict
	// Deprecated: ErrPassthrough is a deprecated alias of ErrorCodePassthrough.
	ErrPassthrough = ErrorCodePassthrough
	// Deprecated: ErrIterOver is a deprecated alias of ErrorCodeIterOver.
	ErrIterOver = ErrorCodeIterOver
	// Deprecated: ErrApplyFail is a deprecated alias of ErrorCodeApplyFail.
	ErrApplyFail = ErrorCodeApplyFail
)

// index.go

// Deprecated: IndexAddOpts is a deprecated alias of IndexAddOption.
type IndexAddOpts = IndexAddOption

// Deprecated: IndexStageOpts is a deprecated alias of IndexStageState.
type IndexStageOpts = IndexStageState

// submodule.go

// Deprecated: SubmoduleCbk is a deprecated alias of SubmoduleCallback.
type SubmoduleCbk = SubmoduleCallback

// Deprecated: SubmoduleVisitor is not used.
func SubmoduleVisitor(csub unsafe.Pointer, name *C.char, handle unsafe.Pointer) C.int {
	sub := &Submodule{ptr: (*C.git_submodule)(csub)}

	callback, ok := pointerHandles.Get(handle).(SubmoduleCallback)
	if !ok {
		panic("invalid submodule visitor callback")
	}
	return (C.int)(callback(sub, C.GoString(name)))
}

// tree.go

// Deprecated: CallbackGitTreeWalk is not used.
func CallbackGitTreeWalk(_root *C.char, entry *C.git_tree_entry, ptr unsafe.Pointer) C.int {
	root := C.GoString(_root)

	if callback, ok := pointerHandles.Get(ptr).(TreeWalkCallback); ok {
		return C.int(callback(root, newTreeEntry(entry)))
	} else {
		panic("invalid treewalk callback")
	}
}
