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
	C.git_merge_head_free(mh.ptr)
	runtime.SetFinalizer(mh, nil)
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

type MergeFlag int

const (
	MergeFlagDefault MergeFlag = iota
	MergeNoFastForward
	MergeFastForwardOnly
)

type MergeOptions struct {
	Version uint
	Flags   MergeFlag

	TreeOptions MergeTreeOptions
	//TODO: CheckoutOptions CheckoutOptions
}

func (mo *MergeOptions) toC() *C.git_merge_opts {
	return &C.git_merge_opts{
		version:         C.uint(mo.Version),
		merge_flags:     C.git_merge_flags_t(mo.Flags),
		merge_tree_opts: *mo.TreeOptions.toC(),
	}
}

type MergeTreeFlag int

const (
	MergeTreeFindRenames MergeTreeFlag = 1 << iota
)

type MergeFileFavorType int

const (
	MergeFileFavorNormal MergeFileFavorType = iota
	MergeFileFavorOurs
	MergeFileFavorTheirs
	MergeFileFavorUnion
)

type MergeTreeOptions struct {
	Version         uint
	Flags           MergeTreeFlag
	RenameThreshold uint
	TargetLimit     uint
	//TODO: SimilarityMetric *DiffSimilarityMetric
	FileFavor MergeFileFavorType
}

func (mo *MergeTreeOptions) toC() *C.git_merge_tree_opts {
	return &C.git_merge_tree_opts{
		version:          C.uint(mo.Version),
		flags:            C.git_merge_tree_flag_t(mo.Flags),
		rename_threshold: C.uint(mo.RenameThreshold),
		target_limit:     C.uint(mo.TargetLimit),
		file_favor:       C.git_merge_file_favor_t(mo.FileFavor),
	}
}

type MergeResult struct {
	ptr *C.git_merge_result
}

func newMergeResultFromC(c *C.git_merge_result) *MergeResult {
	mr := &MergeResult{ptr: c}
	runtime.SetFinalizer(mr, (*MergeResult).Free)
	return mr
}

func (mr *MergeResult) Free() {
	runtime.SetFinalizer(mr, nil)
	C.git_merge_result_free(mr.ptr)
}

func (mr *MergeResult) IsFastForward() bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_merge_result_is_fastforward(mr.ptr)
	return ret != 0
}

func (mr *MergeResult) IsUpToDate() bool {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_merge_result_is_uptodate(mr.ptr)
	return ret != 0
}

func (mr *MergeResult) FastForwardId() (*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var oid C.git_oid
	ret := C.git_merge_result_fastforward_id(&oid, mr.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newOidFromC(&oid), nil
}

func (r *Repository) Merge(theirHeads []MergeHead, options MergeOptions) (*MergeResult, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var result *C.git_merge_result

	copts := options.toC()

	cmerge_head_array := C._go_git_make_merge_head_array(C.size_t(len(theirHeads)))
	defer C.free(unsafe.Pointer(cmerge_head_array))

	for i, _ := range theirHeads {
		C._go_git_merge_head_array_set(cmerge_head_array, theirHeads[i].ptr, C.size_t(i))
	}

	err := C.git_merge(&result, r.ptr, cmerge_head_array, C.size_t(len(theirHeads)), copts)
	if err < 0 {
		return nil, MakeGitError(err)
	}
	return newMergeResultFromC(result), nil
}

func (r *Repository) MergeCommits(ours *Commit, theirs *Commit, options MergeTreeOptions) (*Index, error) {
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

func (r *Repository) MergeTrees(ancestor *Tree, ours *Tree, theirs *Tree, options MergeTreeOptions) (*Index, error) {
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
