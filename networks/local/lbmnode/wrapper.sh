#!/usr/bin/env sh

##
## Input parameters
##
BINARY=/lfb/${BINARY:-lfb}
ID=${ID:-0}
LOG=${LOG:-lfb.log}

##
## Assert linux binary
##
if ! [ -f "${BINARY}" ]; then
	echo "The binary $(basename "${BINARY}") cannot be found. Please add the binary to the shared folder. Please use the BINARY environment variable if the name of the binary is not 'lfb' E.g.: -e BINARY=lfb_my_test_version"
	exit 1
fi
BINARY_CHECK="$(file "$BINARY" | grep 'ELF 64-bit LSB executable, x86-64')"
if [ -z "${BINARY_CHECK}" ]; then
	echo "Binary needs to be OS linux, ARCH amd64"
	exit 1
fi

##
## Run binary with all parameters
##
export LFBHOME="/lfb/node${ID}/lfb"

if [ -d "$(dirname "${LFBHOME}"/"${LOG}")" ]; then
  "${BINARY}" --home "${LFBHOME}" "$@" | tee "${LFBHOME}/${LOG}"
else
  "${BINARY}" --home "${LFBHOME}" "$@"
fi

