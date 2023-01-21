package git

/*
#include <git2.h>

extern void _go_git_populate_apply_callbacks(git_apply_options *options);
extern int _go_git_diff_foreach(git_diff *diff, int eachFile, int eachHunk, int eachLine, void *payload);
extern void _go_git_setup_diff_notify_callbacks(git_diff_options* opts);
extern int _go_git_diff_blobs(git_blob *old, const char *old_path, git_blob *new, const char *new_path, git_diff_options *opts, int eachFile, int eachHunk, int eachLine, void *payload);
extern int _go_git_diff_buffers(const void *old_buffer, size_t old_len, const char *old_as_path, const void *new_buffer, size_t new_len, const char *new_as_path, const git_diff_options *opts, int eachFile, int eachHunk, int eachLine, void *payload);

*/
import "C"
import (
	"errors"
	"runtime"
	"unsafe"
)

type DiffFlag uint32

const (
	DiffFlagBinary    DiffFlag = C.GIT_DIFF_FLAG_BINARY
	DiffFlagNotBinary DiffFlag = C.GIT_DIFF_FLAG_NOT_BINARY
	DiffFlagValidOid  DiffFlag = C.GIT_DIFF_FLAG_VALID_ID
	DiffFlagExists    DiffFlag = C.GIT_DIFF_FLAG_EXISTS
)

type Delta int

const (
	DeltaUnmodified Delta = C.GIT_DELTA_UNMODIFIED
	DeltaAdded      Delta = C.GIT_DELTA_ADDED
	DeltaDeleted    Delta = C.GIT_DELTA_DELETED
	DeltaModified   Delta = C.GIT_DELTA_MODIFIED
	DeltaRenamed    Delta = C.GIT_DELTA_RENAMED
	DeltaCopied     Delta = C.GIT_DELTA_COPIED
	DeltaIgnored    Delta = C.GIT_DELTA_IGNORED
	DeltaUntracked  Delta = C.GIT_DELTA_UNTRACKED
	DeltaTypeChange Delta = C.GIT_DELTA_TYPECHANGE
	DeltaUnreadable Delta = C.GIT_DELTA_UNREADABLE
	DeltaConflicted Delta = C.GIT_DELTA_CONFLICTED
)

//go:generate stringer -type Delta -trimprefix Delta -tags static

type DiffLineType int

const (
	DiffLineContext      DiffLineType = C.GIT_DIFF_LINE_CONTEXT
	DiffLineAddition     DiffLineType = C.GIT_DIFF_LINE_ADDITION
	DiffLineDeletion     DiffLineType = C.GIT_DIFF_LINE_DELETION
	DiffLineContextEOFNL DiffLineType = C.GIT_DIFF_LINE_CONTEXT_EOFNL
	DiffLineAddEOFNL     DiffLineType = C.GIT_DIFF_LINE_ADD_EOFNL
	DiffLineDelEOFNL     DiffLineType = C.GIT_DIFF_LINE_DEL_EOFNL

	DiffLineFileHdr DiffLineType = C.GIT_DIFF_LINE_FILE_HDR
	DiffLineHunkHdr DiffLineType = C.GIT_DIFF_LINE_HUNK_HDR
	DiffLineBinary  DiffLineType = C.GIT_DIFF_LINE_BINARY
)

//go:generate stringer -type DiffLineType -trimprefix DiffLine -tags static

type DiffFile struct {
	Path  string
	Oid   *Oid
	Size  int
	Flags DiffFlag
	Mode  uint16
}

func diffFileFromC(file *C.git_diff_file) DiffFile {
	return DiffFile{
		Path:  C.GoString(file.path),
		Oid:   newOidFromC(&file.id),
		Size:  int(file.size),
		Flags: DiffFlag(file.flags),
		Mode:  uint16(file.mode),
	}
}

type DiffDelta struct {
	Status     Delta
	Flags      DiffFlag
	Similarity uint16
	OldFile    DiffFile
	NewFile    DiffFile
}

func diffDeltaFromC(delta *C.git_diff_delta) DiffDelta {
	return DiffDelta{
		Status:     Delta(delta.status),
		Flags:      DiffFlag(delta.flags),
		Similarity: uint16(delta.similarity),
		OldFile:    diffFileFromC(&delta.old_file),
		NewFile:    diffFileFromC(&delta.new_file),
	}
}

type DiffHunk struct {
	OldStart int
	OldLines int
	NewStart int
	NewLines int
	Header   string
}

func diffHunkFromC(hunk *C.git_diff_hunk) DiffHunk {
	return DiffHunk{
		OldStart: int(hunk.old_start),
		OldLines: int(hunk.old_lines),
		NewStart: int(hunk.new_start),
		NewLines: int(hunk.new_lines),
		Header:   C.GoStringN(&hunk.header[0], C.int(hunk.header_len)),
	}
}

type DiffLine struct {
	Origin    DiffLineType
	OldLineno int
	NewLineno int
	NumLines  int
	Content   string
}

