#include "_cgo_export.h"

#include <git2.h>
#include <git2/sys/odb_backend.h>
#include <git2/sys/refdb_backend.h>
#include <git2/sys/cred.h>

// There are two ways in which to declare a callback:
//
// * If there is a guarantee that the callback will always be called within the
//   same stack (e.g. by passing the callback directly into a function / into a
//   struct that goes into a function), the following pattern is preferred,
//   which preserves the error object as-is:
//
//   // myfile.go
//   type FooCallback func(...) (..., error)
//   type fooCallbackData struct {
//     callback    FooCallback
//     errorTarget *error
//   }
//
//   //export fooCallback
//   func fooCallback(..., handle unsafe.Pointer) C.int {
//     payload := pointerHandles.Get(handle)
//     data := payload.(*fooCallbackData)
//     ...
//     err := data.callback(...)
//     if err != nil {
//       *data.errorTarget = err
//       return C.int(ErrorCodeUser)
//     }
//     return C.int(ErrorCodeOK)
//   }
//
//   func MyFunction(... callback FooCallback) error {
//    var err error
//    data := fooCallbackData{
//      callback:    callback,
//      errorTarget: &err,
//    }
//    handle := pointerHandles.Track(&data)
//    defer pointerHandles.Untrack(handle)
//
//    runtime.LockOSThread()
//    defer runtime.UnlockOSThread()
//
//    ret := C._go_git_my_function(..., handle)
//    if ret == C.int(ErrorCodeUser) && err != nil {
//      return err
//    }
//    if ret < 0 {
//      return MakeGitError(ret)
//    }
//    return nil
//   }
//
//   // wrapper.c
//   int _go_git_my_function(..., void *payload)
//   {
//     return git_my_function(..., (git_foo_cb)&fooCallback, payload);
//   }
//
// * Additionally, if the same callback can be invoked from multiple functions or
//   from different stacks (e.g. when passing the callback to an object), the
//   following pattern should be used in tandem, which has the downside of
//   losing the original error object and converting it to a GitError if the
//   callback happens from a different stack:
//
//   // myfile.go
//   type FooCallback func(...) (..., error)
//   type fooCallbackData struct {
//     callback    FooCallback
//     errorTarget *error
//   }
//
//   //export fooCallback
//   func fooCallback(errorMessage **C.char, ..., handle unsafe.Pointer) C.int {
//     data := pointerHandles.Get(data).(*fooCallbackData)
//     ...
//     err := data.callback(...)
//     if err != nil {
//       if data.errorTarget != nil {
//         *data.errorTarget = err
//       }
//       return setCallbackError(errorMessage, err)
//     }
//     return C.int(ErrorCodeOK)
//   }
//
//   // wrapper.c
//   static int foo_callback(...)
//   {
//     char *error_message = NULL;
//     const int ret = fooCallback(&error_message, ...);
//     return set_callback_error(error_message, ret);
//   }

/**
 * Sets the thread-local error to the provided string. This needs to happen in
 * C because Go might change Goroutines _just_ before returning, which would
 * lose the contents of the error message.
 */
static int set_callback_error(char *error_message, int ret)
{
	if (error_message != NULL) {
		if (ret < 0)
			git_error_set_str(GIT_ERROR_CALLBACK, error_message);
		free(error_message);
	}
	return ret;
}

void _go_git_populate_apply_callbacks(git_apply_options *options)
{
	options->delta_cb = (git_apply_delta_cb)&deltaApplyCallback;
	options->hunk_cb = (git_apply_hunk_cb)&hunkApplyCallback;
}

static int commit_create_callback(
		git_oid *out,
		const git_signature *author,
		const git_signature *committer,
		const char *message_encoding,
		const char *message,
		const git_tree *tree,
		size_t parent_count,
		const git_commit *parents[],
		void *payload)
{
	char *error_message = NULL;
	const int ret = commitCreateCallback(
			&error_message,
			out,
			(git_signature *)author,
			(git_signature *)committer,
			(char *)message_encoding,
			(char *)message,
			(git_tree *)tree,
			parent_count,
			(git_commit **)parents,
			payload
	);
	return set_callback_error(error_message, ret);
}

