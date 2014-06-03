#!/bin/sh

set -ex

export LIBGIT2_LOCATION=$PWD/libgit2/install
export PKG_CONFIG_PATH=$LIBGIT2_LOCATION/lib/pkgconfig
export LIBGIT2_A=$LIBGIT2_LOCATION/lib/libgit2.a
export CGO_LDFLAGS="$LIBGIT2_A $(pkg-config --static --libs libgit2)"

$@