func diffLineFromC(line *C.git_diff_line) DiffLine {
	return DiffLine{
		Origin:    DiffLineType(line.origin),
		OldLineno: int(line.old_lineno),
		NewLineno: int(line.new_lineno),
		NumLines:  int(line.num_lines),
		Content:   C.GoStringN(line.content, C.int(line.content_len)),
	}
}

type Diff struct {
	doNotCompare
	ptr          *C.git_diff
	repo         *Repository
	runFinalizer bool
}

func (diff *Diff) NumDeltas() (int, error) {
	if diff.ptr == nil {
		return -1, ErrInvalid
	}
	ret := int(C.git_diff_num_deltas(diff.ptr))
	runtime.KeepAlive(diff)
	return ret, nil
}

func (diff *Diff) Delta(index int) (DiffDelta, error) {
	if diff.ptr == nil {
		return DiffDelta{}, ErrInvalid
	}
	ptr := C.git_diff_get_delta(diff.ptr, C.size_t(index))
	ret := diffDeltaFromC(ptr)
	runtime.KeepAlive(diff)
	return ret, nil
}

// deprecated: You should use `Diff.Delta()` instead.
func (diff *Diff) GetDelta(index int) (DiffDelta, error) {
	return diff.Delta(index)
}

func newDiffFromC(ptr *C.git_diff, repo *Repository) *Diff {
	if ptr == nil {
		return nil
	}

	diff := &Diff{
		ptr:          ptr,
		repo:         repo,
		runFinalizer: true,
	}

	runtime.SetFinalizer(diff, (*Diff).Free)
	return diff
}

func (diff *Diff) Free() error {
	if diff.ptr == nil {
		return ErrInvalid
	}
	if !diff.runFinalizer {
		// This is the case with the Diff objects that are involved in the DiffNotifyCallback.
		diff.ptr = nil
		return nil
	}
	runtime.SetFinalizer(diff, nil)
	C.git_diff_free(diff.ptr)
	diff.ptr = nil
	return nil
}

func (diff *Diff) FindSimilar(opts *DiffFindOptions) error {

	var copts *C.git_diff_find_options
	if opts != nil {
		copts = &C.git_diff_find_options{
			version:                       C.GIT_DIFF_FIND_OPTIONS_VERSION,
			flags:                         C.uint32_t(opts.Flags),
			rename_threshold:              C.uint16_t(opts.RenameThreshold),
			copy_threshold:                C.uint16_t(opts.CopyThreshold),
			rename_from_rewrite_threshold: C.uint16_t(opts.RenameFromRewriteThreshold),
			break_rewrite_threshold:       C.uint16_t(opts.BreakRewriteThreshold),
			rename_limit:                  C.size_t(opts.RenameLimit),
		}
	}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_diff_find_similar(diff.ptr, copts)
	runtime.KeepAlive(diff)
	if ecode < 0 {
		return MakeGitError(ecode)
	}

	return nil
}

type DiffStats struct {
	doNotCompare
	ptr *C.git_diff_stats
}

func (stats *DiffStats) Free() error {
	if stats.ptr == nil {
		return ErrInvalid
	}
	runtime.SetFinalizer(stats, nil)
	C.git_diff_stats_free(stats.ptr)
	stats.ptr = nil
	return nil
}

func (stats *DiffStats) Insertions() int {
	ret := int(C.git_diff_stats_insertions(stats.ptr))
	runtime.KeepAlive(stats)
	return ret
}

func (stats *DiffStats) Deletions() int {
	ret := int(C.git_diff_stats_deletions(stats.ptr))
	runtime.KeepAlive(stats)
	return ret
}

func (stats *DiffStats) FilesChanged() int {
	ret := int(C.git_diff_stats_files_changed(stats.ptr))
	runtime.KeepAlive(stats)
	return ret
}

type DiffStatsFormat int

const (
	DiffStatsNone           DiffStatsFormat = C.GIT_DIFF_STATS_NONE
	DiffStatsFull           DiffStatsFormat = C.GIT_DIFF_STATS_FULL
	DiffStatsShort          DiffStatsFormat = C.GIT_DIFF_STATS_SHORT
	DiffStatsNumber         DiffStatsFormat = C.GIT_DIFF_STATS_NUMBER
	DiffStatsIncludeSummary DiffStatsFormat = C.GIT_DIFF_STATS_INCLUDE_SUMMARY
)

func (stats *DiffStats) String(format DiffStatsFormat,
	width uint) (string, error) {
	buf := C.git_buf{}
	defer C.git_buf_dispose(&buf)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_diff_stats_to_buf(&buf,
		stats.ptr, C.git_diff_stats_format_t(format), C.size_t(width))
	runtime.KeepAlive(stats)
	if ret < 0 {
		return "", MakeGitError(ret)
	}

	return C.GoString(buf.ptr), nil
}

