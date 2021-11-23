#!/usr/bin/make -f

BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
COMMIT := $(shell git log -1 --format='%H')

# ascribe tag only if on a release/ branch, otherwise pick branch name and concatenate commit hash
VERSION := $(shell echo $(shell git describe --tags) | sed 's/^v//')
ifeq (, $(findstring release/,$(BRANCH)))
  VERSION = $(BRANCH)-$(COMMIT)
endif

PACKAGES_SIMTEST=$(shell go list ./... | grep '/simulation')
LEDGER_ENABLED ?= false
SDK_PACK := $(shell go list -m github.com/line/lbm-sdk | sed  's/ /\@/g')
OST_VERSION := $(shell go list -m github.com/line/ostracon | sed 's:.* ::') # grab everything after the space in "github.com/line/ostracon v0.34.7"
DOCKER := $(shell which docker)
BUILDDIR ?= $(CURDIR)/build
TEST_DOCKER_REPO=jackzampolin/linktest
CGO_ENABLED ?= 1

export GO111MODULE = on

# process build tags

build_tags = netgo
ifeq ($(LEDGER_ENABLED),true)
  ifeq ($(OS),Windows_NT)
    GCCEXE = $(shell where gcc.exe 2> NUL)
    ifeq ($(GCCEXE),)
      $(error gcc.exe not installed for ledger support, please install or set LEDGER_ENABLED=false)
    else
      build_tags += ledger
    endif
  else
    UNAME_S = $(shell uname -s)
    ifeq ($(UNAME_S),OpenBSD)
      $(warning OpenBSD detected, disabling ledger support (https://github.com/cosmos/cosmos-sdk/issues/1988))
    else
      GCC = $(shell command -v gcc 2> /dev/null)
      ifeq ($(GCC),)
        $(error gcc not installed for ledger support, please install or set LEDGER_ENABLED=false)
      else
        build_tags += ledger
      endif
    endif
  endif
endif

# DB backend selection; use default for testing; use rocksdb or cleveldb for performance; build automation is not ready for boltdb and badgerdb yet.
ifeq (,$(filter $(LFB_BUILD_OPTIONS), cleveldb rocksdb boltdb badgerdb))
  BUILD_TAGS += goleveldb
  DB_BACKEND = goleveldb
else
  ifeq (cleveldb,$(findstring cleveldb,$(LFB_BUILD_OPTIONS)))
    CGO_ENABLED=1
    BUILD_TAGS += gcc cleveldb
    DB_BACKEND = cleveldb
    CLEVELDB_DIR = leveldb
    CGO_CFLAGS=-I$(shell pwd)/$(CLEVELDB_DIR)/include
    CGO_LDFLAGS="-L$(shell pwd)/$(CLEVELDB_DIR)/build -L$(shell pwd)/snappy/build -lleveldb -lm -lstdc++ -lsnappy"
  endif
  ifeq (badgerdb,$(findstring badgerdb,$(LFB_BUILD_OPTIONS)))
    BUILD_TAGS += badgerdb
    DB_BACKEND = badgerdb
  endif
  ifeq (rocksdb,$(findstring rocksdb,$(LFB_BUILD_OPTIONS)))
    CGO_ENABLED=1
    BUILD_TAGS += gcc rocksdb
    DB_BACKEND = rocksdb
    ROCKSDB_DIR=$(shell pwd)/rocksdb
    CGO_CFLAGS=-I$(ROCKSDB_DIR)/include
    CGO_LDFLAGS="-L$(ROCKSDB_DIR) -lrocksdb -lm -lstdc++ $(shell awk '/PLATFORM_LDFLAGS/ {sub("PLATFORM_LDFLAGS=", ""); print}' < $(ROCKSDB_DIR)/make_config.mk)"
  endif
  ifeq (boltdb,$(findstring boltdb,$(LFB_BUILD_OPTIONS)))
    BUILD_TAGS += boltdb
    DB_BACKEND = boltdb
  endif
endif

# secp256k1 implementation selection
ifeq (libsecp256k1,$(findstring libsecp256k1,$(LFB_BUILD_OPTIONS)))
  CGO_ENABLED=1
  BUILD_TAGS += libsecp256k1
endif

build_tags += $(BUILD_TAGS)
build_tags := $(strip $(build_tags))

whitespace :=
whitespace += $(whitespace)
comma := ,
build_tags_comma_sep := $(subst $(whitespace),$(comma),$(build_tags))

# process linker flags

ldflags = -X github.com/line/lbm-sdk/version.Name=lfb \
		  -X github.com/line/lbm-sdk/version.AppName=lfb \
		  -X github.com/line/lbm-sdk/version.Version=$(VERSION) \
		  -X github.com/line/lbm-sdk/version.Commit=$(COMMIT) \
		  -X github.com/line/lbm-sdk/types.DBBackend=$(DB_BACKEND) \
		  -X "github.com/line/lbm-sdk/version.BuildTags=$(build_tags_comma_sep)" \
		  -X github.com/line/ostracon/version.TMCoreSemVer=$(OST_VERSION)

ifeq (,$(findstring nostrip,$(LFB_BUILD_OPTIONS)))
  ldflags += -w -s
endif
ldflags += $(LDFLAGS)
ldflags := $(strip $(ldflags))

BUILD_FLAGS := -tags "$(build_tags)" -ldflags '$(ldflags)'
CLI_TEST_BUILD_FLAGS := -tags "cli_test $(build_tags)"
CLI_MULTI_BUILD_FLAGS := -tags "cli_multi_node_test $(build_tags)"
# check for nostrip option
ifeq (,$(findstring nostrip,$(LFB_BUILD_OPTIONS)))
  BUILD_FLAGS += -trimpath
endif

#$(info $$BUILD_FLAGS is [$(BUILD_FLAGS)])

# The below include contains the tools target.
include contrib/devtools/Makefile

###############################################################################
###                              Documentation                              ###
###############################################################################

all: install lint test

build: BUILD_ARGS=-o $(BUILDDIR)/

build: go.sum $(BUILDDIR)/ dbbackend
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) CGO_ENABLED=$(CGO_ENABLED) go build -mod=readonly $(BUILD_FLAGS) $(BUILD_ARGS) ./...

build-static: go.sum $(BUILDDIR)/
	docker build -t line/lfb-builder:static -f builders/Dockerfile.static .
	docker run -it --rm -v $(shell pwd):/code -e LFB_BUILD_OPTIONS="$(LFB_BUILD_OPTIONS)" line/lfb-builder:static

install: go.sum $(BUILDDIR)/ dbbackend
	CGO_CFLAGS=$(CGO_CFLAGS) CGO_LDFLAGS=$(CGO_LDFLAGS) CGO_ENABLED=$(CGO_ENABLED) go install $(BUILD_FLAGS) $(BUILD_ARGS) ./cmd/lfb

$(BUILDDIR)/:
	mkdir -p $(BUILDDIR)/

.PHONY: dbbackend
# for more faster building use -j8; but it will be failed in docker building because of low memory
ifeq ($(DB_BACKEND), rocksdb)
dbbackend:
	@if [ ! -e $(ROCKSDB_DIR) ]; then          \
		sh ./contrib/get_rocksdb.sh;         \
	fi
	@if [ ! -e $(ROCKSDB_DIR)/librocksdb.a ]; then    \
		cd $(ROCKSDB_DIR) && make -j2 static_lib; \
	fi
	@if [ ! -e $(ROCKSDB_DIR)/libsnappy.a ]; then    \
                cd $(ROCKSDB_DIR) && make libsnappy.a DEBUG_LEVEL=0; \
        fi
else ifeq ($(DB_BACKEND), cleveldb)
dbbackend:
	@if [ ! -e $(CLEVELDB_DIR) ]; then         \
		sh contrib/get_cleveldb.sh;        \
	fi
	@if [ ! -e $(CLEVELDB_DIR)/libcleveldb.a ]; then   \
		cd $(CLEVELDB_DIR);                        \
		mkdir build;                               \
		cd build;                                  \
		cmake -DCMAKE_BUILD_TYPE=Release -DBUILD_SHARED_LIBS=OFF -DLEVELDB_BUILD_TESTS=OFF -DLEVELDB_BUILD_BENCHMARKS=OFF ..; \
		make;                                      \
	fi
	@if [ ! -e snappy ]; then \
		sh contrib/get_snappy.sh; \
		cd snappy; \
		mkdir build && cd build; \
		cmake -DBUILD_SHARED_LIBS=OFF -DSNAPPY_BUILD_TESTS=OFF -DSNAPPY_REQUIRE_AVX2=ON ..;\
		make; \
	fi
else
dbbackend:
endif

build-reproducible: go.sum
	$(DOCKER) rm latest-build || true
	$(DOCKER) run --volume=$(CURDIR):/sources:ro \
        --env TARGET_PLATFORMS='linux/amd64 darwin/amd64 linux/arm64 windows/amd64' \
        --env APP=lfb \
        --env VERSION=$(VERSION) \
        --env COMMIT=$(COMMIT) \
        --env LEDGER_ENABLED=$(LEDGER_ENABLED) \
        --name latest-build cosmossdk/rbuilder:latest
	$(DOCKER) cp -a latest-build:/home/builder/artifacts/ $(CURDIR)/

build-docker:
	docker build --build-arg LFB_BUILD_OPTIONS="$(LFB_BUILD_OPTIONS)" -t line/lfb .

build-contract-tests-hooks:
	mkdir -p $(BUILDDIR)
	go build -mod=readonly $(BUILD_FLAGS) -o $(BUILDDIR)/ ./cmd/contract_tests

go-mod-cache: go.sum
	@echo "--> Download go modules to local cache"
	@go mod download

go.sum: go.mod
	@echo "--> Ensure dependencies have not been modified"
	@go mod verify

draw-deps:
	@# requires brew install graphviz or apt-get install graphviz
	go get github.com/RobotsAndPencils/goviz
	@goviz -i ./cmd/lfb -d 2 | dot -Tpng -o dependency-graph.png

clean:
	rm -rf $(BUILDDIR)/ artifacts/
	@ROCKSDB_DIR=rocksdb;				\
	if [ -e $${ROCKSDB_DIR}/Makefile ]; then	\
		cd $${ROCKSDB_DIR};			\
		make clean;				\
	fi

distclean: clean
	rm -rf vendor/

###############################################################################
###                                 Devdoc                                  ###
###############################################################################

build-docs:
	@cd docs && \
	while read p; do \
		(git checkout $${p} && npm install && VUEPRESS_BASE="/$${p}/" npm run build) ; \
		mkdir -p ~/output/$${p} ; \
		cp -r .vuepress/dist/* ~/output/$${p}/ ; \
		cp ~/output/$${p}/index.html ~/output ; \
	done < versions ;
.PHONY: build-docs

sync-docs:
	cd ~/output && \
	echo "role_arn = ${DEPLOYMENT_ROLE_ARN}" >> /root/.aws/config ; \
	echo "CI job = ${CIRCLE_BUILD_URL}" >> version.html ; \
	aws s3 sync . s3://${WEBSITE_BUCKET} --profile terraform --delete ; \
	aws cloudfront create-invalidation --distribution-id ${CF_DISTRIBUTION_ID} --profile terraform --path "/*" ;
.PHONY: sync-docs


###############################################################################
###                           Tests & Simulation                            ###
###############################################################################

include sims.mk

test: test-unit test-build

test-all: check test-race test-cover

test-unit:
	@VERSION=$(VERSION) go test -mod=readonly -tags='ledger test_ledger_mock' ./...

test-race:
	@VERSION=$(VERSION) go test -mod=readonly -race -tags='ledger test_ledger_mock' ./...

test-cover:
	@go test -mod=readonly -timeout 30m -race -coverprofile=coverage.txt -covermode=atomic -tags='ledger test_ledger_mock' ./...

benchmark:
	@go test -mod=readonly -bench=. ./...

test-integration: build
	@go test -mod=readonly -p 4 `go list ./cli_test/...` $(CLI_TEST_BUILD_FLAGS) -v

test-integration-multi-node: build-docker
	@go test -mod=readonly -p 4 `go list ./cli_test/...` $(CLI_MULTI_BUILD_FLAGS) -v


###############################################################################
###                                Linting                                  ###
###############################################################################

lint:
	golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s

format:
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs gofmt -w -s
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs misspell -w
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" -not -path "./client/lcd/statik/statik.go" | xargs goimports -w -local github.com/line/lbm-sdk

###############################################################################
###                                Localnet                                 ###
###############################################################################

build-docker-lfbnode:
	$(MAKE) -C networks/local

# Run a 4-node testnet locally
localnet-start: build-docker-lfbnode build-static localnet-stop
	@if ! [ -f build/node0/lfb/config/genesis.json ]; then docker run --rm -v $(CURDIR)/build:/lfb:Z line/lfbnode testnet --v 4 -o . --starting-ip-address 192.168.10.2 --keyring-backend=test ; fi
	docker-compose up -d

# Stop testnet
localnet-stop:
	docker-compose down

test-docker:
	@docker build -f contrib/Dockerfile.test -t ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) .
	@docker tag ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) ${TEST_DOCKER_REPO}:$(shell git rev-parse --abbrev-ref HEAD | sed 's#/#_#g')
	@docker tag ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD) ${TEST_DOCKER_REPO}:latest

test-docker-push: test-docker
	@docker push ${TEST_DOCKER_REPO}:$(shell git rev-parse --short HEAD)
	@docker push ${TEST_DOCKER_REPO}:$(shell git rev-parse --abbrev-ref HEAD | sed 's#/#_#g')
	@docker push ${TEST_DOCKER_REPO}:latest

.PHONY: all install format lint \
	go-mod-cache draw-deps clean build \
	setup-transactions setup-contract-tests-data start-link run-lcd-contract-tests contract-tests \
	test test-all test-build test-cover test-unit test-race \
	benchmark \
	build-docker-lfbnode localnet-start localnet-stop \
	docker-single-node
