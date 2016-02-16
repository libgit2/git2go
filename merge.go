package git

/*
#include <git2.h>

extern git_annotated_commit** _go_git_make_merge_head_array(size_t len);
extern void _go_git_annotated_commit_array_set(git_annotated_commit** array, git_annotated_commit* ptr, size_t n);
extern git_annotated_commit* _go_git_annotated_commit_array_get(git_annotated_commit** array, size_t n);

extern int _go_git_merge_file_wrapper(git_merge_file_result *out, git_merge_file_input *ancestor,
				      git_merge_file_input *ours, git_merge_file_input *theirs,
				      const git_merge_file_options *opts, void *content_ancestor,
				      void *content_ours, void *content_theirs);

*/
import "C"
import (
	"reflect"
	"runtime"
	"unsafe"
)

type AnnotatedCommit struct {
	ptr *C.git_annotated_commit
}

func newAnnotatedCommitFromC(c *C.git_annotated_commit) *AnnotatedCommit {
	mh := &AnnotatedCommit{ptr: c}
	runtime.SetFinalizer(mh, (*AnnotatedCommit).Free)
	return mh
}

func (mh *AnnotatedCommit) Free() {
	runtime.SetFinalizer(mh, nil)
	C.git_annotated_commit_free(mh.ptr)
}

func (r *Repository) AnnotatedCommitFromFetchHead(branchName string, remoteURL string, oid *Oid) (*AnnotatedCommit, error) {
	mh := &AnnotatedCommit{}

	cbranchName := C.CString(branchName)
	defer C.free(unsafe.Pointer(cbranchName))

	cremoteURL := C.CString(remoteURL)
	defer C.free(unsafe.Pointer(cremoteURL))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_annotated_commit_from_fetchhead(&mh.ptr, r.ptr, cbranchName, cremoteURL, oid.toC())
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(mh, (*AnnotatedCommit).Free)
	return mh, nil
}

func (r *Repository) LookupAnnotatedCommit(oid *Oid) (*AnnotatedCommit, error) {
	mh := &AnnotatedCommit{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_annotated_commit_lookup(&mh.ptr, r.ptr, oid.toC())
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(mh, (*AnnotatedCommit).Free)
	return mh, nil
}

func (r *Repository) AnnotatedCommitFromRef(ref *Reference) (*AnnotatedCommit, error) {
	mh := &AnnotatedCommit{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_annotated_commit_from_ref(&mh.ptr, r.ptr, ref.ptr)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(mh, (*AnnotatedCommit).Free)
	return mh, nil
}

type MergeTreeFlag int

const (
	MergeTreeFindRenames MergeTreeFlag = C.GIT_MERGE_TREE_FIND_RENAMES
)

type MergeOptions struct {
	Version     uint
	TreeFlags   MergeTreeFlag

	RenameThreshold uint
	TargetLimit     uint
	FileFavor       MergeFileFavor

	//TODO: Diff similarity metric
}

func mergeOptionsFromC(opts *C.git_merge_options) MergeOptions {
	return MergeOptions{
		Version:         uint(opts.version),
		TreeFlags:           MergeTreeFlag(opts.tree_flags),
		RenameThreshold: uint(opts.rename_threshold),
		TargetLimit:     uint(opts.target_limit),
		FileFavor:       MergeFileFavor(opts.file_favor),
	}
}

func DefaultMergeOptions() (MergeOptions, error) {
	opts := C.git_merge_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

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
		tree_flags:       C.git_merge_tree_flag_t(mo.TreeFlags),
		rename_threshold: C.uint(mo.RenameThreshold),
		target_limit:     C.uint(mo.TargetLimit),
		file_favor:       C.git_merge_file_favor_t(mo.FileFavor),
	}
}

type MergeFileFavor int

const (
	MergeFileFavorNormal MergeFileFavor = C.GIT_MERGE_FILE_FAVOR_NORMAL
	MergeFileFavorOurs   MergeFileFavor = C.GIT_MERGE_FILE_FAVOR_OURS
	MergeFileFavorTheirs MergeFileFavor = C.GIT_MERGE_FILE_FAVOR_THEIRS
	MergeFileFavorUnion  MergeFileFavor = C.GIT_MERGE_FILE_FAVOR_UNION
)

func (r *Repository) Merge(theirHeads []*AnnotatedCommit, mergeOptions *MergeOptions, checkoutOptions *CheckoutOpts) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	cMergeOpts := mergeOptions.toC()
	cCheckoutOpts := checkoutOptions.toC()
	defer freeCheckoutOpts(cCheckoutOpts)

	gmerge_head_array := make([]*C.git_annotated_commit, len(theirHeads))
	for i := 0; i < len(theirHeads); i++ {
		gmerge_head_array[i] = theirHeads[i].ptr
	}
	ptr := unsafe.Pointer(&gmerge_head_array[0])
	err := C.git_merge(r.ptr, (**C.git_annotated_commit)(ptr), C.size_t(len(theirHeads)), cMergeOpts, cCheckoutOpts)
	if err < 0 {
		return MakeGitError(err)
	}
	return nil
}

