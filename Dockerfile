# Simple usage with a mounted data directory:
# > docker build -t line/lfb .
# > docker run -it -p 26656:26656 -p 26657:26657 -v ~/.lfb:/root/.lfb -v line/lfb lfb init
# > docker run -it -p 26656:26656 -p 26657:26657 -v ~/.lfb:/root/.lfb -v line/lfb lfb start --rpc.laddr=tcp://0.0.0.0:26657 --p2p.laddr=tcp://0.0.0.0:26656
FROM golang:1.15-alpine3.13 AS build-env
ARG LFB_BUILD_OPTIONS=""

# Set up OS dependencies
ENV PACKAGES curl wget make cmake git libc-dev bash gcc g++ linux-headers eudev-dev python3 perl
RUN apk add --update --no-cache $PACKAGES

# Set WORKDIR to lfb
WORKDIR /lfb-build/lfb

# Add source files
COPY . .

# Install GO dependencies
RUN go mod download

# Make install
RUN make install LFB_BUILD_OPTIONS="$LFB_BUILD_OPTIONS"

# Final image
FROM alpine:edge

WORKDIR /root

# Set up OS dependencies
RUN apk add --update --no-cache libstdc++ ca-certificates

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/lfb /usr/bin/lfb