void _go_git_populate_rebase_callbacks(git_rebase_options *opts)
{
	opts->commit_create_cb = commit_create_callback;
}

void _go_git_populate_clone_callbacks(git_clone_options *opts)
{
	opts->remote_cb = (git_remote_create_cb)&remoteCreateCallback;
}

void _go_git_populate_checkout_callbacks(git_checkout_options *opts)
{
	opts->notify_cb = (git_checkout_notify_cb)&checkoutNotifyCallback;
	opts->progress_cb = (git_checkout_progress_cb)&checkoutProgressCallback;
}

int _go_git_visit_submodule(git_repository *repo, void *fct)
{
	return git_submodule_foreach(repo, (git_submodule_cb)&submoduleCallback, fct);
}

int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr)
{
	return git_tree_walk(tree, mode, (git_treewalk_cb)&treeWalkCallback, ptr);
}

int _go_git_packbuilder_foreach(git_packbuilder *pb, void *payload)
{
	return git_packbuilder_foreach(pb, (git_packbuilder_foreach_cb)&packbuilderForEachCallback, payload);
}

int _go_git_odb_foreach(git_odb *db, void *payload)
{
	return git_odb_foreach(db, (git_odb_foreach_cb)&odbForEachCallback, payload);
}

void _go_git_odb_backend_free(git_odb_backend *backend)
{
	if (!backend->free)
		return;
	backend->free(backend);
}

void _go_git_refdb_backend_free(git_refdb_backend *backend)
{
	if (!backend->free)
		return;
	backend->free(backend);
}

int _go_git_diff_foreach(git_diff *diff, int eachFile, int eachHunk, int eachLine, void *payload)
{
	git_diff_file_cb fcb = NULL;
	git_diff_hunk_cb hcb = NULL;
	git_diff_line_cb lcb = NULL;

	if (eachFile)
		fcb = (git_diff_file_cb)&diffForEachFileCallback;
	if (eachHunk)
		hcb = (git_diff_hunk_cb)&diffForEachHunkCallback;
	if (eachLine)
		lcb = (git_diff_line_cb)&diffForEachLineCallback;

	return git_diff_foreach(diff, fcb, NULL, hcb, lcb, payload);
}

int _go_git_diff_blobs(
		git_blob *old,
		const char *old_path,
		git_blob *new,
		const char *new_path,
		git_diff_options *opts,
		int eachFile,
		int eachHunk,
		int eachLine,
		void *payload)
{
	git_diff_file_cb fcb = NULL;
	git_diff_hunk_cb hcb = NULL;
	git_diff_line_cb lcb = NULL;

	if (eachFile)
		fcb = (git_diff_file_cb)&diffForEachFileCallback;
	if (eachHunk)
		hcb = (git_diff_hunk_cb)&diffForEachHunkCallback;
	if (eachLine)
		lcb = (git_diff_line_cb)&diffForEachLineCallback;

	return git_diff_blobs(old, old_path, new, new_path, opts, fcb, NULL, hcb, lcb, payload);
}

int _go_git_diff_buffers(
		const void *old_buffer,
		size_t old_len,
		const char *old_as_path,
		const void *new_buffer,
		size_t new_len,
		const char *new_as_path,
		const git_diff_options *opts,
		int eachFile,
		int eachHunk,
		int eachLine,
		void *payload)
{
	git_diff_file_cb fcb = NULL;
	git_diff_hunk_cb hcb = NULL;
	git_diff_line_cb lcb = NULL;

	if (eachFile)
		fcb = (git_diff_file_cb)&diffForEachFileCallback;
	if (eachHunk)
		hcb = (git_diff_hunk_cb)&diffForEachHunkCallback;
	if (eachLine)
		lcb = (git_diff_line_cb)&diffForEachLineCallback;

	return git_diff_buffers(old_buffer, old_len, old_as_path, new_buffer, new_len, new_as_path, opts, fcb, NULL, hcb, lcb, payload);
}


