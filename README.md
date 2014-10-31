git2go
======
[![GoDoc](https://godoc.org/github.com/libgit2/git2go?status.svg)](http://godoc.org/github.com/libgit2/git2go)


Go bindings for [libgit2](http://libgit2.github.com/). The master branch follows the latest libgit2 release.

Installing
----------

This project needs libgit2, which is written in C so we need to build that as well. In order to build libgit2, you need `cmake`, `pkg-config` and a C compiler. You will also need the development packages for OpenSSL and LibSSH2 if you want to use HTTPS and SSH respectively.

Run `go get -d github.com/libgit2/git2go` to download the code and go to your `$GOPATH/src/github.com/libgit2/git2go` dir. From there, we need to build the C code and put it into the resulting go binary.

    git submodule update --init # get libgit2
    make install

will compile libgit2 and run `go install` such that it's statically linked to the git2go package.

Running the tests
-----------------

Similarly to installing, running the tests requires linking against the local libgit2 library, so the Makefile provides a wrapper

    make test

alternatively, if you want to pass arguments to `go test`, you can use the script that sets it all up

    ./script/with-static.sh go test -v

which will run the specified arguments with the correct environment variables.

License
-------

M to the I to the T. See the LICENSE file if you've never seen a MIT license before.

Authors
-------

- Carlos Martín (@carlosmn)
- Vicent Martí (@vmg)

