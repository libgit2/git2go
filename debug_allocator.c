#include "_cgo_export.h"

#include <execinfo.h>
#include <stdio.h>
#include <stdlib.h>
#include <sys/socket.h>
#include <sys/uio.h>
#include <sys/un.h>

#include <git2.h>
#include <git2/common.h>
#include <git2/sys/alloc.h>

static git_allocator _go_git_system_allocator;
static git_allocator _go_git_debug_allocator;

static int __alloc_fd = -1;
typedef struct {
	int type;
	int line;
	uintptr_t ptr;
	size_t len;
	size_t filelen;
	size_t btlen;
} __alloc_message;

static void log_alloc_event(char type, const void* ptr, size_t len, const char *file, int line) {
	void *btaddr[16];
	__alloc_message msg = {
		.type = type,
		.line = line,
		.ptr = (uintptr_t)ptr,
		.len = len,
		.filelen = (file ? strlen(file) : 0),
		.btlen = 0,
	};
	struct iovec iov[18] = {
		{.iov_base = &msg, .iov_len = sizeof(msg)},
		{.iov_base = (char *)file, .iov_len = msg.filelen},
	};
	ssize_t expected = iov[0].iov_len + iov[1].iov_len;
	char **strings = NULL;
	int iov_count = 2;

	if (type == 'A') {
		size_t i;
		msg.btlen = backtrace(btaddr, 16);
		strings = backtrace_symbols(btaddr, (int)msg.btlen);

		for (i = 0; i < msg.btlen; ++i) {
			iov[iov_count].iov_base = strings[i];
			iov[iov_count].iov_len = strlen(strings[i]) + 1;
			expected += iov[iov_count].iov_len;
			++iov_count;
		}
	}
	if (writev(__alloc_fd, iov, iov_count) != expected) {
		perror("writev");
		abort();
	}
	free(strings);
}

static void *_go_git_debug_allocator__malloc(size_t len, const char *file, int line)
{
	void *ptr = _go_git_system_allocator.gmalloc(len, file, line);
	if (ptr)
		log_alloc_event('A', ptr, len, file, line);
	return ptr;
}

static void *_go_git_debug_allocator__calloc(size_t nelem, size_t elsize, const char *file, int line)
{
	void *ptr = _go_git_system_allocator.gcalloc(nelem, elsize, file, line);
	if (ptr)
		log_alloc_event('A', ptr, nelem * elsize, file, line);
	return ptr;
}

static char *_go_git_debug_allocator__strdup(const char *str, const char *file, int line)
{
	char *ptr = _go_git_system_allocator.gstrdup(str, file, line);
	if (ptr)
		log_alloc_event('A', ptr, strlen(ptr) + 1, file, line);
	return ptr;
}

static char *_go_git_debug_allocator__strndup(const char *str, size_t n, const char *file, int line)
{
	char *ptr = _go_git_system_allocator.gstrndup(str, n, file, line);
	if (ptr)
		log_alloc_event('A', ptr, strlen(ptr) + 1, file, line);
	return ptr;
}

static char *_go_git_debug_allocator__substrdup(const char *start, size_t n, const char *file, int line)
{
	char *ptr = _go_git_system_allocator.gsubstrdup(start, n, file, line);
	if (ptr)
		log_alloc_event('A', ptr, strlen(ptr) + 1, file, line);
	return ptr;
}

static void *_go_git_debug_allocator__realloc(void *ptr, size_t size, const char *file, int line)
{
	void *new_ptr = _go_git_system_allocator.grealloc(ptr, size, file, line);
	if (new_ptr != ptr) {
		if (ptr)
			log_alloc_event('D', ptr, 0, NULL, 0);
		if (new_ptr)
			log_alloc_event('A', new_ptr, size, file, line);
	} else if (new_ptr) {
		log_alloc_event('R', new_ptr, size, file, line);
	}
	return new_ptr;
}

static void *_go_git_debug_allocator__reallocarray(void *ptr, size_t nelem, size_t elsize, const char *file, int line)
{
	void *new_ptr = _go_git_system_allocator.greallocarray(ptr, nelem, elsize, file, line);
	if (new_ptr != ptr) {
		if (ptr)
			log_alloc_event('D', ptr, 0, NULL, 0);
		if (new_ptr)
			log_alloc_event('A', new_ptr, nelem * elsize, file, line);
	} else if (new_ptr) {
		log_alloc_event('R', new_ptr, nelem * elsize, file, line);
	}
	return new_ptr;
}

static void *_go_git_debug_allocator__mallocarray(size_t nelem, size_t elsize, const char *file, int line)
{
	void *ptr = _go_git_system_allocator.gmallocarray(nelem, elsize, file, line);
	if (ptr)
		log_alloc_event('A', ptr, nelem * elsize, file, line);
	return ptr;
}

static void _go_git_debug_allocator__free(void *ptr)
{
	_go_git_system_allocator.gfree(ptr);
	if (ptr)
		log_alloc_event('D', ptr, 0, NULL, 0);
}

int _go_git_setup_debug_allocator()
{
#if defined(LIBGIT2_STATIC)
	struct sockaddr_un name = {};
	int error;

	__alloc_fd = socket(AF_UNIX, SOCK_SEQPACKET, 0);
	if (__alloc_fd == -1) {
		perror("socket");
		return -1;
	}

	name.sun_family = AF_UNIX;
	strncpy(name.sun_path, "/run/alloc.sock", sizeof(name.sun_path) - 1);

	if (connect(__alloc_fd, (const struct sockaddr *)&name, sizeof(name)) == -1) {
		perror("connect");
		return -1;
	}

	error = git_stdalloc_init_allocator(&_go_git_system_allocator);
	if (error < 0)
		return error;
	_go_git_debug_allocator.gmalloc = _go_git_debug_allocator__malloc;
	_go_git_debug_allocator.gcalloc = _go_git_debug_allocator__calloc;
	_go_git_debug_allocator.gstrdup = _go_git_debug_allocator__strdup;
	_go_git_debug_allocator.gstrndup = _go_git_debug_allocator__strndup;
	_go_git_debug_allocator.gsubstrdup = _go_git_debug_allocator__substrdup;
	_go_git_debug_allocator.grealloc = _go_git_debug_allocator__realloc;
	_go_git_debug_allocator.greallocarray = _go_git_debug_allocator__reallocarray;
	_go_git_debug_allocator.gmallocarray = _go_git_debug_allocator__mallocarray;
	_go_git_debug_allocator.gfree = _go_git_debug_allocator__free;
	error = git_libgit2_opts(GIT_OPT_SET_ALLOCATOR, &_go_git_debug_allocator);
	if (error < 0)
		return error;

	return 0;
#elif defined(LIBGIT2_DYNAMIC)
	fprintf(stderr, "debug allocator is only enabled in static builds\n");
	return -1;
#else
#error no LIBGIT2_STATIC or LIBGIT2_DYNAMIC defined!
#endif
}
