package git

/*
#include <git2.h>
#include <git2/errors.h>

extern git_merge_head** _go_git_make_merge_head_array(size_t len);
extern void _go_git_merge_head_array_set(git_merge_head** array, git_merge_head* ptr, size_t n);
extern git_merge_head* _go_git_merge_head_array_get(git_merge_head** array, size_t n);

*/
import "C"
import (
	"runtime"
	"unsafe"
)

type MergeHead struct {
	ptr *C.git_merge_head
}

func newMergeHeadFromC(c *C.git_merge_head) *MergeHead {
	mh := &MergeHead{ptr: c}
	runtime.SetFinalizer(mh, (*MergeHead).Free)
	return mh
}

func (mh *MergeHead) Free() {
	runtime.SetFinalizer(mh, nil)
	C.git_merge_head_free(mh.ptr)
}

func (r *Repository) MergeHeadFromFetchHead(branchName string, remoteURL string, oid *Oid) (*MergeHead, error) {
	mh := &MergeHead{}

	cbranchName := C.CString(branchName)
	defer C.free(unsafe.Pointer(cbranchName))

	cremoteURL := C.CString(remoteURL)
	defer C.free(unsafe.Pointer(cremoteURL))

	ret := C.git_merge_head_from_fetchhead(&mh.ptr, r.ptr, cbranchName, cremoteURL, oid.toC())
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(mh, (*MergeHead).Free)
	return mh, nil
}

func (r *Repository) MergeHeadFromId(oid *Oid) (*MergeHead, error) {
	mh := &MergeHead{}

	ret := C.git_merge_head_from_id(&mh.ptr, r.ptr, oid.toC())
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(mh, (*MergeHead).Free)
	return mh, nil
}

func (r *Repository) MergeHeadFromRef(ref *Reference) (*MergeHead, error) {
	mh := &MergeHead{}

	ret := C.git_merge_head_from_ref(&mh.ptr, r.ptr, ref.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(mh, (*MergeHead).Free)
	return mh, nil
}

type MergeTreeFlag int

const (
	MergeTreeFindRenames MergeTreeFlag = C.GIT_MERGE_TREE_FIND_RENAMES
)

type MergeOptions struct {
	Version uint
	Flags   MergeTreeFlag

	RenameThreshold uint
	TargetLimit     uint
	FileFavor       MergeFileFavorType

	//TODO: Diff similarity metric
}

func mergeOptionsFromC(opts *C.git_merge_options) MergeOptions {
	return MergeOptions{
		Version:         uint(opts.version),
		Flags:           MergeTreeFlag(opts.flags),
		RenameThreshold: uint(opts.rename_threshold),
		TargetLimit:     uint(opts.target_limit),
		FileFavor:       MergeFileFavorType(opts.file_favor),
	}
}

func DefaultMergeOptions() (MergeOptions, error) {
	opts := C.git_merge_options{}
	ecode := C.git_merge_init_options(&opts, C.GIT_MERGE_OPTIONS_VERSION)
	if ecode < 0 {
		return MergeOptions{}, MakeGitError(ecode)
	}
	return mergeOptionsFromC(&opts), nil
}

func (mo *MergeOptions) toC() *C.git_merge_options {
	if mo == nil {
		return nil
	}
	return &C.git_merge_options{
		version:          C.uint(mo.Version),
		flags:            C.git_merge_tree_flag_t(mo.Flags),
		rename_threshold: C.uint(mo.RenameThreshold),
		target_limit:     C.uint(mo.TargetLimit),
		file_favor:       C.git_merge_file_favor_t(mo.FileFavor),
	}
}

type MergeFileFavorType int

const (
	MergeFileFavorNormal MergeFileFavorType = C.GIT_MERGE_FILE_FAVOR_NORMAL
	MergeFileFavorOurs                      = C.GIT_MERGE_FILE_FAVOR_OURS
	MergeFileFavorTheirs                    = C.GIT_MERGE_FILE_FAVOR_THEIRS
	MergeFileFavorUnion                     = C.GIT_MERGE_FILE_FAVOR_UNION
)

func (r *Repository) Merge(theirHeads []*MergeHead, mergeOptions *MergeOptions, checkoutOptions *CheckoutOpts) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cMergeOpts := mergeOptions.toC()
	cCheckoutOpts := checkoutOptions.toC()

	gmerge_head_array := make([]*C.git_merge_head, len(theirHeads))
	for i := 0; i < len(theirHeads); i++ {
		gmerge_head_array[i] = theirHeads[i].ptr
	}
	ptr := unsafe.Pointer(&gmerge_head_array[0])
	err := C.git_merge(r.ptr, (**C.git_merge_head)(ptr), C.size_t(len(theirHeads)), cMergeOpts, cCheckoutOpts)
	if err < 0 {
		return MakeGitError(err)
	}
	return nil
}

func (r *Repository) MergeCommits(ours *Commit, theirs *Commit, options MergeOptions) (*Index, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	copts := options.toC()

	idx := &Index{}

	ret := C.git_merge_commits(&idx.ptr, r.ptr, ours.ptr, theirs.ptr, copts)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(idx, (*Index).Free)
	return idx, nil
}

func (r *Repository) MergeTrees(ancestor *Tree, ours *Tree, theirs *Tree, options MergeOptions) (*Index, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	copts := options.toC()

	idx := &Index{}

	ret := C.git_merge_trees(&idx.ptr, r.ptr, ancestor.ptr, ours.ptr, theirs.ptr, copts)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(idx, (*Index).Free)
	return idx, nil
}

func (r *Repository) MergeBase(one *Oid, two *Oid) (*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var oid C.git_oid
	ret := C.git_merge_base(&oid, r.ptr, one.toC(), two.toC())
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newOidFromC(&oid), nil
}

//TODO: int git_merge_base_many(git_oid *out, git_repository *repo, size_t length, const git_oid input_array[]);
