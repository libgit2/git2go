package git

/*
#include <git2.h>

extern git_annotated_commit** _go_git_make_merge_head_array(size_t len);
extern void _go_git_annotated_commit_array_set(git_annotated_commit** array, git_annotated_commit* ptr, size_t n);
extern git_annotated_commit* _go_git_annotated_commit_array_get(git_annotated_commit** array, size_t n);
extern int _go_git_merge_file(git_merge_file_result*, char*, size_t, char*, unsigned int, char*, size_t, char*, unsigned int, char*, size_t, char*, unsigned int, git_merge_file_options*);

*/
import "C"
import (
	"reflect"
	"runtime"
	"unsafe"
)

type AnnotatedCommit struct {
	ptr *C.git_annotated_commit
	r   *Repository
}

func newAnnotatedCommitFromC(ptr *C.git_annotated_commit, r *Repository) *AnnotatedCommit {
	mh := &AnnotatedCommit{ptr: ptr, r: r}
	runtime.SetFinalizer(mh, (*AnnotatedCommit).Free)
	return mh
}

func (mh *AnnotatedCommit) Id() *Oid {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := newOidFromC(C.git_annotated_commit_id(mh.ptr))
	runtime.KeepAlive(mh)
	return ret
}

func (mh *AnnotatedCommit) Free() {
	runtime.SetFinalizer(mh, nil)
	C.git_annotated_commit_free(mh.ptr)
}

func (r *Repository) AnnotatedCommitFromFetchHead(branchName string, remoteURL string, oid *Oid) (*AnnotatedCommit, error) {
	cbranchName := C.CString(branchName)
	defer C.free(unsafe.Pointer(cbranchName))

	cremoteURL := C.CString(remoteURL)
	defer C.free(unsafe.Pointer(cremoteURL))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_annotated_commit
	ret := C.git_annotated_commit_from_fetchhead(&ptr, r.ptr, cbranchName, cremoteURL, oid.toC())
	runtime.KeepAlive(oid)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	annotatedCommit := newAnnotatedCommitFromC(ptr, r)
	runtime.KeepAlive(r)
	return annotatedCommit, nil
}

func (r *Repository) LookupAnnotatedCommit(oid *Oid) (*AnnotatedCommit, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_annotated_commit
	ret := C.git_annotated_commit_lookup(&ptr, r.ptr, oid.toC())
	runtime.KeepAlive(oid)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	annotatedCommit := newAnnotatedCommitFromC(ptr, r)
	runtime.KeepAlive(r)
	return annotatedCommit, nil
}

func (r *Repository) AnnotatedCommitFromRef(ref *Reference) (*AnnotatedCommit, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_annotated_commit
	ret := C.git_annotated_commit_from_ref(&ptr, r.ptr, ref.ptr)
	runtime.KeepAlive(r)
	runtime.KeepAlive(ref)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	annotatedCommit := newAnnotatedCommitFromC(ptr, r)
	runtime.KeepAlive(r)
	return annotatedCommit, nil
}

func (r *Repository) AnnotatedCommitFromRevspec(spec string) (*AnnotatedCommit, error) {
	crevspec := C.CString(spec)
	defer C.free(unsafe.Pointer(crevspec))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var ptr *C.git_annotated_commit
	ret := C.git_annotated_commit_from_revspec(&ptr, r.ptr, crevspec)
	runtime.KeepAlive(r)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	annotatedCommit := newAnnotatedCommitFromC(ptr, r)
	runtime.KeepAlive(r)
	return annotatedCommit, nil
}

type MergeTreeFlag int

const (
	// Detect renames that occur between the common ancestor and the "ours"
	// side or the common ancestor and the "theirs" side.  This will enable
	// the ability to merge between a modified and renamed file.
	MergeTreeFindRenames MergeTreeFlag = C.GIT_MERGE_FIND_RENAMES
	// If a conflict occurs, exit immediately instead of attempting to
	// continue resolving conflicts.  The merge operation will fail with
	// GIT_EMERGECONFLICT and no index will be returned.
	MergeTreeFailOnConflict MergeTreeFlag = C.GIT_MERGE_FAIL_ON_CONFLICT
	// MergeTreeSkipREUC specifies not to write the REUC extension on the
	// generated index.
	MergeTreeSkipREUC MergeTreeFlag = C.GIT_MERGE_SKIP_REUC
	// MergeTreeNoRecursive specifies not to build a recursive merge base (by
	// merging the multiple merge bases) if the commits being merged have
	// multiple merge bases. Instead, the first base is used.
	// This flag provides a similar merge base to `git-merge-resolve`.
	MergeTreeNoRecursive MergeTreeFlag = C.GIT_MERGE_NO_RECURSIVE
)

