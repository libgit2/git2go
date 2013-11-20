#ifndef _GIT2GO_WRAP_H_
#define _GIT2GO_WRAP_H_

#include <string.h>

static inline git_error *copy_error(void)
{
	git_error *err;
	const git_error *last;

	last = giterr_last();
	if (last == NULL)
		return NULL;

	err = malloc(sizeof(git_error));
	if (err == NULL)
		return NULL;

	err->klass = last->klass;
	err->message = last->message ? strdup(last->message) : NULL;
	return err;
}

#define WRAP(name, call)			\
	int e_##name			\
	{					\
		int ret;			\
		ret = call;			\
		if (ret < 0)			\
			*err = copy_error();	\
		return ret;			\
	}

WRAP(git_config_get_string(const char **out, git_config *cfg, const char *name, git_error **err), git_config_get_string(out, cfg, name))
WRAP(git_config_set_string(git_config *cfg, const char *name, const char *value, git_error **err), git_config_set_string(cfg, name, value))

WRAP(git_config_get_int32(int32_t *out, git_config *cfg, const char *name, git_error **err), git_config_get_int32(out, cfg, name))
WRAP(git_config_get_int64(int64_t *out, git_config *cfg, const char *name, git_error **err), git_config_get_int64(out, cfg, name))

#endif /* _GIT2GO_WRAP_H_ */