type MergeAnalysis int

const (
	MergeAnalysisNone        MergeAnalysis = C.GIT_MERGE_ANALYSIS_NONE
	MergeAnalysisNormal      MergeAnalysis = C.GIT_MERGE_ANALYSIS_NORMAL
	MergeAnalysisUpToDate    MergeAnalysis = C.GIT_MERGE_ANALYSIS_UP_TO_DATE
	MergeAnalysisFastForward MergeAnalysis = C.GIT_MERGE_ANALYSIS_FASTFORWARD
	MergeAnalysisUnborn      MergeAnalysis = C.GIT_MERGE_ANALYSIS_UNBORN
)

type MergePreference int

const (
	MergePreferenceNone            MergePreference = C.GIT_MERGE_PREFERENCE_NONE
	MergePreferenceNoFastForward   MergePreference = C.GIT_MERGE_PREFERENCE_NO_FASTFORWARD
	MergePreferenceFastForwardOnly MergePreference = C.GIT_MERGE_PREFERENCE_FASTFORWARD_ONLY
)

// MergeAnalysis returns the possible actions which could be taken by
// a 'git-merge' command. There may be multiple answers, so the first
// return value is a bitmask of MergeAnalysis values.
func (r *Repository) MergeAnalysis(theirHeads []*AnnotatedCommit) (MergeAnalysis, MergePreference, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	gmerge_head_array := make([]*C.git_annotated_commit, len(theirHeads))
	for i := 0; i < len(theirHeads); i++ {
		gmerge_head_array[i] = theirHeads[i].ptr
	}
	ptr := unsafe.Pointer(&gmerge_head_array[0])
	var analysis C.git_merge_analysis_t
	var preference C.git_merge_preference_t
	err := C.git_merge_analysis(&analysis, &preference, r.ptr, (**C.git_annotated_commit)(ptr), C.size_t(len(theirHeads)))
	if err < 0 {
		return MergeAnalysisNone, MergePreferenceNone, MakeGitError(err)
	}
	return MergeAnalysis(analysis), MergePreference(preference), nil

}

func (r *Repository) MergeCommits(ours *Commit, theirs *Commit, options *MergeOptions) (*Index, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	copts := options.toC()

	idx := &Index{}

	ret := C.git_merge_commits(&idx.ptr, r.ptr, ours.cast_ptr, theirs.cast_ptr, copts)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	runtime.SetFinalizer(idx, (*Index).Free)
	return idx, nil
}

