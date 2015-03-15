package git

/*
#include <git2.h>
*/
import "C"
import (
	"os"
	"runtime"
	"unsafe"
)

type CheckoutStrategy uint

const (
	CheckoutNone                      CheckoutStrategy = C.GIT_CHECKOUT_NONE                         // Dry run, no actual updates
	CheckoutSafe                      CheckoutStrategy = C.GIT_CHECKOUT_SAFE                         // Allow safe updates that cannot overwrite uncommitted data
	CheckoutSafeCreate                CheckoutStrategy = C.GIT_CHECKOUT_SAFE_CREATE                  // Allow safe updates plus creation of missing files
	CheckoutForce                     CheckoutStrategy = C.GIT_CHECKOUT_FORCE                        // Allow all updates to force working directory to look like index
	CheckoutAllowConflicts            CheckoutStrategy = C.GIT_CHECKOUT_ALLOW_CONFLICTS              // Allow checkout to make safe updates even if conflicts are found
	CheckoutRemoveUntracked           CheckoutStrategy = C.GIT_CHECKOUT_REMOVE_UNTRACKED             // Remove untracked files not in index (that are not ignored)
	CheckoutRemoveIgnored             CheckoutStrategy = C.GIT_CHECKOUT_REMOVE_IGNORED               // Remove ignored files not in index
	CheckoutUpdateOnly                CheckoutStrategy = C.GIT_CHECKOUT_UPDATE_ONLY                  // Only update existing files, don't create new ones
	CheckoutDontUpdateIndex           CheckoutStrategy = C.GIT_CHECKOUT_DONT_UPDATE_INDEX            // Normally checkout updates index entries as it goes; this stops that
	CheckoutNoRefresh                 CheckoutStrategy = C.GIT_CHECKOUT_NO_REFRESH                   // Don't refresh index/config/etc before doing checkout
	CheckoutDisablePathspecMatch      CheckoutStrategy = C.GIT_CHECKOUT_DISABLE_PATHSPEC_MATCH       // Treat pathspec as simple list of exact match file paths
	CheckoutSkipUnmerged              CheckoutStrategy = C.GIT_CHECKOUT_SKIP_UNMERGED                // Allow checkout to skip unmerged files (NOT IMPLEMENTED)
	CheckoutUserOurs                  CheckoutStrategy = C.GIT_CHECKOUT_USE_OURS                     // For unmerged files, checkout stage 2 from index (NOT IMPLEMENTED)
	CheckoutUseTheirs                 CheckoutStrategy = C.GIT_CHECKOUT_USE_THEIRS                   // For unmerged files, checkout stage 3 from index (NOT IMPLEMENTED)
	CheckoutUpdateSubmodules          CheckoutStrategy = C.GIT_CHECKOUT_UPDATE_SUBMODULES            // Recursively checkout submodules with same options (NOT IMPLEMENTED)
	CheckoutUpdateSubmodulesIfChanged CheckoutStrategy = C.GIT_CHECKOUT_UPDATE_SUBMODULES_IF_CHANGED // Recursively checkout submodules if HEAD moved in super repo (NOT IMPLEMENTED)
)

type CheckoutOpts struct {
	Strategy        CheckoutStrategy // Default will be a dry run
	DisableFilters  bool             // Don't apply filters like CRLF conversion
	DirMode         os.FileMode      // Default is 0755
	FileMode        os.FileMode      // Default is 0644 or 0755 as dictated by blob
	FileOpenFlags   int              // Default is O_CREAT | O_TRUNC | O_WRONLY
	TargetDirectory string           // Alternative checkout path to workdir
}

func checkoutOptionsFromC(c *C.git_checkout_options) CheckoutOpts {
	opts := CheckoutOpts{}
	opts.Strategy = CheckoutStrategy(c.checkout_strategy)
	opts.DisableFilters = c.disable_filters != 0
	opts.DirMode = os.FileMode(c.dir_mode)
	opts.FileMode = os.FileMode(c.file_mode)
	opts.FileOpenFlags = int(c.file_open_flags)
	if c.target_directory != nil {
		opts.TargetDirectory = C.GoString(c.target_directory)
	}
	return opts
}

func (opts *CheckoutOpts) toC() *C.git_checkout_options {
	if opts == nil {
		return nil
	}
	c := C.git_checkout_options{}
	populateCheckoutOpts(&c, opts)
	return &c
}

// Convert the CheckoutOpts struct to the corresponding
// C-struct. Returns a pointer to ptr, or nil if opts is nil, in order
// to help with what to pass.
func populateCheckoutOpts(ptr *C.git_checkout_options, opts *CheckoutOpts) *C.git_checkout_options {
	if opts == nil {
		return nil
	}

	C.git_checkout_init_options(ptr, 1)
	ptr.checkout_strategy = C.uint(opts.Strategy)
	ptr.disable_filters = cbool(opts.DisableFilters)
	ptr.dir_mode = C.uint(opts.DirMode.Perm())
	ptr.file_mode = C.uint(opts.FileMode.Perm())
	if opts.TargetDirectory != "" {
		ptr.target_directory = C.CString(opts.TargetDirectory)
	}
	return ptr
}

func freeCheckoutOpts(ptr *C.git_checkout_options) {
	if ptr == nil {
		return
	}
	C.free(unsafe.Pointer(ptr.target_directory))
}

// Updates files in the index and the working tree to match the content of
// the commit pointed at by HEAD. opts may be nil.
func (v *Repository) CheckoutHead(opts *CheckoutOpts) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cOpts := opts.toC()
	defer freeCheckoutOpts(cOpts)

	ret := C.git_checkout_head(v.ptr, cOpts)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// Updates files in the working tree to match the content of the given
// index. If index is nil, the repository's index will be used. opts
// may be nil.
func (v *Repository) CheckoutIndex(index *Index, opts *CheckoutOpts) error {
	var iptr *C.git_index = nil
	if index != nil {
		iptr = index.ptr
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cOpts := opts.toC()
	defer freeCheckoutOpts(cOpts)

	ret := C.git_checkout_index(v.ptr, iptr, cOpts)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (v *Repository) CheckoutTree(tree *Tree, opts *CheckoutOpts) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cOpts := opts.toC()
	defer freeCheckoutOpts(cOpts)

	ret := C.git_checkout_tree(v.ptr, tree.ptr, cOpts)
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}
