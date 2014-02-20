#include "_cgo_export.h"
#include "git2.h"
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

int _go_git_diff_foreach(git_diff *diff, int eachFile, int eachHunk, int eachLine, void *payload)
{
	git_diff_file_cb fcb = NULL;	
	git_diff_hunk_cb hcb = NULL;
	git_diff_line_cb lcb = NULL;

	if (eachFile) {
		fcb = (git_diff_file_cb)&diffForEachFileCb;
	}

	if (eachHunk) {
		hcb = (git_diff_hunk_cb)&diffForEachHunkCb;
	}

	if (eachLine) {
		lcb = (git_diff_line_cb)&diffForEachLineCb;
	}

	return git_diff_foreach(diff, fcb, hcb, lcb, payload);
}
/* EOF */