type MergeOptions struct {
	Version   uint
	TreeFlags MergeTreeFlag

	RenameThreshold uint
	TargetLimit     uint
	RecursionLimit  uint
	FileFavor       MergeFileFavor

	//TODO: Diff similarity metric
}

func mergeOptionsFromC(opts *C.git_merge_options) MergeOptions {
	return MergeOptions{
		Version:         uint(opts.version),
		TreeFlags:       MergeTreeFlag(opts.flags),
		RenameThreshold: uint(opts.rename_threshold),
		TargetLimit:     uint(opts.target_limit),
		RecursionLimit:  uint(opts.recursion_limit),
		FileFavor:       MergeFileFavor(opts.file_favor),
	}
}

func DefaultMergeOptions() (MergeOptions, error) {
	opts := C.git_merge_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_merge_options_init(&opts, C.GIT_MERGE_OPTIONS_VERSION)
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
		flags:            C.uint32_t(mo.TreeFlags),
		rename_threshold: C.uint(mo.RenameThreshold),
		target_limit:     C.uint(mo.TargetLimit),
		recursion_limit:  C.uint(mo.RecursionLimit),
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
	runtime.KeepAlive(theirHeads)
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
	runtime.KeepAlive(theirHeads)
	if err < 0 {
		return MergeAnalysisNone, MergePreferenceNone, MakeGitError(err)
	}
	return MergeAnalysis(analysis), MergePreference(preference), nil

}

func (r *Repository) MergeCommits(ours *Commit, theirs *Commit, options *MergeOptions) (*Index, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	copts := options.toC()

	var ptr *C.git_index
	ret := C.git_merge_commits(&ptr, r.ptr, ours.cast_ptr, theirs.cast_ptr, copts)
	runtime.KeepAlive(ours)
	runtime.KeepAlive(theirs)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newIndexFromC(ptr, r), nil
}

func (r *Repository) MergeTrees(ancestor *Tree, ours *Tree, theirs *Tree, options *MergeOptions) (*Index, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	copts := options.toC()

	var ancestor_ptr *C.git_tree
	if ancestor != nil {
		ancestor_ptr = ancestor.cast_ptr
	}
	var ptr *C.git_index
	ret := C.git_merge_trees(&ptr, r.ptr, ancestor_ptr, ours.cast_ptr, theirs.cast_ptr, copts)
	runtime.KeepAlive(ancestor)
	runtime.KeepAlive(ours)
	runtime.KeepAlive(theirs)
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newIndexFromC(ptr, r), nil
}

