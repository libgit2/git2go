package git

/*
#cgo pkg-config: libgit2
#include <git2.h>
git_checkout_opts git_checkout_opts_init() {
	git_checkout_opts ret = GIT_CHECKOUT_OPTS_INIT;
	return ret;
}
*/
import "C"
import (
	"os"
)

type CheckoutStrategy uint

const (
	CHECKOUT_NONE                         CheckoutStrategy = C.GIT_CHECKOUT_NONE                         // Dry run, no actual updates
	CHECKOUT_SAFE                                          = C.GIT_CHECKOUT_SAFE                         // Allow safe updates that cannot overwrite uncommitted data
	CHECKOUT_SAFE_CREATE                                   = C.GIT_CHECKOUT_SAFE_CREATE                  // Allow safe updates plus creation of missing files
	CHECKOUT_FORCE                                         = C.GIT_CHECKOUT_FORCE                        // Allow all updates to force working directory to look like index
	CHECKOUT_ALLOW_CONFLICTS                               = C.GIT_CHECKOUT_ALLOW_CONFLICTS              // Allow checkout to make safe updates even if conflicts are found
	CHECKOUT_REMOVE_UNTRACKED                              = C.GIT_CHECKOUT_REMOVE_UNTRACKED             // Remove untracked files not in index (that are not ignored)
	CHECKOUT_REMOVE_IGNORED                                = C.GIT_CHECKOUT_REMOVE_IGNORED               // Remove ignored files not in index
	CHECKOUT_UPDATE_ONLY                                   = C.GIT_CHECKOUT_UPDATE_ONLY                  // Only update existing files, don't create new ones
	CHECKOUT_DONT_UPDATE_INDEX                             = C.GIT_CHECKOUT_DONT_UPDATE_INDEX            // Normally checkout updates index entries as it goes; this stops that
	CHECKOUT_NO_REFRESH                                    = C.GIT_CHECKOUT_NO_REFRESH                   // Don't refresh index/config/etc before doing checkout
	CHECKOUT_DISABLE_PATHSPEC_MATCH                        = C.GIT_CHECKOUT_DISABLE_PATHSPEC_MATCH       // Treat pathspec as simple list of exact match file paths
	CHECKOUT_SKIP_UNMERGED                                 = C.GIT_CHECKOUT_SKIP_UNMERGED                // Allow checkout to skip unmerged files (NOT IMPLEMENTED)
	CHECKOUT_USE_OURS                                      = C.GIT_CHECKOUT_USE_OURS                     // For unmerged files, checkout stage 2 from index (NOT IMPLEMENTED)
	CHECKOUT_USE_THEIRS                                    = C.GIT_CHECKOUT_USE_THEIRS                   // For unmerged files, checkout stage 3 from index (NOT IMPLEMENTED)
	CHECKOUT_UPDATE_SUBMODULES                             = C.GIT_CHECKOUT_UPDATE_SUBMODULES            // Recursively checkout submodules with same options (NOT IMPLEMENTED)
	CHECKOUT_UPDATE_SUBMODULES_IF_CHANGED                  = C.GIT_CHECKOUT_UPDATE_SUBMODULES_IF_CHANGED // Recursively checkout submodules if HEAD moved in super repo (NOT IMPLEMENTED)
)

type CheckoutOpts struct {
	Strategy       CheckoutStrategy // Default will be a dry run
	DisableFilters bool				// Don't apply filters like CRLF conversion
	DirMode        os.FileMode		// Default is 0755
	FileMode       os.FileMode		// Default is 0644 or 0755 as dictated by blob
	FileOpenFlags  int				// Default is O_CREAT | O_TRUNC | O_WRONLY
}

// Convert the CheckoutOpts struct to the corresponding C-struct
func populateCheckoutOpts(ptr *C.git_checkout_opts, opts *CheckoutOpts) {
	*ptr = C.git_checkout_opts_init()
	if opts == nil {
		return
	}
	ptr.checkout_strategy = C.uint(opts.Strategy)
	ptr.disable_filters = cbool(opts.DisableFilters)
	ptr.dir_mode = C.uint(opts.DirMode.Perm())
	ptr.file_mode = C.uint(opts.FileMode.Perm())
}

// Updates files in the index and the working tree to match the content of
// the commit pointed at by HEAD.
func (v *Repository) Checkout(opts *CheckoutOpts) error {
	var copts C.git_checkout_opts
	populateCheckoutOpts(&copts, opts)

	ret := C.git_checkout_head(v.ptr, &copts)
	if ret < 0 {
		return LastError()
	}

	return nil
}

// Updates files in the working tree to match the content of the index.
func (v *Repository) CheckoutIndex(index *Index, opts *CheckoutOpts) error {
	var copts C.git_checkout_opts
	populateCheckoutOpts(&copts, opts)

	ret := C.git_checkout_index(v.ptr, index.ptr, &copts)
	if ret < 0 {
		return LastError()
	}

	return nil
}