func (diff *Diff) Stats() (*DiffStats, error) {
	stats := new(DiffStats)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_diff_get_stats(&stats.ptr, diff.ptr)
	runtime.KeepAlive(diff)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	runtime.SetFinalizer(stats, (*DiffStats).Free)

	return stats, nil
}

type diffForEachCallbackData struct {
	fileCallback DiffForEachFileCallback
	hunkCallback DiffForEachHunkCallback
	lineCallback DiffForEachLineCallback
	errorTarget  *error
}

type DiffForEachFileCallback func(delta DiffDelta, progress float64) (DiffForEachHunkCallback, error)

type DiffDetail int

const (
	DiffDetailFiles DiffDetail = iota
	DiffDetailHunks
	DiffDetailLines
)

func (diff *Diff) ForEach(cbFile DiffForEachFileCallback, detail DiffDetail) error {
	if diff.ptr == nil {
		return ErrInvalid
	}

	intHunks := C.int(0)
	if detail >= DiffDetailHunks {
		intHunks = C.int(1)
	}

	intLines := C.int(0)
	if detail >= DiffDetailLines {
		intLines = C.int(1)
	}

	var err error
	data := &diffForEachCallbackData{
		fileCallback: cbFile,
		errorTarget:  &err,
	}

	handle := pointerHandles.Track(data)
	defer pointerHandles.Untrack(handle)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_diff_foreach(diff.ptr, 1, intHunks, intLines, handle)
	runtime.KeepAlive(diff)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

//export diffForEachFileCallback
func diffForEachFileCallback(delta *C.git_diff_delta, progress C.float, handle unsafe.Pointer) C.int {
	payload := pointerHandles.Get(handle)
	data, ok := payload.(*diffForEachCallbackData)
	if !ok {
		panic("could not retrieve data for handle")
	}

	data.hunkCallback = nil
	if data.fileCallback != nil {
		cb, err := data.fileCallback(diffDeltaFromC(delta), float64(progress))
		if err != nil {
			*data.errorTarget = err
			return C.int(ErrorCodeUser)
		}
		data.hunkCallback = cb
	}

	return C.int(ErrorCodeOK)
}

type DiffForEachHunkCallback func(DiffHunk) (DiffForEachLineCallback, error)

//export diffForEachHunkCallback
func diffForEachHunkCallback(delta *C.git_diff_delta, hunk *C.git_diff_hunk, handle unsafe.Pointer) C.int {
	payload := pointerHandles.Get(handle)
	data, ok := payload.(*diffForEachCallbackData)
	if !ok {
		panic("could not retrieve data for handle")
	}

	data.lineCallback = nil
	if data.hunkCallback != nil {
		cb, err := data.hunkCallback(diffHunkFromC(hunk))
		if err != nil {
			*data.errorTarget = err
			return C.int(ErrorCodeUser)
		}
		data.lineCallback = cb
	}

	return C.int(ErrorCodeOK)
}

type DiffForEachLineCallback func(DiffLine) error

//export diffForEachLineCallback
func diffForEachLineCallback(delta *C.git_diff_delta, hunk *C.git_diff_hunk, line *C.git_diff_line, handle unsafe.Pointer) C.int {
	payload := pointerHandles.Get(handle)
	data, ok := payload.(*diffForEachCallbackData)
	if !ok {
		panic("could not retrieve data for handle")
	}

	err := data.lineCallback(diffLineFromC(line))
	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}

	return C.int(ErrorCodeOK)
}

func (diff *Diff) Patch(deltaIndex int) (*Patch, error) {
	if diff.ptr == nil {
		return nil, ErrInvalid
	}
	var patchPtr *C.git_patch

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_patch_from_diff(&patchPtr, diff.ptr, C.size_t(deltaIndex))
	runtime.KeepAlive(diff)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}

	return newPatchFromC(patchPtr), nil
}

type DiffFormat int

const (
	DiffFormatPatch       DiffFormat = C.GIT_DIFF_FORMAT_PATCH
	DiffFormatPatchHeader DiffFormat = C.GIT_DIFF_FORMAT_PATCH_HEADER
	DiffFormatRaw         DiffFormat = C.GIT_DIFF_FORMAT_RAW
	DiffFormatNameOnly    DiffFormat = C.GIT_DIFF_FORMAT_NAME_ONLY
	DiffFormatNameStatus  DiffFormat = C.GIT_DIFF_FORMAT_NAME_STATUS
)

