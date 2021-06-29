# LFB(LINE Financial Blockchain)

[![codecov](https://codecov.io/gh/line/lfb/branch/main/graph/badge.svg?token=JFFuUevpzJ)](https://codecov.io/gh/line/lfb)

This repository hosts `LFB(LINE Financial Blockchain)`. This repository is forked from [gaia](https://github.com/cosmos/gaia) at 2021-03-15. LFB is a mainnet app implementation using [lfb-sdk](https://github.com/line/lfb-sdk) and [ostracon](https://github.com/line/ostracon).

**Node**: Requires [Go 1.15+](https://golang.org/dl/)

**Warnings**: Initial development is in progress, but there has not yet been a stable.

# Quick Start
## Docker
**Build Docker Image**
```
make build-docker GITHUB_TOKEN=${YOUR_GITHUB_TOKEN}                # build docker image
```
or
```
make build-docker WITH_CLEVELDB=yes GITHUB_TOKEN=${YOUR_GITHUB_TOKEN}  # build docker image with cleveldb
```

**Configure**
```
./.initialize.sh docker          # prepare keys, validators, initial state, etc.
```
or
```
./.initialize.sh docker testnet  # prepare keys, validators, initial state, etc. for testnet
```

**Run**
```
docker-compose up                # Run a node
```

**visit with your browser**
* Node: http://localhost:26657/
* REST: http://localhost:1317/swagger-ui/

## Local
**Set up permissions**
```
go env -w GOPRIVATE="github.com/line/*"
git config --global url."https://${YOUR_GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"
```

_Note1_

You have to replace ${YOUR_GITHUB_TOKEN} with your token.

To create a token, 
see: https://docs.github.com/en/free-pro-team@latest/github/authenticating-to-github/creating-a-personal-access-token

_Note2_

Please check `GOPRIVATE` is set by run export and check the result. 
```
go env
```
if you can see `GOPRIVATE`, then you're good to go. 

Otherwise you need to set `GOPRIVATE` as environment variable.

**Build**
```
make build
make install 
```

**Configure**
```
./.initialize.sh
```
or
```
./.initialize.sh testnet  # for testnet
```

**Run**
```
lfb start                # Run a node
```

**visit with your browser**
* Node: http://localhost:26657/
* REST: http://localhost:1317/swagger-ui/