void _go_git_setup_diff_notify_callbacks(git_diff_options *opts)
{
	opts->notify_cb = (git_diff_notify_cb)&diffNotifyCallback;
}

static int sideband_progress_callback(const char *str, int len, void *payload)
{
	char *error_message = NULL;
	const int ret = sidebandProgressCallback(&error_message, (char *)str, len, payload);
	return set_callback_error(error_message, ret);
}

static int completion_callback(git_remote_completion_type completion_type, void *data)
{
	char *error_message = NULL;
	const int ret = completionCallback(&error_message, completion_type, data);
	return set_callback_error(error_message, ret);
}

static int credentials_callback(
		git_credential **cred,
		const char *url,
		const char *username_from_url,
		unsigned int allowed_types,
		void *data)
{
	char *error_message = NULL;
	const int ret = credentialsCallback(
			&error_message,
			cred,
			(char *)url,
			(char *)username_from_url,
			allowed_types,
			data
	);
	return set_callback_error(error_message, ret);
}

static int transfer_progress_callback(const git_transfer_progress *stats, void *data)
{
	char *error_message = NULL;
	const int ret = transferProgressCallback(
			&error_message,
			(git_transfer_progress *)stats,
			data
	);
	return set_callback_error(error_message, ret);
}

static int update_tips_callback(const char *refname, const git_oid *a, const git_oid *b, void *data)
{
	char *error_message = NULL;
	const int ret = updateTipsCallback(
			&error_message,
			(char *)refname,
			(git_oid *)a,
			(git_oid *)b,
			data
	);
	return set_callback_error(error_message, ret);
}

static int certificate_check_callback(git_cert *cert, int valid, const char *host, void *data)
{
	char *error_message = NULL;
	const int ret = certificateCheckCallback(
			&error_message,
			cert,
			valid,
			(char *)host,
			data
	);
	return set_callback_error(error_message, ret);
}

static int pack_progress_callback(int stage, unsigned int current, unsigned int total, void *data)
{
	char *error_message = NULL;
	const int ret = packProgressCallback(
			&error_message,
			stage,
			current,
			total,
			data
	);
	return set_callback_error(error_message, ret);
}

static int push_transfer_progress_callback(
		unsigned int current,
		unsigned int total,
		size_t bytes,
		void *data)
{
	char *error_message = NULL;
	const int ret = pushTransferProgressCallback(
			&error_message,
			current,
			total,
			bytes,
			data
	);
	return set_callback_error(error_message, ret);
}

static int push_update_reference_callback(const char *refname, const char *status, void *data)
{
	char *error_message = NULL;
	const int ret = pushUpdateReferenceCallback(
			&error_message,
			(char *)refname,
			(char *)status,
			data
	);
	return set_callback_error(error_message, ret);
}

void _go_git_populate_remote_callbacks(git_remote_callbacks *callbacks)
{
	callbacks->sideband_progress = sideband_progress_callback;
	callbacks->completion = completion_callback;
	callbacks->credentials = credentials_callback;
	callbacks->transfer_progress = transfer_progress_callback;
	callbacks->update_tips = update_tips_callback;
	callbacks->certificate_check = certificate_check_callback;
	callbacks->pack_progress = pack_progress_callback;
	callbacks->push_transfer_progress = push_transfer_progress_callback;
	callbacks->push_update_reference = push_update_reference_callback;
}

int _go_git_index_add_all(git_index *index, const git_strarray *pathspec, unsigned int flags, void *callback)
{
	git_index_matched_path_cb cb = callback ? (git_index_matched_path_cb)&indexMatchedPathCallback : NULL;
	return git_index_add_all(index, pathspec, flags, cb, callback);
}