func (r *Repository) MergeTrees(ancestor *Tree, ours *Tree, theirs *Tree, options *MergeOptions) (*Index, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	copts := options.toC()

	idx := &Index{}
	var ancestor_ptr *C.git_tree
	if ancestor != nil {
		ancestor_ptr = ancestor.cast_ptr
	}
	ret := C.git_merge_trees(&idx.ptr, r.ptr, ancestor_ptr, ours.cast_ptr, theirs.cast_ptr, copts)
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

// MergeBases retrieves the list of merge bases between two commits.
//
// If none are found, an empty slice is returned and the error is set
// approprately
func (r *Repository) MergeBases(one, two *Oid) ([]*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var coids C.git_oidarray
	ret := C.git_merge_bases(&coids, r.ptr, one.toC(), two.toC())
	if ret < 0 {
		return make([]*Oid, 0), MakeGitError(ret)
	}

	oids := make([]*Oid, coids.count)
	hdr := reflect.SliceHeader {
		Data: uintptr(unsafe.Pointer(coids.ids)),
		Len: int(coids.count),
		Cap: int(coids.count),
	}

	goSlice := *(*[]C.git_oid)(unsafe.Pointer(&hdr))

	for i, cid := range goSlice {
		oids[i] = newOidFromC(&cid)
	}

	return oids, nil
}

//TODO: int git_merge_base_many(git_oid *out, git_repository *repo, size_t length, const git_oid input_array[]);
//TODO: GIT_EXTERN(int) git_merge_base_octopus(git_oid *out,git_repository *repo,size_t length,const git_oid input_array[]);

type MergeFileResult struct {
	Automergeable bool
	Path          string
	Mode          uint
	Contents      []byte
	ptr           *C.git_merge_file_result
}

func newMergeFileResultFromC(c *C.git_merge_file_result) *MergeFileResult {
	var path string
	if c.path != nil {
		path = C.GoString(c.path)
	}

	originalBytes := C.GoBytes(unsafe.Pointer(c.ptr), C.int(c.len))
	gobytes := make([]byte, len(originalBytes))
	copy(gobytes, originalBytes)
	r := &MergeFileResult{
		Automergeable: c.automergeable != 0,
		Path:          path,
		Mode:          uint(c.mode),
		Contents:      gobytes,
		ptr:           c,
	}

	runtime.SetFinalizer(r, (*MergeFileResult).Free)
	return r
}

func (r *MergeFileResult) Free() {
	runtime.SetFinalizer(r, nil)
	C.git_merge_file_result_free(r.ptr)
}

type MergeFileInput struct {
	Path     string
	Mode     uint
	Contents []byte
}

// populate a C struct with merge file input, make sure to use freeMergeFileInput to clean up allocs
func populateCMergeFileInput(c *C.git_merge_file_input, input MergeFileInput) {
	c.path = C.CString(input.Path)
	if input.Contents != nil {
		c.size = C.size_t(len(input.Contents))
	}
	c.mode = C.uint(input.Mode)
}

func freeCMergeFileInput(c *C.git_merge_file_input) {
	C.free(unsafe.Pointer(c.path))
}

type MergeFileFlags int

const (
	MergeFileDefault MergeFileFlags = C.GIT_MERGE_FILE_DEFAULT

	MergeFileStyleMerge         MergeFileFlags = C.GIT_MERGE_FILE_STYLE_MERGE
	MergeFileStyleDiff          MergeFileFlags = C.GIT_MERGE_FILE_STYLE_DIFF3
	MergeFileStyleSimplifyAlnum MergeFileFlags = C.GIT_MERGE_FILE_SIMPLIFY_ALNUM
)

type MergeFileOptions struct {
	AncestorLabel string
	OurLabel      string
	TheirLabel    string
	Favor         MergeFileFavor
	Flags         MergeFileFlags
}

func mergeFileOptionsFromC(c C.git_merge_file_options) MergeFileOptions {
	return MergeFileOptions{
		AncestorLabel: C.GoString(c.ancestor_label),
		OurLabel:      C.GoString(c.our_label),
		TheirLabel:    C.GoString(c.their_label),
		Favor:         MergeFileFavor(c.favor),
		Flags:         MergeFileFlags(c.flags),
	}
}

func populateCMergeFileOptions(c *C.git_merge_file_options, options MergeFileOptions) {
	c.ancestor_label = C.CString(options.AncestorLabel)
	c.our_label = C.CString(options.OurLabel)
	c.their_label = C.CString(options.TheirLabel)
	c.favor = C.git_merge_file_favor_t(options.Favor)
	c.flags = C.uint(options.Flags)
}

func freeCMergeFileOptions(c *C.git_merge_file_options) {
	C.free(unsafe.Pointer(c.ancestor_label))
	C.free(unsafe.Pointer(c.our_label))
	C.free(unsafe.Pointer(c.their_label))
}

func MergeFile(ancestor MergeFileInput, ours MergeFileInput, theirs MergeFileInput, options *MergeFileOptions) (*MergeFileResult, error) {

	var cancestor C.git_merge_file_input
	var cours C.git_merge_file_input
	var ctheirs C.git_merge_file_input

	populateCMergeFileInput(&cancestor, ancestor)
	defer freeCMergeFileInput(&cancestor)
	populateCMergeFileInput(&cours, ours)
	defer freeCMergeFileInput(&cours)
	populateCMergeFileInput(&ctheirs, theirs)
	defer freeCMergeFileInput(&ctheirs)

	var copts *C.git_merge_file_options
	if options != nil {
		copts = &C.git_merge_file_options{}
		ecode := C.git_merge_file_init_options(copts, C.GIT_MERGE_FILE_OPTIONS_VERSION)
		if ecode < 0 {
			return nil, MakeGitError(ecode)
		}
		populateCMergeFileOptions(copts, *options)
		defer freeCMergeFileOptions(copts)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var result C.git_merge_file_result
	ecode := C._go_git_merge_file_wrapper(&result, &cancestor, &cours, &ctheirs, copts,
		unsafe.Pointer(&ancestor.Contents[0]),
		unsafe.Pointer(&ours.Contents[0]),
		unsafe.Pointer(&theirs.Contents[0]))
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newMergeFileResultFromC(&result), nil

}

// TODO: GIT_EXTERN(int) git_merge_file_from_index(git_merge_file_result *out,git_repository *repo,const git_index_entry *ancestor,	const git_index_entry *ours,	const git_index_entry *theirs,	const git_merge_file_options *opts);
