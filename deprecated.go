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

// BlobChunkCallback is not used.
type BlobChunkCallback func(maxLen int) ([]byte, error)

// BlobCallbackData is not used.
type BlobCallbackData struct {
	Callback BlobChunkCallback
	Error    error
}

// checkout.go

// CheckoutOpts is a deprecated alias of CheckoutOptions.
type CheckoutOpts = CheckoutOptions

// features.go

const (
	// FeatureHttps is a deprecated alias of FeatureHTTPS.
	FeatureHttps = FeatureHTTPS

	// FeatureSsh is a deprecated alias of FeatureSSH.
	FeatureSsh = FeatureSSH
)

// git.go

const (
	ErrClassNone       = ErrorClassNone
	ErrClassNoMemory   = ErrorClassNoMemory
	ErrClassOs         = ErrorClassOS
	ErrClassInvalid    = ErrorClassInvalid
	ErrClassReference  = ErrorClassReference
	ErrClassZlib       = ErrorClassZlib
	ErrClassRepository = ErrorClassRepository
	ErrClassConfig     = ErrorClassConfig
	ErrClassRegex      = ErrorClassRegex
	ErrClassOdb        = ErrorClassOdb
	ErrClassIndex      = ErrorClassIndex
	ErrClassObject     = ErrorClassObject
	ErrClassNet        = ErrorClassNet
	ErrClassTag        = ErrorClassTag
	ErrClassTree       = ErrorClassTree
	ErrClassIndexer    = ErrorClassIndexer
	ErrClassSSL        = ErrorClassSSL
	ErrClassSubmodule  = ErrorClassSubmodule
	ErrClassThread     = ErrorClassThread
	ErrClassStash      = ErrorClassStash
	ErrClassCheckout   = ErrorClassCheckout
	ErrClassFetchHead  = ErrorClassFetchHead
	ErrClassMerge      = ErrorClassMerge
	ErrClassSsh        = ErrorClassSSH
	ErrClassFilter     = ErrorClassFilter
	ErrClassRevert     = ErrorClassRevert
	ErrClassCallback   = ErrorClassCallback
	ErrClassRebase     = ErrorClassRebase
	ErrClassPatch      = ErrorClassPatch
)

const (
	ErrOk             = ErrorCodeOK
	ErrGeneric        = ErrorCodeGeneric
	ErrNotFound       = ErrorCodeNotFound
	ErrExists         = ErrorCodeExists
	ErrAmbiguous      = ErrorCodeAmbiguous
	ErrAmbigious      = ErrorCodeAmbiguous
	ErrBuffs          = ErrorCodeBuffs
	ErrUser           = ErrorCodeUser
	ErrBareRepo       = ErrorCodeBareRepo
	ErrUnbornBranch   = ErrorCodeUnbornBranch
	ErrUnmerged       = ErrorCodeUnmerged
	ErrNonFastForward = ErrorCodeNonFastForward
	ErrInvalidSpec    = ErrorCodeInvalidSpec
	ErrConflict       = ErrorCodeConflict
	ErrLocked         = ErrorCodeLocked
	ErrModified       = ErrorCodeModified
	ErrAuth           = ErrorCodeAuth
	ErrCertificate    = ErrorCodeCertificate
	ErrApplied        = ErrorCodeApplied
	ErrPeel           = ErrorCodePeel
	ErrEOF            = ErrorCodeEOF
	ErrUncommitted    = ErrorCodeUncommitted
	ErrDirectory      = ErrorCodeDirectory
	ErrMergeConflict  = ErrorCodeMergeConflict
	ErrPassthrough    = ErrorCodePassthrough
	ErrIterOver       = ErrorCodeIterOver
	ErrApplyFail      = ErrorCodeApplyFail
)

// index.go

// IndexAddOpts is a deprecated alias of IndexAddOption.
type IndexAddOpts = IndexAddOption

// IndexStageOpts is a deprecated alias of IndexStageState.
type IndexStageOpts = IndexStageState

// submodule.go

// SubmoduleCbk is a deprecated alias of SubmoduleCallback.
type SubmoduleCbk = SubmoduleCallback

// SubmoduleVisitor is not used.
func SubmoduleVisitor(csub unsafe.Pointer, name *C.char, handle unsafe.Pointer) C.int {
	sub := &Submodule{(*C.git_submodule)(csub), nil}

	callback, ok := pointerHandles.Get(handle).(SubmoduleCallback)
	if !ok {
		panic("invalid submodule visitor callback")
	}
	return (C.int)(callback(sub, C.GoString(name)))
}

// tree.go

// CallbackGitTreeWalk is not used.
func CallbackGitTreeWalk(_root *C.char, entry *C.git_tree_entry, ptr unsafe.Pointer) C.int {
	root := C.GoString(_root)

	if callback, ok := pointerHandles.Get(ptr).(TreeWalkCallback); ok {
		return C.int(callback(root, newTreeEntry(entry)))
	} else {
		panic("invalid treewalk callback")
	}
}
