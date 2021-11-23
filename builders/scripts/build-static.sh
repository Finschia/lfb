#!/usr/bin/env sh

PROJECT_ROOT=$(realpath "$(dirname "$0")/../..")

# install dependencies
go mod download

# build wasmvm static
cd "$(go list -f "{{ .Dir }}" -m github.com/line/wasmvm)" || exit 1
RUSTFLAGS='-C target-feature=-crt-static' cargo build --release --example staticlib
mv -f target/release/examples/libstaticlib.a /usr/lib/libwasmvm_static.a
rm -rf target

cd "${PROJECT_ROOT}" || exit 1

# build lfb
BUILD_TAGS=static make build LFB_BUILD_OPTIONS="${LFB_BUILD_OPTIONS}"
