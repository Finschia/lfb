<!--
order: 2
-->

# Install lfb

This guide will explain how to install the `lfb` entrypoint
onto your system. With these installed on a server, you can participate in the
mainnet as either a [Full Node](./join-mainnet.md) or a
[Validator](../validators/validator-setup.md).

## Install Go

Install `go` by following the [official docs](https://golang.org/doc/install).
Remember to set your `$PATH` environment variable, for example:

```bash
mkdir -p $HOME/go/bin
echo "export PATH=$PATH:$(go env GOPATH)/bin" >> ~/.bash_profile
source ~/.bash_profile
```

::: tip
**Go 1.15+** is required for the Cosmos SDK.
:::

## Install the binaries

Next, let's install the latest version of Gaia. Make sure you `git checkout` the
correct [released version](https://github.com/line/lfb/releases).

```bash
git clone -b <latest-release-tag> https://github.com/line/lfb
cd lfb && make install
```

If this command fails due to the following error message, you might have already set `LDFLAGS` prior to running this step.

```
# github.com/lfb/lfb/cmd/lfb
flag provided but not defined: -L
usage: link [options] main.o
...
make: *** [install] Error 2
```

Unset this environment variable and try again.

```
LDFLAGS="" make install
```

> _NOTE_: If you still have issues at this step, please check that you have the latest stable version of GO installed.

That will install the `lfb` binary. Verify that everything is OK:

```bash
lfb version --long
```

`lfb` for instance should output something similar to:

```bash
name: lfb
server_name: lfb
version: 1.0.0
commit: 8692310a5361006f8c02d44cd7df2d41f130089b
build_tags: netgo,goleveldb
go: go version go1.15.2 darwin/amd64
build_deps:
- github.com/...
- github.com/...
...
```

### Build Tags

Build tags indicate special features that have been enabled in the binary.

| Build Tag | Description                                     |
| --------- | ----------------------------------------------- |
| netgo     | Name resolution will use pure Go code           |
| goleveldb | DB backend used for persistent DB               |
| ledger    | Ledger devices are supported (hardware wallets) |

### Install binary distribution via snap (Linux only)

**Do not use snap at this time to install the binaries for production until we have a reproducible binary system.**

## Developer Workflow

To test any changes made in the SDK or Ostracon, a `replace` clause needs to be added to `go.mod` providing the correct import path.

- Make appropriate changes
- Add `replace github.com/line/lfb-sdk => /path/to/clone/lfb-sdk` to `go.mod`
- Run `make clean install` or `make clean build`
- Test changes