func (diff *Diff) ToBuf(format DiffFormat) ([]byte, error) {
	if diff.ptr == nil {
		return nil, ErrInvalid
	}

	diffBuf := C.git_buf{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_diff_to_buf(&diffBuf, diff.ptr, C.git_diff_format_t(format))
	runtime.KeepAlive(diff)
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	defer C.git_buf_dispose(&diffBuf)

	return C.GoBytes(unsafe.Pointer(diffBuf.ptr), C.int(diffBuf.size)), nil
}

type DiffOptionsFlag int

const (
	DiffNormal                 DiffOptionsFlag = C.GIT_DIFF_NORMAL
	DiffReverse                DiffOptionsFlag = C.GIT_DIFF_REVERSE
	DiffIncludeIgnored         DiffOptionsFlag = C.GIT_DIFF_INCLUDE_IGNORED
	DiffRecurseIgnoredDirs     DiffOptionsFlag = C.GIT_DIFF_RECURSE_IGNORED_DIRS
	DiffIncludeUntracked       DiffOptionsFlag = C.GIT_DIFF_INCLUDE_UNTRACKED
	DiffRecurseUntracked       DiffOptionsFlag = C.GIT_DIFF_RECURSE_UNTRACKED_DIRS
	DiffIncludeUnmodified      DiffOptionsFlag = C.GIT_DIFF_INCLUDE_UNMODIFIED
	DiffIncludeTypeChange      DiffOptionsFlag = C.GIT_DIFF_INCLUDE_TYPECHANGE
	DiffIncludeTypeChangeTrees DiffOptionsFlag = C.GIT_DIFF_INCLUDE_TYPECHANGE_TREES
	DiffIgnoreFilemode         DiffOptionsFlag = C.GIT_DIFF_IGNORE_FILEMODE
	DiffIgnoreSubmodules       DiffOptionsFlag = C.GIT_DIFF_IGNORE_SUBMODULES
	DiffIgnoreCase             DiffOptionsFlag = C.GIT_DIFF_IGNORE_CASE
	DiffIncludeCaseChange      DiffOptionsFlag = C.GIT_DIFF_INCLUDE_CASECHANGE

	DiffDisablePathspecMatch    DiffOptionsFlag = C.GIT_DIFF_DISABLE_PATHSPEC_MATCH
	DiffSkipBinaryCheck         DiffOptionsFlag = C.GIT_DIFF_SKIP_BINARY_CHECK
	DiffEnableFastUntrackedDirs DiffOptionsFlag = C.GIT_DIFF_ENABLE_FAST_UNTRACKED_DIRS

	DiffForceText   DiffOptionsFlag = C.GIT_DIFF_FORCE_TEXT
	DiffForceBinary DiffOptionsFlag = C.GIT_DIFF_FORCE_BINARY

	DiffIgnoreWhitespace       DiffOptionsFlag = C.GIT_DIFF_IGNORE_WHITESPACE
	DiffIgnoreWhitespaceChange DiffOptionsFlag = C.GIT_DIFF_IGNORE_WHITESPACE_CHANGE
	DiffIgnoreWhitespaceEOL    DiffOptionsFlag = C.GIT_DIFF_IGNORE_WHITESPACE_EOL

	DiffShowUntrackedContent DiffOptionsFlag = C.GIT_DIFF_SHOW_UNTRACKED_CONTENT
	DiffShowUnmodified       DiffOptionsFlag = C.GIT_DIFF_SHOW_UNMODIFIED
	DiffPatience             DiffOptionsFlag = C.GIT_DIFF_PATIENCE
	DiffMinimal              DiffOptionsFlag = C.GIT_DIFF_MINIMAL
	DiffShowBinary           DiffOptionsFlag = C.GIT_DIFF_SHOW_BINARY
	DiffIndentHeuristic      DiffOptionsFlag = C.GIT_DIFF_INDENT_HEURISTIC
)

type DiffNotifyCallback func(diffSoFar *Diff, deltaToAdd DiffDelta, matchedPathspec string) error

type DiffOptions struct {
	Flags            DiffOptionsFlag
	IgnoreSubmodules SubmoduleIgnore
	Pathspec         []string
	NotifyCallback   DiffNotifyCallback

	ContextLines   uint32
	InterhunkLines uint32
	IdAbbrev       uint16

	MaxSize int

	OldPrefix string
	NewPrefix string
}

func DefaultDiffOptions() (DiffOptions, error) {
	opts := C.git_diff_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_diff_options_init(&opts, C.GIT_DIFF_OPTIONS_VERSION)
	if ecode < 0 {
		return DiffOptions{}, MakeGitError(ecode)
	}

	return DiffOptions{
		Flags:            DiffOptionsFlag(opts.flags),
		IgnoreSubmodules: SubmoduleIgnore(opts.ignore_submodules),
		Pathspec:         makeStringsFromCStrings(opts.pathspec.strings, int(opts.pathspec.count)),
		ContextLines:     uint32(opts.context_lines),
		InterhunkLines:   uint32(opts.interhunk_lines),
		IdAbbrev:         uint16(opts.id_abbrev),
		MaxSize:          int(opts.max_size),
		OldPrefix:        "a",
		NewPrefix:        "b",
	}, nil
}

type DiffFindOptionsFlag int

const (
	DiffFindByConfig                    DiffFindOptionsFlag = C.GIT_DIFF_FIND_BY_CONFIG
	DiffFindRenames                     DiffFindOptionsFlag = C.GIT_DIFF_FIND_RENAMES
	DiffFindRenamesFromRewrites         DiffFindOptionsFlag = C.GIT_DIFF_FIND_RENAMES_FROM_REWRITES
	DiffFindCopies                      DiffFindOptionsFlag = C.GIT_DIFF_FIND_COPIES
	DiffFindCopiesFromUnmodified        DiffFindOptionsFlag = C.GIT_DIFF_FIND_COPIES_FROM_UNMODIFIED
	DiffFindRewrites                    DiffFindOptionsFlag = C.GIT_DIFF_FIND_REWRITES
	DiffFindBreakRewrites               DiffFindOptionsFlag = C.GIT_DIFF_BREAK_REWRITES
	DiffFindAndBreakRewrites            DiffFindOptionsFlag = C.GIT_DIFF_FIND_AND_BREAK_REWRITES
	DiffFindForUntracked                DiffFindOptionsFlag = C.GIT_DIFF_FIND_FOR_UNTRACKED
	DiffFindAll                         DiffFindOptionsFlag = C.GIT_DIFF_FIND_ALL
	DiffFindIgnoreLeadingWhitespace     DiffFindOptionsFlag = C.GIT_DIFF_FIND_IGNORE_LEADING_WHITESPACE
	DiffFindIgnoreWhitespace            DiffFindOptionsFlag = C.GIT_DIFF_FIND_IGNORE_WHITESPACE
	DiffFindDontIgnoreWhitespace        DiffFindOptionsFlag = C.GIT_DIFF_FIND_DONT_IGNORE_WHITESPACE
	DiffFindExactMatchOnly              DiffFindOptionsFlag = C.GIT_DIFF_FIND_EXACT_MATCH_ONLY
	DiffFindBreakRewritesForRenamesOnly DiffFindOptionsFlag = C.GIT_DIFF_BREAK_REWRITES_FOR_RENAMES_ONLY
	DiffFindRemoveUnmodified            DiffFindOptionsFlag = C.GIT_DIFF_FIND_REMOVE_UNMODIFIED
)

// TODO implement git_diff_similarity_metric
type DiffFindOptions struct {
	Flags                      DiffFindOptionsFlag
	RenameThreshold            uint16
	CopyThreshold              uint16
	RenameFromRewriteThreshold uint16
	BreakRewriteThreshold      uint16
	RenameLimit                uint
}

func DefaultDiffFindOptions() (DiffFindOptions, error) {
	opts := C.git_diff_find_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_diff_find_options_init(&opts, C.GIT_DIFF_FIND_OPTIONS_VERSION)
	if ecode < 0 {
		return DiffFindOptions{}, MakeGitError(ecode)
	}

	return DiffFindOptions{
		Flags:                      DiffFindOptionsFlag(opts.flags),
		RenameThreshold:            uint16(opts.rename_threshold),
		CopyThreshold:              uint16(opts.copy_threshold),
		RenameFromRewriteThreshold: uint16(opts.rename_from_rewrite_threshold),
		BreakRewriteThreshold:      uint16(opts.break_rewrite_threshold),
		RenameLimit:                uint(opts.rename_limit),
	}, nil
}

var (
	ErrDeltaSkip = errors.New("Skip delta")
)

type diffNotifyCallbackData struct {
	callback    DiffNotifyCallback
	repository  *Repository
	errorTarget *error
}

//export diffNotifyCallback
func diffNotifyCallback(_diff_so_far unsafe.Pointer, delta_to_add *C.git_diff_delta, matched_pathspec *C.char, handle unsafe.Pointer) C.int {
	diff_so_far := (*C.git_diff)(_diff_so_far)

	payload := pointerHandles.Get(handle)
	data, ok := payload.(*diffNotifyCallbackData)
	if !ok {
		panic("could not retrieve data for handle")
	}

	if data == nil {
		return C.int(ErrorCodeOK)
	}

	// We are not taking ownership of this diff pointer, so no finalizer is set.
	diff := &Diff{
		ptr:          diff_so_far,
		repo:         data.repository,
		runFinalizer: false,
	}

	err := data.callback(diff, diffDeltaFromC(delta_to_add), C.GoString(matched_pathspec))

	// Since the callback could theoretically keep a reference to the diff
	// (which could be freed by libgit2 if an error occurs later during the
	// diffing process), this converts a use-after-free (terrible!) into a nil
	// dereference ("just" pretty bad).
	diff.ptr = nil

	if err == ErrDeltaSkip {
		return 1
	}
	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}

	return C.int(ErrorCodeOK)
}