func (r *Repository) MergeBase(one *Oid, two *Oid) (*Oid, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var oid C.git_oid
	ret := C.git_merge_base(&oid, r.ptr, one.toC(), two.toC())
	runtime.KeepAlive(one)
	runtime.KeepAlive(two)
	runtime.KeepAlive(r)
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
	runtime.KeepAlive(one)
	runtime.KeepAlive(two)
	if ret < 0 {
		return make([]*Oid, 0), MakeGitError(ret)
	}

	oids := make([]*Oid, coids.count)
	hdr := reflect.SliceHeader{
		Data: uintptr(unsafe.Pointer(coids.ids)),
		Len:  int(coids.count),
		Cap:  int(coids.count),
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

type MergeFileFlags int

const (
	MergeFileDefault MergeFileFlags = C.GIT_MERGE_FILE_DEFAULT

	// Create standard conflicted merge files
	MergeFileStyleMerge MergeFileFlags = C.GIT_MERGE_FILE_STYLE_MERGE

	// Create diff3-style files
	MergeFileStyleDiff MergeFileFlags = C.GIT_MERGE_FILE_STYLE_DIFF3

	// Condense non-alphanumeric regions for simplified diff file
	MergeFileStyleSimplifyAlnum MergeFileFlags = C.GIT_MERGE_FILE_SIMPLIFY_ALNUM

	// Ignore all whitespace
	MergeFileIgnoreWhitespace MergeFileFlags = C.GIT_MERGE_FILE_IGNORE_WHITESPACE

	// Ignore changes in amount of whitespace
	MergeFileIgnoreWhitespaceChange MergeFileFlags = C.GIT_MERGE_FILE_IGNORE_WHITESPACE_CHANGE

	// Ignore whitespace at end of line
	MergeFileIgnoreWhitespaceEOL MergeFileFlags = C.GIT_MERGE_FILE_IGNORE_WHITESPACE_EOL

	// Use the "patience diff" algorithm
	MergeFileDiffPatience MergeFileFlags = C.GIT_MERGE_FILE_DIFF_PATIENCE

	// Take extra time to find minimal diff
	MergeFileDiffMinimal MergeFileFlags = C.GIT_MERGE_FILE_DIFF_MINIMAL
)

type MergeFileOptions struct {
	AncestorLabel string
	OurLabel      string
	TheirLabel    string
	Favor         MergeFileFavor
	Flags         MergeFileFlags
	MarkerSize    uint16
}

func mergeFileOptionsFromC(c C.git_merge_file_options) MergeFileOptions {
	return MergeFileOptions{
		AncestorLabel: C.GoString(c.ancestor_label),
		OurLabel:      C.GoString(c.our_label),
		TheirLabel:    C.GoString(c.their_label),
		Favor:         MergeFileFavor(c.favor),
		Flags:         MergeFileFlags(c.flags),
		MarkerSize:    uint16(c.marker_size),
	}
}

func populateCMergeFileOptions(c *C.git_merge_file_options, options MergeFileOptions) {
	c.ancestor_label = C.CString(options.AncestorLabel)
	c.our_label = C.CString(options.OurLabel)
	c.their_label = C.CString(options.TheirLabel)
	c.favor = C.git_merge_file_favor_t(options.Favor)
	c.flags = C.uint32_t(options.Flags)
	c.marker_size = C.ushort(options.MarkerSize)
}

func freeCMergeFileOptions(c *C.git_merge_file_options) {
	C.free(unsafe.Pointer(c.ancestor_label))
	C.free(unsafe.Pointer(c.our_label))
	C.free(unsafe.Pointer(c.their_label))
}

func MergeFile(ancestor MergeFileInput, ours MergeFileInput, theirs MergeFileInput, options *MergeFileOptions) (*MergeFileResult, error) {

	ancestorPath := C.CString(ancestor.Path)
	defer C.free(unsafe.Pointer(ancestorPath))
	var ancestorContents *byte
	if len(ancestor.Contents) > 0 {
		ancestorContents = &ancestor.Contents[0]
	}

	oursPath := C.CString(ours.Path)
	defer C.free(unsafe.Pointer(oursPath))
	var oursContents *byte
	if len(ours.Contents) > 0 {
		oursContents = &ours.Contents[0]
	}

	theirsPath := C.CString(theirs.Path)
	defer C.free(unsafe.Pointer(theirsPath))
	var theirsContents *byte
	if len(theirs.Contents) > 0 {
		theirsContents = &theirs.Contents[0]
	}

	var copts *C.git_merge_file_options
	if options != nil {
		copts = &C.git_merge_file_options{}
		ecode := C.git_merge_file_options_init(copts, C.GIT_MERGE_FILE_OPTIONS_VERSION)
		if ecode < 0 {
			return nil, MakeGitError(ecode)
		}
		populateCMergeFileOptions(copts, *options)
		defer freeCMergeFileOptions(copts)
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var result C.git_merge_file_result
	ecode := C._go_git_merge_file(&result,
		(*C.char)(unsafe.Pointer(ancestorContents)), C.size_t(len(ancestor.Contents)), ancestorPath, C.uint(ancestor.Mode),
		(*C.char)(unsafe.Pointer(oursContents)), C.size_t(len(ours.Contents)), oursPath, C.uint(ours.Mode),
		(*C.char)(unsafe.Pointer(theirsContents)), C.size_t(len(theirs.Contents)), theirsPath, C.uint(theirs.Mode),
		copts)
	runtime.KeepAlive(ancestor)
	runtime.KeepAlive(ours)
	runtime.KeepAlive(theirs)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newMergeFileResultFromC(&result), nil

}

// TODO: GIT_EXTERN(int) git_merge_file_from_index(git_merge_file_result *out,git_repository *repo,const git_index_entry *ancestor,	const git_index_entry *ours,	const git_index_entry *theirs,	const git_merge_file_options *opts);
