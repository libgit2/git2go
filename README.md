git2go
======
[![GoDoc](https://godoc.org/github.com/libgit2/git2go?status.svg)](http://godoc.org/github.com/libgit2/git2go) [![Build Status](https://travis-ci.org/libgit2/git2go.svg?branch=master)](https://travis-ci.org/libgit2/git2go)

Go bindings for [libgit2](http://libgit2.github.com/).

### Which branch to use

The numbered branches work against the version of libgit2 as specified by their number. You can import them in your project via gopkg.in, e.g. if you have libgit2 v0.25 installed you'd import with

    import "gopkg.in/libgit2/git2go.v25"

which will ensure there are no sudden changes to the API.

The `master` branch follows the tip of libgit2 itself (with some lag) and as such has no guarantees on its own API nor does it have expectations the stability of libgit2's. Thus this only supports statically linking against libgit2.

Installing
----------

This project wraps the functionality provided by libgit2. It thus needs it in order to perform the work.

This project wraps the functionality provided by libgit2. If you're using a versioned branch, install it to your system via your system's package manager and then install git2go.


### Versioned branch, dynamic linking

When linking dynamically against a released version of libgit2, install it via your system's package manager. CGo will take care of finding its pkg-config file and set up the linking. Import via gopkg.in, e.g. to work against libgit2 v0.25

    import "gopkg.in/libgit2/git2go.v25"

### Master branch, or static linking

If using `master` or building a branch statically, we need to build libgit2 first. In order to build it, you need `cmake`, `pkg-config` and a C compiler. You will also need the development packages for OpenSSL (outside of Windows or macOS) and LibSSH2 installed if you want libgit2 to support HTTPS and SSH respectively. Note that even if libgit2 is included in the resulting binary, its dependencies will not be.

Run `go get -d github.com/libgit2/git2go` to download the code and go to your `$GOPATH/src/github.com/libgit2/git2go` directory. From there, we need to build the C code and put it into the resulting go binary.

    git submodule update --init # get libgit2
    make install-static

will compile libgit2, link it into git2go and install it. The `master` branch is set up to follow the specific libgit2 version that is vendored, so trying dynamic linking may or may not work depending on the exact versions involved.

Parallelism and network operations
----------------------------------

libgit2 may use OpenSSL and LibSSH2 for performing encrypted network connections. For now, git2go asks libgit2 to set locking for OpenSSL. This makes HTTPS connections thread-safe, but it is fragile and will likely stop doing it soon. This may also make SSH connections thread-safe if your copy of libssh2 is linked against OpenSSL. Check libgit2's `THREADSAFE.md` for more information.

Running the tests
-----------------

For the stable version, `go test` will work as usual. For the `master` branch, similarly to installing, running the tests requires building a local libgit2 library, so the Makefile provides a wrapper that makes sure it's built

    make test-static

Alternatively, you can build the library manually first and then run the tests

    ./script/build-libgit2-static.sh
    go test -v --tags "static" ./...

License
-------

M to the I to the T. See the LICENSE file if you've never seen an MIT license before.

Authors
-------

- Carlos Martín (@carlosmn)
- Vicent Martí (@vmg)

