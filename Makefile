TEST_ARGS ?= --count=1

default: test

# System library
# ==============
# This uses whatever version of libgit2 can be found in the system.
test:
	go run script/check-MakeGitError-thread-lock.go
	go test $(TEST_ARGS) ./...

install:
	go install ./...

# Bundled dynamic library
# =======================
# In order to avoid having to manipulate `git_dynamic.go`, which would prevent
# the system-wide libgit2.so from being used in a sort of ergonomic way, this
# instead moves the complexity of overriding the paths so that the built
# libraries can be found by the build and tests.
.PHONY: build-libgit2-dynamic
build-libgit2-dynamic:
	./script/build-libgit2-dynamic.sh

dynamic-build/install/lib/libgit2.so:
	./script/build-libgit2-dynamic.sh

test-dynamic: dynamic-build/install/lib/libgit2.so
	PKG_CONFIG_PATH=dynamic-build/install/lib/pkgconfig \
			go run script/check-MakeGitError-thread-lock.go
	PKG_CONFIG_PATH=dynamic-build/install/lib/pkgconfig \
			LD_LIBRARY_PATH=dynamic-build/install/lib \
			go test $(TEST_ARGS) ./...

install-dynamic: dynamic-build/install/lib/libgit2.so
	PKG_CONFIG_PATH=dynamic-build/install/lib/pkgconfig \
			go install ./...

# Bundled static library
# ======================
# This is mostly used in tests, but can also be used to provide a
# statically-linked library with the bundled version of libgit2.
.PHONY: build-libgit2-static
build-libgit2-static:
	./script/build-libgit2-static.sh

static-build/install/lib/libgit2.a:
	./script/build-libgit2-static.sh

test-static: static-build/install/lib/libgit2.a
	go run script/check-MakeGitError-thread-lock.go
	go test --tags "static" $(TEST_ARGS) ./...

install-static: static-build/install/lib/libgit2.a
	go install --tags "static" ./...