int _go_git_index_update_all(git_index *index, const git_strarray *pathspec, void *callback)
{
	git_index_matched_path_cb cb = callback ? (git_index_matched_path_cb)&indexMatchedPathCallback : NULL;
	return git_index_update_all(index, pathspec, cb, callback);
}

int _go_git_index_remove_all(git_index *index, const git_strarray *pathspec, void *callback)
{
	git_index_matched_path_cb cb = callback ? (git_index_matched_path_cb)&indexMatchedPathCallback : NULL;
	return git_index_remove_all(index, pathspec, cb, callback);
}

int _go_git_tag_foreach(git_repository *repo, void *payload)
{
	return git_tag_foreach(repo, (git_tag_foreach_cb)&tagForeachCallback, payload);
}

int _go_git_merge_file(
		git_merge_file_result* out,
		char* ancestorContents,
		size_t ancestorLen,
		char* ancestorPath,
		unsigned int ancestorMode,
		char* oursContents,
		size_t oursLen,
		char* oursPath,
		unsigned int oursMode,
		char* theirsContents,
		size_t theirsLen,
		char* theirsPath,
		unsigned int theirsMode,
		git_merge_file_options* copts)
{
	git_merge_file_input ancestor = GIT_MERGE_FILE_INPUT_INIT;
	git_merge_file_input ours = GIT_MERGE_FILE_INPUT_INIT;
	git_merge_file_input theirs = GIT_MERGE_FILE_INPUT_INIT;

	ancestor.ptr = ancestorContents;
	ancestor.size = ancestorLen;
	ancestor.path = ancestorPath;
	ancestor.mode = ancestorMode;

	ours.ptr = oursContents;
	ours.size = oursLen;
	ours.path = oursPath;
	ours.mode = oursMode;

	theirs.ptr = theirsContents;
	theirs.size = theirsLen;
	theirs.path = theirsPath;
	theirs.mode = theirsMode;

	return git_merge_file(out, &ancestor, &ours, &theirs, copts);
}

void _go_git_populate_stash_apply_callbacks(git_stash_apply_options *opts)
{
	opts->progress_cb = (git_stash_apply_progress_cb)&stashApplyProgressCallback;
}

int _go_git_stash_foreach(git_repository *repo, void *payload)
{
	return git_stash_foreach(repo, (git_stash_cb)&stashForeachCallback, payload);
}

int _go_git_writestream_write(git_writestream *stream, const char *buffer, size_t len)
{
	return stream->write(stream, buffer, len);
}

int _go_git_writestream_close(git_writestream *stream)
{
	return stream->close(stream);
}

void _go_git_writestream_free(git_writestream *stream)
{
	stream->free(stream);
}

git_credential_t _go_git_credential_credtype(git_credential *cred)
{
	return cred->credtype;
}

static int credential_ssh_sign_callback(
		LIBSSH2_SESSION *session,
		unsigned char **sig, size_t *sig_len,
		const unsigned char *data, size_t data_len,
		void **abstract)
{
	char *error_message = NULL;
	const int ret = credentialSSHSignCallback(
			&error_message,
			sig,
			sig_len,
			(unsigned char *)data,
			data_len,
			(void *)*(uintptr_t *)abstract);
	return set_callback_error(error_message, ret);
}

void _go_git_populate_credential_ssh_custom(git_credential_ssh_custom *cred)
{
	cred->parent.free = (void (*)(git_credential *))credentialSSHCustomFree;
	cred->sign_callback = credential_ssh_sign_callback;
}

int _go_git_odb_write_pack(git_odb_writepack **out, git_odb *db, void *progress_payload)
{
	return git_odb_write_pack(out, db, transfer_progress_callback, progress_payload);
}

int _go_git_odb_writepack_append(
		git_odb_writepack *writepack,
		const void *data,
		size_t size,
		git_transfer_progress *stats)
{
	return writepack->append(writepack, data, size, stats);
}