func populateDiffOptions(copts *C.git_diff_options, opts *DiffOptions, repo *Repository, errorTarget *error) *C.git_diff_options {
	C.git_diff_options_init(copts, C.GIT_DIFF_OPTIONS_VERSION)
	if opts == nil {
		return nil
	}

	copts.flags = C.uint32_t(opts.Flags)
	copts.ignore_submodules = C.git_submodule_ignore_t(opts.IgnoreSubmodules)
	if len(opts.Pathspec) > 0 {
		copts.pathspec.count = C.size_t(len(opts.Pathspec))
		copts.pathspec.strings = makeCStringsFromStrings(opts.Pathspec)
	}
	copts.context_lines = C.uint32_t(opts.ContextLines)
	copts.interhunk_lines = C.uint32_t(opts.InterhunkLines)
	copts.id_abbrev = C.uint16_t(opts.IdAbbrev)
	copts.max_size = C.git_off_t(opts.MaxSize)
	copts.old_prefix = C.CString(opts.OldPrefix)
	copts.new_prefix = C.CString(opts.NewPrefix)

	if opts.NotifyCallback != nil {
		notifyData := &diffNotifyCallbackData{
			callback:    opts.NotifyCallback,
			repository:  repo,
			errorTarget: errorTarget,
		}
		C._go_git_setup_diff_notify_callbacks(copts)
		copts.payload = pointerHandles.Track(notifyData)
	}
	return copts
}

