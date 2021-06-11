# Simple usage with a mounted data directory:
# > docker build -t line/lfb .
# > docker run -it -p 26656:26656 -p 26657:26657 -v ~/.lfb:/root/.lfb -v line/lfb lfb init
# > docker run -it -p 26656:26656 -p 26657:26657 -v ~/.lfb:/root/.lfb -v line/lfb lfb start --rpc.laddr=tcp://0.0.0.0:26657 --p2p.laddr=tcp://0.0.0.0:26656
FROM golang:1.15-alpine AS build-env
ARG GITHUB_TOKEN=""
ARG LFB_BUILD_OPTIONS=""

# Set up OS dependencies
ENV PACKAGES curl wget make cmake git libc-dev bash gcc g++ linux-headers eudev-dev python3 perl
RUN apk add --update --no-cache $PACKAGES

# Install cleveldb, rocksdb w/ snappy
WORKDIR /lfb-build
COPY ./contrib /lfb-build/lfb/contrib
RUN mkdir -p /usr/local/lib64
RUN ./lfb/contrib/get_snappy.sh $LFB_BUILD_OPTIONS
RUN ./lfb/contrib/get_cleveldb.sh $LFB_BUILD_OPTIONS
RUN ./lfb/contrib/get_rocksdb.sh $LFB_BUILD_OPTIONS
RUN tar cvzf lib.tar.gz -C /usr/local/lib64 .

# Install rust build tools
ENV RUSTUP_HOME=/usr/local/rustup
ENV CARGO_HOME=/usr/local/cargo
ENV PATH=$CARGO_HOME/bin:$PATH

RUN wget "https://static.rust-lang.org/rustup/dist/x86_64-unknown-linux-musl/rustup-init"
RUN chmod +x rustup-init
RUN ./rustup-init -y --no-modify-path --default-toolchain 1.50.0

RUN rm rustup-init
RUN chmod -R a+w $RUSTUP_HOME $CARGO_HOME

# Set WORKDIR to lfb
WORKDIR /lfb-build/lfb

# Install GO dependencies
COPY ./go.mod /lfb-build/lfb/go.mod
COPY ./go.sum /lfb-build/lfb/go.sum
RUN go env -w GOPRIVATE=github.com/line/*
RUN git config --global url."https://$GITHUB_TOKEN:x-oauth-basic@github.com/".insteadOf "https://github.com/"
RUN go mod download

# Build cosmwasm
RUN cd $(go list -f "{{ .Dir }}" -m github.com/line/wasmvm) && \
    cargo build --release --example muslc && \
    mv target/release/examples/libmuslc.a /usr/lib/libwasmvm_muslc.a

# Add source files
COPY . .

# Make install
RUN BUILD_TAGS=muslc make install CGO_ENABLED=1 LFB_BUILD_OPTIONS=$LFB_BUILD_OPTIONS

# Final image
FROM alpine:edge

WORKDIR /root

# Set up OS dependencies
RUN apk add --update --no-cache libstdc++ ca-certificates

# Copy over binaries from the build-env
COPY --from=build-env /go/bin/lfb /usr/bin/lfb
COPY --from=build-env /lfb-build/lib.tar.gz .

RUN tar xvzf lib.tar.gz --directory /usr/lib && rm lib.tar.gz


