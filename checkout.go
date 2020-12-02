package git

/*
#include <git2.h>

extern void _go_git_populate_checkout_callbacks(git_checkout_options *opts);
*/
import "C"
import (
	"os"
	"runtime"
	"unsafe"
)

type CheckoutNotifyType uint
type CheckoutStrategy uint

const (
	CheckoutNotifyNone      CheckoutNotifyType = C.GIT_CHECKOUT_NOTIFY_NONE
	CheckoutNotifyConflict  CheckoutNotifyType = C.GIT_CHECKOUT_NOTIFY_CONFLICT
	CheckoutNotifyDirty     CheckoutNotifyType = C.GIT_CHECKOUT_NOTIFY_DIRTY
	CheckoutNotifyUpdated   CheckoutNotifyType = C.GIT_CHECKOUT_NOTIFY_UPDATED
	CheckoutNotifyUntracked CheckoutNotifyType = C.GIT_CHECKOUT_NOTIFY_UNTRACKED
	CheckoutNotifyIgnored   CheckoutNotifyType = C.GIT_CHECKOUT_NOTIFY_IGNORED
	CheckoutNotifyAll       CheckoutNotifyType = C.GIT_CHECKOUT_NOTIFY_ALL

	CheckoutNone                      CheckoutStrategy = C.GIT_CHECKOUT_NONE                         // Dry run, no actual updates
	CheckoutSafe                      CheckoutStrategy = C.GIT_CHECKOUT_SAFE                         // Allow safe updates that cannot overwrite uncommitted data
	CheckoutForce                     CheckoutStrategy = C.GIT_CHECKOUT_FORCE                        // Allow all updates to force working directory to look like index
	CheckoutRecreateMissing           CheckoutStrategy = C.GIT_CHECKOUT_RECREATE_MISSING             // Allow checkout to recreate missing files
	CheckoutAllowConflicts            CheckoutStrategy = C.GIT_CHECKOUT_ALLOW_CONFLICTS              // Allow checkout to make safe updates even if conflicts are found
	CheckoutRemoveUntracked           CheckoutStrategy = C.GIT_CHECKOUT_REMOVE_UNTRACKED             // Remove untracked files not in index (that are not ignored)
	CheckoutRemoveIgnored             CheckoutStrategy = C.GIT_CHECKOUT_REMOVE_IGNORED               // Remove ignored files not in index
	CheckoutUpdateOnly                CheckoutStrategy = C.GIT_CHECKOUT_UPDATE_ONLY                  // Only update existing files, don't create new ones
	CheckoutDontUpdateIndex           CheckoutStrategy = C.GIT_CHECKOUT_DONT_UPDATE_INDEX            // Normally checkout updates index entries as it goes; this stops that
	CheckoutNoRefresh                 CheckoutStrategy = C.GIT_CHECKOUT_NO_REFRESH                   // Don't refresh index/config/etc before doing checkout
	CheckoutSkipUnmerged              CheckoutStrategy = C.GIT_CHECKOUT_SKIP_UNMERGED                // Allow checkout to skip unmerged files
	CheckoutUseOurs                   CheckoutStrategy = C.GIT_CHECKOUT_USE_OURS                     // For unmerged files, checkout stage 2 from index
	CheckoutUseTheirs                 CheckoutStrategy = C.GIT_CHECKOUT_USE_THEIRS                   // For unmerged files, checkout stage 3 from index
	CheckoutDisablePathspecMatch      CheckoutStrategy = C.GIT_CHECKOUT_DISABLE_PATHSPEC_MATCH       // Treat pathspec as simple list of exact match file paths
	CheckoutSkipLockedDirectories     CheckoutStrategy = C.GIT_CHECKOUT_SKIP_LOCKED_DIRECTORIES      // Ignore directories in use, they will be left empty
	CheckoutDontOverwriteIgnored      CheckoutStrategy = C.GIT_CHECKOUT_DONT_OVERWRITE_IGNORED       // Don't overwrite ignored files that exist in the checkout target
	CheckoutConflictStyleMerge        CheckoutStrategy = C.GIT_CHECKOUT_CONFLICT_STYLE_MERGE         // Write normal merge files for conflicts
	CheckoutConflictStyleDiff3        CheckoutStrategy = C.GIT_CHECKOUT_CONFLICT_STYLE_DIFF3         // Include common ancestor data in diff3 format files for conflicts
	CheckoutDontRemoveExisting        CheckoutStrategy = C.GIT_CHECKOUT_DONT_REMOVE_EXISTING         // Don't overwrite existing files or folders
	CheckoutDontWriteIndex            CheckoutStrategy = C.GIT_CHECKOUT_DONT_WRITE_INDEX             // Normally checkout writes the index upon completion; this prevents that
	CheckoutUpdateSubmodules          CheckoutStrategy = C.GIT_CHECKOUT_UPDATE_SUBMODULES            // Recursively checkout submodules with same options (NOT IMPLEMENTED)
	CheckoutUpdateSubmodulesIfChanged CheckoutStrategy = C.GIT_CHECKOUT_UPDATE_SUBMODULES_IF_CHANGED // Recursively checkout submodules if HEAD moved in super repo (NOT IMPLEMENTED)
)

