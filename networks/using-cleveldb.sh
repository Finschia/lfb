#!/bin/bash

make install LINK_BUILD_OPTIONS="cleveldb"

lfb init "t6" --home ./t6 --chain-id t6

lfb unsafe-reset-all --home ./t6

mkdir -p ./t6/data/snapshots/metadata.db

lfb keys add validator --keyring-backend test --home ./t6

lfb add-genesis-account $(lfb keys show validator -a --keyring-backend test --home ./t6) 100000000stake --keyring-backend test --home ./t6

lfb gentx validator 100000000stake --keyring-backend test --home ./t6 --chain-id t6

lfb collect-gentxs --home ./t6

lfb start --db_backend cleveldb --home ./t6