func freeDiffOptions(copts *C.git_diff_options) {
	if copts == nil {
		return
	}
	freeStrarray(&copts.pathspec)
	C.free(unsafe.Pointer(copts.old_prefix))
	C.free(unsafe.Pointer(copts.new_prefix))
	if copts.payload != nil {
		pointerHandles.Untrack(copts.payload)
	}
}

func (v *Repository) DiffTreeToTree(oldTree, newTree *Tree, opts *DiffOptions) (*Diff, error) {
	var diffPtr *C.git_diff
	var oldPtr, newPtr *C.git_tree

	if oldTree != nil {
		oldPtr = oldTree.cast_ptr
	}

	if newTree != nil {
		newPtr = newTree.cast_ptr
	}

	var err error
	copts := populateDiffOptions(&C.git_diff_options{}, opts, v, &err)
	defer freeDiffOptions(copts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_diff_tree_to_tree(&diffPtr, v.ptr, oldPtr, newPtr, copts)
	runtime.KeepAlive(oldTree)
	runtime.KeepAlive(newTree)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return nil, err
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}
	return newDiffFromC(diffPtr, v), nil
}

func (v *Repository) DiffTreeToWorkdir(oldTree *Tree, opts *DiffOptions) (*Diff, error) {
	var diffPtr *C.git_diff
	var oldPtr *C.git_tree

	if oldTree != nil {
		oldPtr = oldTree.cast_ptr
	}

	var err error
	copts := populateDiffOptions(&C.git_diff_options{}, opts, v, &err)
	defer freeDiffOptions(copts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_diff_tree_to_workdir(&diffPtr, v.ptr, oldPtr, copts)
	runtime.KeepAlive(oldTree)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return nil, err
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newDiffFromC(diffPtr, v), nil
}

func (v *Repository) DiffTreeToIndex(oldTree *Tree, index *Index, opts *DiffOptions) (*Diff, error) {
	var diffPtr *C.git_diff
	var oldPtr *C.git_tree
	var indexPtr *C.git_index

	if oldTree != nil {
		oldPtr = oldTree.cast_ptr
	}

	if index != nil {
		indexPtr = index.ptr
	}

	var err error
	copts := populateDiffOptions(&C.git_diff_options{}, opts, v, &err)
	defer freeDiffOptions(copts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_diff_tree_to_index(&diffPtr, v.ptr, oldPtr, indexPtr, copts)
	runtime.KeepAlive(oldTree)
	runtime.KeepAlive(index)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return nil, err
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newDiffFromC(diffPtr, v), nil
}

func (v *Repository) DiffTreeToWorkdirWithIndex(oldTree *Tree, opts *DiffOptions) (*Diff, error) {
	var diffPtr *C.git_diff
	var oldPtr *C.git_tree

	if oldTree != nil {
		oldPtr = oldTree.cast_ptr
	}

	var err error
	copts := populateDiffOptions(&C.git_diff_options{}, opts, v, &err)
	defer freeDiffOptions(copts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_diff_tree_to_workdir_with_index(&diffPtr, v.ptr, oldPtr, copts)
	runtime.KeepAlive(oldTree)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return nil, err
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newDiffFromC(diffPtr, v), nil
}

func (v *Repository) DiffIndexToWorkdir(index *Index, opts *DiffOptions) (*Diff, error) {
	var diffPtr *C.git_diff
	var indexPtr *C.git_index

	if index != nil {
		indexPtr = index.ptr
	}

	var err error
	copts := populateDiffOptions(&C.git_diff_options{}, opts, v, &err)
	defer freeDiffOptions(copts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C.git_diff_index_to_workdir(&diffPtr, v.ptr, indexPtr, copts)
	runtime.KeepAlive(index)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return nil, err
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newDiffFromC(diffPtr, v), nil
}

// DiffBlobs performs a diff between two arbitrary blobs. You can pass
// whatever file names you'd like for them to appear as in the diff.
func DiffBlobs(oldBlob *Blob, oldAsPath string, newBlob *Blob, newAsPath string, opts *DiffOptions, fileCallback DiffForEachFileCallback, detail DiffDetail) error {
	var err error
	data := &diffForEachCallbackData{
		fileCallback: fileCallback,
		errorTarget:  &err,
	}

	intHunks := C.int(0)
	if detail >= DiffDetailHunks {
		intHunks = C.int(1)
	}

	intLines := C.int(0)
	if detail >= DiffDetailLines {
		intLines = C.int(1)
	}

	handle := pointerHandles.Track(data)
	defer pointerHandles.Untrack(handle)

	var repo *Repository
	var oldBlobPtr, newBlobPtr *C.git_blob
	if oldBlob != nil {
		oldBlobPtr = oldBlob.cast_ptr
		repo = oldBlob.repo
	}
	if newBlob != nil {
		newBlobPtr = newBlob.cast_ptr
		repo = newBlob.repo
	}

	oldBlobPath := C.CString(oldAsPath)
	defer C.free(unsafe.Pointer(oldBlobPath))
	newBlobPath := C.CString(newAsPath)
	defer C.free(unsafe.Pointer(newBlobPath))

	copts := populateDiffOptions(&C.git_diff_options{}, opts, repo, &err)
	defer freeDiffOptions(copts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_diff_blobs(oldBlobPtr, oldBlobPath, newBlobPtr, newBlobPath, copts, 1, intHunks, intLines, handle)
	runtime.KeepAlive(oldBlob)
	runtime.KeepAlive(newBlob)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// DiffBuffers performs a diff between two arbitrary buffers. You can pass
// whatever file names you'd like for them to appear as in the diff.
func DiffBuffers(oldBuffer []byte, oldAsPath string, newBufer []byte, newAsPath string, opts *DiffOptions, fileCallback DiffForEachFileCallback, detail DiffDetail) error {
	var err error
	data := &diffForEachCallbackData{
		fileCallback: fileCallback,
		errorTarget:  &err,
	}

	intHunks := C.int(0)
	if detail >= DiffDetailHunks {
		intHunks = C.int(1)
	}

	intLines := C.int(0)
	if detail >= DiffDetailLines {
		intLines = C.int(1)
	}

	handle := pointerHandles.Track(data)
	defer pointerHandles.Untrack(handle)

	cOldBuffer := C.CBytes(oldBuffer)
	defer C.free(unsafe.Pointer(cOldBuffer))
	cNewBuffer := C.CBytes(newBufer)
	defer C.free(unsafe.Pointer(cNewBuffer))

	cOldLen := C.size_t(len(oldBuffer))
	cNewLen := C.size_t(len(newBufer))

	oldPath := C.CString(oldAsPath)
	defer C.free(unsafe.Pointer(oldPath))
	newPath := C.CString(newAsPath)
	defer C.free(unsafe.Pointer(newPath))

	copts := populateDiffOptions(&C.git_diff_options{}, opts, nil, &err)
	defer freeDiffOptions(copts)

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ret := C._go_git_diff_buffers(cOldBuffer, cOldLen, oldPath, cNewBuffer, cNewLen, newPath, copts, 1, intHunks, intLines, handle)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// ApplyHunkCallback is a callback that will be made per delta (file) when applying a patch.
type ApplyHunkCallback func(*DiffHunk) (apply bool, err error)

// ApplyDeltaCallback is a callback that will be made per hunk when applying a patch.
type ApplyDeltaCallback func(*DiffDelta) (apply bool, err error)

// ApplyOptions has 2 callbacks that are called for hunks or deltas
// If these functions return an error, abort the apply process immediately.
// If the first return value is true, the delta/hunk will be applied. If it is false, the  delta/hunk will not be applied. In either case, the rest of the apply process will continue.
type ApplyOptions struct {
	ApplyHunkCallback  ApplyHunkCallback
	ApplyDeltaCallback ApplyDeltaCallback
	Flags              uint
}

type applyCallbackData struct {
	options     *ApplyOptions
	errorTarget *error
}

//export hunkApplyCallback
func hunkApplyCallback(_hunk *C.git_diff_hunk, _payload unsafe.Pointer) C.int {
	data, ok := pointerHandles.Get(_payload).(*applyCallbackData)
	if !ok {
		panic("invalid apply options payload")
	}

	if data.options.ApplyHunkCallback == nil {
		return C.int(ErrorCodeOK)
	}

	hunk := diffHunkFromC(_hunk)

	apply, err := data.options.ApplyHunkCallback(&hunk)
	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}

	if !apply {
		return 1
	}
	return C.int(ErrorCodeOK)
}

//export deltaApplyCallback
func deltaApplyCallback(_delta *C.git_diff_delta, _payload unsafe.Pointer) C.int {
	data, ok := pointerHandles.Get(_payload).(*applyCallbackData)
	if !ok {
		panic("invalid apply options payload")
	}

	if data.options.ApplyDeltaCallback == nil {
		return C.int(ErrorCodeOK)
	}

	delta := diffDeltaFromC(_delta)

	apply, err := data.options.ApplyDeltaCallback(&delta)
	if err != nil {
		*data.errorTarget = err
		return C.int(ErrorCodeUser)
	}

	if !apply {
		return 1
	}
	return C.int(ErrorCodeOK)
}

// DefaultApplyOptions returns default options for applying diffs or patches.
func DefaultApplyOptions() (*ApplyOptions, error) {
	opts := C.git_apply_options{}

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_apply_options_init(&opts, C.GIT_APPLY_OPTIONS_VERSION)
	if int(ecode) != 0 {
		return nil, MakeGitError(ecode)
	}

	return applyOptionsFromC(&opts), nil
}

func populateApplyOptions(copts *C.git_apply_options, opts *ApplyOptions, errorTarget *error) *C.git_apply_options {
	C.git_apply_options_init(copts, C.GIT_APPLY_OPTIONS_VERSION)
	if opts == nil {
		return nil
	}

	copts.flags = C.uint(opts.Flags)

	if opts.ApplyDeltaCallback != nil || opts.ApplyHunkCallback != nil {
		data := &applyCallbackData{
			options:     opts,
			errorTarget: errorTarget,
		}
		C._go_git_populate_apply_callbacks(copts)
		copts.payload = pointerHandles.Track(data)
	}

	return copts
}

func freeApplyOptions(copts *C.git_apply_options) {
	if copts == nil {
		return
	}
	if copts.payload != nil {
		pointerHandles.Untrack(copts.payload)
	}
}

func applyOptionsFromC(copts *C.git_apply_options) *ApplyOptions {
	return &ApplyOptions{
		Flags: uint(copts.flags),
	}
}

// ApplyLocation represents the possible application locations for applying
// diffs.
type ApplyLocation int

const (
	// ApplyLocationWorkdir applies the patch to the workdir, leaving the
	// index untouched. This is the equivalent of `git apply` with no location
	// argument.
	ApplyLocationWorkdir ApplyLocation = C.GIT_APPLY_LOCATION_WORKDIR
	// ApplyLocationIndex applies the patch to the index, leaving the working
	// directory untouched. This is the equivalent of `git apply --cached`.
	ApplyLocationIndex ApplyLocation = C.GIT_APPLY_LOCATION_INDEX
	// ApplyLocationBoth applies the patch to both the working directory and
	// the index. This is the equivalent of `git apply --index`.
	ApplyLocationBoth ApplyLocation = C.GIT_APPLY_LOCATION_BOTH
)

// ApplyDiff appllies a Diff to the given repository, making changes directly
// in the working directory, the index, or both.
func (v *Repository) ApplyDiff(diff *Diff, location ApplyLocation, opts *ApplyOptions) error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	cOpts := populateApplyOptions(&C.git_apply_options{}, opts, &err)
	defer freeApplyOptions(cOpts)

	ret := C.git_apply(v.ptr, diff.ptr, C.git_apply_location_t(location), cOpts)
	runtime.KeepAlive(v)
	runtime.KeepAlive(diff)
	runtime.KeepAlive(cOpts)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return err
	}
	if ret < 0 {
		return MakeGitError(ret)
	}

	return nil
}

// ApplyToTree applies a Diff to a Tree and returns the resulting image as an Index.
func (v *Repository) ApplyToTree(diff *Diff, tree *Tree, opts *ApplyOptions) (*Index, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var err error
	cOpts := populateApplyOptions(&C.git_apply_options{}, opts, &err)
	defer freeApplyOptions(cOpts)

	var indexPtr *C.git_index
	ret := C.git_apply_to_tree(&indexPtr, v.ptr, tree.cast_ptr, diff.ptr, cOpts)
	runtime.KeepAlive(diff)
	runtime.KeepAlive(tree)
	runtime.KeepAlive(cOpts)
	if ret == C.int(ErrorCodeUser) && err != nil {
		return nil, err
	}
	if ret < 0 {
		return nil, MakeGitError(ret)
	}

	return newIndexFromC(indexPtr, v), nil
}

// DiffFromBuffer reads the contents of a git patch file into a Diff object.
//
// The diff object produced is similar to the one that would be produced if you
// actually produced it computationally by comparing two trees, however there
// may be subtle differences. For example, a patch file likely contains
// abbreviated object IDs, so the object IDs in a git_diff_delta produced by
// this function will also be abbreviated.
//
// This function will only read patch files created by a git implementation, it
// will not read unified diffs produced by the diff program, nor any other
// types of patch files.
func DiffFromBuffer(buffer []byte, repo *Repository) (*Diff, error) {
	var diff *C.git_diff

	cBuffer := C.CBytes(buffer)
	defer C.free(unsafe.Pointer(cBuffer))

	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ecode := C.git_diff_from_buffer(&diff, (*C.char)(cBuffer), C.size_t(len(buffer)))
	if ecode < 0 {
		return nil, MakeGitError(ecode)
	}
	runtime.KeepAlive(diff)

	return newDiffFromC(diff, repo), nil
}