type CheckoutNotifyCallback func(why CheckoutNotifyType, path string, baseline, target, workdir DiffFile) error
type CheckoutProgressCallback func(path string, completed, total uint)

type CheckoutOptions struct {
	Strategy         CheckoutStrategy   // Default will be a dry run
	DisableFilters   bool               // Don't apply filters like CRLF conversion
	DirMode          os.FileMode        // Default is 0755
	FileMode         os.FileMode        // Default is 0644 or 0755 as dictated by blob
	FileOpenFlags    int                // Default is O_CREAT | O_TRUNC | O_WRONLY
	NotifyFlags      CheckoutNotifyType // Default will be none
	NotifyCallback   CheckoutNotifyCallback
	ProgressCallback CheckoutProgressCallback
	TargetDirectory  string // Alternative checkout path to workdir
	Paths            []string
	Baseline         *Tree
}

func checkoutOptionsFromC(c *C.git_checkout_options) CheckoutOptions {
	opts := CheckoutOptions{
		Strategy:       CheckoutStrategy(c.checkout_strategy),
		DisableFilters: c.disable_filters != 0,
		DirMode:        os.FileMode(c.dir_mode),
		FileMode:       os.FileMode(c.file_mode),
		FileOpenFlags:  int(c.file_open_flags),
		NotifyFlags:    CheckoutNotifyType(c.notify_flags),
	}
	if c.notify_payload != nil {
		opts.NotifyCallback = pointerHandles.Get(c.notify_payload).(*checkoutCallbackData).options.NotifyCallback
	}
	if c.progress_payload != nil {
		opts.ProgressCallback = pointerHandles.Get(c.progress_payload).(*checkoutCallbackData).options.ProgressCallback
	}
	if c.target_directory != nil {
		opts.TargetDirectory = C.GoString(c.target_directory)
	}
	return opts
}

type checkoutCallbackData struct {
	options     *CheckoutOptions
	errorTarget *error
}

//export checkoutNotifyCallback
func checkoutNotifyCallback(
	why C.git_checkout_notify_t,
	cpath *C.char,
	cbaseline, ctarget, cworkdir, handle unsafe.Pointer,
) C.int {
	if handle == nil {
		return C.int(ErrorCodeOK)
	}
	path := C.GoString(cpath)
	var baseline, target, workdir DiffFile
	if cbaseline != nil {
		baseline = diffFileFromC((*C.git_diff_file)(cbaseline))
	}
	if ctarget != nil {
		target = diffFileFromC((*C.git_diff_file)(ctarget))
	}
	if cworkdir != nil {
		workdir = diffFileFromC((*C.git_diff_file)(cworkdir))
	}
	data := pointerHandles.Get(handle).(*checkoutCallbackData)
	if data.options.NotifyCallback == nil {
		return C.int(ErrorCodeOK)
	}
	err := data.options.NotifyCallback(CheckoutNotifyType(why), path, baseline, target, workdir)
	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}
	return C.int(ErrorCodeOK)
}

