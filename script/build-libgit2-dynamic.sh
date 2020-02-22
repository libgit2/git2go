#!/bin/sh

set -e

exec "$(dirname "$0")/build-libgit2.sh" --dynamic