int _go_git_odb_writepack_commit(git_odb_writepack *writepack, git_transfer_progress *stats)
{
	return writepack->commit(writepack, stats);
}

void _go_git_odb_writepack_free(git_odb_writepack *writepack)
{
	writepack->free(writepack);
}

int _go_git_indexer_new(
		git_indexer **out,
		const char *path,
		unsigned int mode,
		git_odb *odb,
		void *progress_cb_payload)
{
	git_indexer_options indexer_options = GIT_INDEXER_OPTIONS_INIT;
	indexer_options.progress_cb = transfer_progress_callback;
	indexer_options.progress_cb_payload = progress_cb_payload;
	return git_indexer_new(out, path, mode, odb, &indexer_options);
}

static int smart_transport_callback(
		git_transport **out,
		git_remote *owner,
		void *param)
{
	char *error_message = NULL;
	const int ret = smartTransportCallback(
			&error_message,
			out,
			owner,
			param);
	return set_callback_error(error_message, ret);
}

int _go_git_transport_register(const char *prefix, void *param)
{
	return git_transport_register(prefix, smart_transport_callback, param);
}

static int smart_subtransport_action_callback(
		git_smart_subtransport_stream **out,
		git_smart_subtransport *transport,
		const char *url,
		git_smart_service_t action)
{
	char *error_message = NULL;
	const int ret = smartSubtransportActionCallback(
			&error_message,
			out,
			transport,
			(char *)url,
			action);
	return set_callback_error(error_message, ret);
}

static int smart_subtransport_close_callback(git_smart_subtransport *transport)
{
	char *error_message = NULL;
	const int ret = smartSubtransportCloseCallback(
			&error_message,
			transport);
	return set_callback_error(error_message, ret);
}

static int smart_subtransport_callback(
		git_smart_subtransport **out,
		git_transport *owner,
		void *param)
{
	_go_managed_smart_subtransport *subtransport = (_go_managed_smart_subtransport *)param;
	subtransport->parent.action = smart_subtransport_action_callback;
	subtransport->parent.close = smart_subtransport_close_callback;
	subtransport->parent.free = smartSubtransportFreeCallback;

	*out = &subtransport->parent;
	char *error_message = NULL;
	const int ret = smartTransportSubtransportCallback(&error_message, subtransport, owner);
	return set_callback_error(error_message, ret);
}

int _go_git_transport_smart(
		git_transport **out,
		git_remote *owner,
		int stateless,
		_go_managed_smart_subtransport *subtransport_payload)
{
	git_smart_subtransport_definition definition = {
		smart_subtransport_callback,
		stateless,
		subtransport_payload,
	};

	return git_transport_smart(out, owner, &definition);
}

static int smart_subtransport_stream_read_callback(
		git_smart_subtransport_stream *stream,
		char *buffer,
		size_t buf_size,
		size_t *bytes_read)
{
	char *error_message = NULL;
	const int ret = smartSubtransportStreamReadCallback(
			&error_message,
			stream,
			buffer,
			buf_size,
			bytes_read);
	return set_callback_error(error_message, ret);
}

static int smart_subtransport_stream_write_callback(
		git_smart_subtransport_stream *stream,
		const char *buffer,
		size_t len)
{
	char *error_message = NULL;
	const int ret = smartSubtransportStreamWriteCallback(
			&error_message,
			stream,
			(char *)buffer,
			len);
	return set_callback_error(error_message, ret);
}

void _go_git_setup_smart_subtransport_stream(_go_managed_smart_subtransport_stream *stream)
{
	_go_managed_smart_subtransport_stream *managed_stream = (_go_managed_smart_subtransport_stream *)stream;
	managed_stream->parent.read = smart_subtransport_stream_read_callback;
	managed_stream->parent.write = smart_subtransport_stream_write_callback;
	managed_stream->parent.free = smartSubtransportStreamFreeCallback;
}
