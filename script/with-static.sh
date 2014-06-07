#!/bin/sh

set -ex

export BUILD="$PWD/vendor/libgit2/build"
export PCFILE="$BUILD/libgit2.pc"

export CGO_LDFLAGS="$BUILD/libgit2.a -L$BUILD $(pkg-config --static --libs $PCFILE)"
export CGO_CFLAGS="-I$PWD/vendor/libgit2/include"

$@