//export checkoutProgressCallback
func checkoutProgressCallback(
	path *C.char,
	completed_steps, total_steps C.size_t,
	handle unsafe.Pointer,
) {
	data := pointerHandles.Get(handle).(*checkoutCallbackData)
	if data.options.ProgressCallback == nil {
		return
	}
	data.options.ProgressCallback(C.GoString(path), uint(completed_steps), uint(total_steps))
}

// populateCheckoutOptions populates the provided C-struct with the contents of
// the provided CheckoutOptions struct.  Returns copts, or nil if opts is nil,
// in order to help with what to pass.
func populateCheckoutOptions(copts *C.git_checkout_options, opts *CheckoutOptions, errorTarget *error) *C.git_checkout_options {
	C.git_checkout_options_init(copts, C.GIT_CHECKOUT_OPTIONS_VERSION)
	if opts == nil {
		return nil
	}

	copts.checkout_strategy = C.uint(opts.Strategy)
	copts.disable_filters = cbool(opts.DisableFilters)
	copts.dir_mode = C.uint(opts.DirMode.Perm())
	copts.file_mode = C.uint(opts.FileMode.Perm())
	copts.notify_flags = C.uint(opts.NotifyFlags)
	if opts.NotifyCallback != nil || opts.ProgressCallback != nil {
		C._go_git_populate_checkout_callbacks(copts)
		data := &checkoutCallbackData{
			options:     opts,
			errorTarget: errorTarget,
		}
		payload := pointerHandles.Track(data)
		if opts.NotifyCallback != nil {
			copts.notify_payload = payload
		}
		if opts.ProgressCallback != nil {
			copts.progress_payload = payload
		}
	}
	if opts.TargetDirectory != "" {
		copts.target_directory = C.CString(opts.TargetDirectory)
	}
	if len(opts.Paths) > 0 {
		copts.paths.strings = makeCStringsFromStrings(opts.Paths)
		copts.paths.count = C.size_t(len(opts.Paths))
	}

	if opts.Baseline != nil {
		copts.baseline = opts.Baseline.cast_ptr
	}

	return copts
}

func freeCheckoutOptions(copts *C.git_checkout_options) {
	if copts == nil {
		return
	}
	C.free(unsafe.Pointer(copts.target_directory))
	if copts.paths.count > 0 {
		freeStrarray(&copts.paths)
	}
	if copts.notify_payload != nil {
		pointerHandles.Untrack(copts.notify_payload)
	} else if copts.progress_payload != nil {
		pointerHandles.Untrack(copts.progress_payload)
	}
}

// Updates files in the index and the working tree to match the content of
// the commit pointed at by HEAD. opts may be nil.
func (v *Repository) CheckoutHead(opts *CheckoutOptions) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	cOpts := populateCheckoutOptions(&C.git_checkout_options{}, opts, &err)
	defer freeCheckoutOptions(cOpts)

	ret := C.git_checkout_head(v.ptr, cOpts)
	runtime.KeepAlive(v)

	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// Updates files in the working tree to match the content of the given
// index. If index is nil, the repository's index will be used. opts
// may be nil.
func (v *Repository) CheckoutIndex(index *Index, opts *CheckoutOptions) error {
	var iptr *C.git_index = nil
	if index != nil {
		iptr = index.ptr
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	cOpts := populateCheckoutOptions(&C.git_checkout_options{}, opts, &err)
	defer freeCheckoutOptions(cOpts)

	ret := C.git_checkout_index(v.ptr, iptr, cOpts)
	runtime.KeepAlive(v)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

func (v *Repository) CheckoutTree(tree *Tree, opts *CheckoutOptions) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	cOpts := populateCheckoutOptions(&C.git_checkout_options{}, opts, &err)
	defer freeCheckoutOptions(cOpts)

	ret := C.git_checkout_tree(v.ptr, tree.ptr, cOpts)
	runtime.KeepAlive(v)
	runtime.KeepAlive(tree)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}
