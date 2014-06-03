#!/bin/sh

set -ex

export INSTALL_LOCATION=$PWD/vendor/install
export PKG_CONFIG_PATH=$INSTALL_LOCATION/lib/pkgconfig

$@
