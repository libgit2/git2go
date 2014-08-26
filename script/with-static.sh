#!/bin/sh

set -ex

export BUILD="$PWD/vendor/libgit2/build"
export PCFILE="$BUILD/libgit2.pc"

FLAGS=$(pkg-config --static --libs $PCFILE) || exit 1
export CGO_LDFLAGS="$BUILD/libgit2.a -L$BUILD ${FLAGS}"
export CGO_CFLAGS="-I$PWD/vendor/libgit2/include"

$@
