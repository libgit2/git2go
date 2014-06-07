#!/bin/sh

set -ex

VENDORED_PATH=vendor/libgit2

cd $VENDORED_PATH

cmake -DTHREADSAFE=ON \
      -DBUILD_CLAR=OFF \
      -DBUILD_SHARED_LIBS=OFF \
      -DCMAKE_C_FLAGS=-fPIC \
      -DCMAKE_INSTALL_PREFIX=../install \
      .

make install
