#include "_cgo_export.h"
#include "git2.h"
#include "git2/sys/odb_backend.h"
#include "git2/sys/refdb_backend.h"
#include "git2/submodule.h"
#include "git2/pack.h"

typedef int (*gogit_submodule_cbk)(git_submodule *sm, const char *name, void *payload);

int _go_git_visit_submodule(git_repository *repo, void *fct)
{
	  return git_submodule_foreach(repo, (gogit_submodule_cbk)&SubmoduleVisitor, fct);
}

int _go_git_treewalk(git_tree *tree, git_treewalk_mode mode, void *ptr)
{
	return git_tree_walk(tree, mode, (git_treewalk_cb)&CallbackGitTreeWalk, ptr);
}

int _go_git_packbuilder_foreach(git_packbuilder *pb, void *payload)
{
    return git_packbuilder_foreach(pb, (git_packbuilder_foreach_cb)&packbuilderForEachCb, payload);
}

int _go_git_odb_foreach(git_odb *db, void *payload)
{
    return git_odb_foreach(db, (git_odb_foreach_cb)&odbForEachCb, payload);
}

void _go_git_odb_backend_free(git_odb_backend *backend)
{
    if (backend->free)
      backend->free(backend);

    return;
}

void _go_git_refdb_backend_free(git_refdb_backend *backend)
{
    if (backend->free)
      backend->free(backend);

    return;
}

void _go_git_setup_callbacks(git_remote_callbacks *callbacks) {
	typedef int (*completion_cb)(git_remote_completion_type type, void *data);
	typedef int (*credentials_cb)(git_cred **cred, const char *url, const char *username_from_url, unsigned int allowed_types,	void *data);
	typedef int (*transfer_progress_cb)(const git_transfer_progress *stats, void *data);
	typedef int (*update_tips_cb)(const char *refname, const git_oid *a, const git_oid *b, void *data);
	callbacks->sideband_progress = (git_transport_message_cb)sidebandProgressCallback;
	callbacks->completion = (completion_cb)completionCallback;
	callbacks->credentials = (credentials_cb)credentialsCallback;
	callbacks->transfer_progress = (transfer_progress_cb)transferProgressCallback;
	callbacks->update_tips = (update_tips_cb)updateTipsCallback;
}

typedef int (*status_foreach_cb)(const char *ref, const char *msg, void *data);

int _go_git_push_status_foreach(git_push *push, void *data)
{
	return git_push_status_foreach(push, (status_foreach_cb)statusForeach, data);
}

int _go_git_push_set_callbacks(git_push *push, void *packbuilder_progress_data, void *transfer_progress_data)
{
	return git_push_set_callbacks(push, packbuilderProgress, packbuilder_progress_data, pushTransferProgress, transfer_progress_data);
}

int _go_blob_chunk_cb(char *buffer, size_t maxLen, void *payload)
{
    return blobChunkCb(buffer, maxLen, payload);
}

int _go_git_blob_create_fromchunks(git_oid *id,
	git_repository *repo,
	const char *hintpath,
	void *payload)
{
    return git_blob_create_fromchunks(id, repo, hintpath, _go_blob_chunk_cb, payload);
}
/* EOF */
