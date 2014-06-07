git2go
======

Go bindings for [libgit2](http://libgit2.github.com/). These bindings are for top-of-the-branch libgit2, and they move fast, things may or may not work. Operator get me Beijing-jing-jing-jing!

Installing
----------

This project needs libgit2, which is written in C so we need to take an extra step. Run `go get github.com/libgit2/git2go` and go to your `$GOROOT/src/github.com/libgt2/git2go` dir. From there, we need to build the C code and put it into the resulting go binary.

    git submodule update --init
	make install

will compile libgit2, build it statically into git2go and install the resulting object file where your Go project can use it.

License
-------

M to the I to the T. See the LICENSE file if you've never seen a MIT license before.

Authors
-------

- Carlos Martín (@carlosmn)
- Vicent Martí (@vmg)

