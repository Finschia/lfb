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

# prepare dbbackend before building; this can be cached
COPY ./Makefile ./
COPY ./contrib ./contrib
COPY ./sims.mk ./
RUN make dbbackend LFB_BUILD_OPTIONS="$LFB_BUILD_OPTIONS"

# Install GO dependencies
COPY ./go.mod /lfb-build/lfb/go.mod
COPY ./go.sum /lfb-build/lfb/go.sum
RUN go mod download

# Build cosmwasm
RUN cd $(go list -f "{{ .Dir }}" -m github.com/line/wasmvm) && \
    RUSTFLAGS='-C target-feature=-crt-static' cargo build --release --example muslc && \
    mv target/release/examples/libmuslc.a /usr/lib/libwasmvm_muslc.a && \
    rm -rf target

# Add source files
COPY . .

# Make install
RUN BUILD_TAGS=muslc make install CGO_ENABLED=1 LFB_BUILD_OPTIONS="$LFB_BUILD_OPTIONS"

# Final image
FROM alpine:edge

WORKDIR /root

# Set up OS dependencies
RUN apk add --update --no-cache libstdc++ ca-certificates

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/lfb /usr/bin/lfb

