#!/bin/sh

set -ex

export INSTALL_LOCATION=$PWD/vendor/install
export PKG_CONFIG_PATH=$INSTALL_LOCATION/lib/pkgconfig

export PCFILE="$PWD/vendor/libgit2/libgit2.pc"

export CGO_LDFLAGS="$(pkg-config --static --libs $PCFILE)"
export CGO_CFLAGS="$(pkg-config --static --cflags $PCFILE)"

$@
