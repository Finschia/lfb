#!/usr/bin/env bash
set -ex

mode="mainnet"

if [[ $1 == "docker" ]]
then
    if [[ $2 == "testnet" ]]
    then
        mode="testnet"
    fi
    LFB="docker run -i -p 26656:26656 -p 26657:26657 -v ${HOME}/.lfb:/root/.lfb line/lfb lfb"
elif [[ $1 == "testnet" ]]
then
    mode="testnet"
fi

LFB=${LFB:-lfb}

# initialize
rm -rf ~/.lfb

# TODO
# Configure your CLI to eliminate need for chain-id flag
#${LFB} config chain-id lfb
#${LFB} config output json
#${LFB} config indent true
#${LFB} config trust-node true
#${LFB} config keyring-backend test

# Initialize configuration files and genesis file
# moniker is the name of your node
${LFB} init solo --chain-id=lfb

# configure for testnet
if [[ ${mode} == "testnet" ]]
then
    if [[ $1 == "docker" ]]
    then
        docker run -i -p 26656:26656 -p 26657:26657 -v ${HOME}/.lfb:/root/.lfb line/lfb sh -c "export LFB_TESTNET=true"
    else
       export LFB_TESTNET=true
    fi
fi

# Please do not use the TEST_MNEMONIC for production purpose
TEST_MNEMONIC="mind flame tobacco sense move hammer drift crime ring globe art gaze cinnamon helmet cruise special produce notable negative wait path scrap recall have"

${LFB} keys add jack --keyring-backend=test --recover --account=0 <<< ${TEST_MNEMONIC}
${LFB} keys add alice --keyring-backend=test --recover --account=1 <<< ${TEST_MNEMONIC}
${LFB} keys add bob --keyring-backend=test --recover --account=2 <<< ${TEST_MNEMONIC}
${LFB} keys add rinah --keyring-backend=test --recover --account=3 <<< ${TEST_MNEMONIC}
${LFB} keys add sam --keyring-backend=test --recover --account=4 <<< ${TEST_MNEMONIC}
${LFB} keys add evelyn --keyring-backend=test --recover --account=5 <<< ${TEST_MNEMONIC}

# TODO
#if [[ ${mode} == "testnet" ]]
#then
#   ${LFB} add-genesis-account tlink15la35q37j2dcg427kfy4el2l0r227xwhc2v3lg 9223372036854775807link,1stake
#else
#   ${LFB} add-genesis-account link15la35q37j2dcg427kfy4el2l0r227xwhuaapxd 9223372036854775807link,1stake
#fi
# Add both accounts, with coins to the genesis file
${LFB} add-genesis-account $(${LFB} keys show jack -a --keyring-backend=test) 1000link,1000000000000stake
${LFB} add-genesis-account $(${LFB} keys show alice -a --keyring-backend=test) 1000link,1000000000000stake
${LFB} add-genesis-account $(${LFB} keys show bob -a --keyring-backend=test) 1000link,1000000000000stake
${LFB} add-genesis-account $(${LFB} keys show rinah -a --keyring-backend=test) 1000link,1000000000000stake
${LFB} add-genesis-account $(${LFB} keys show sam -a --keyring-backend=test) 1000link,1000000000000stake
${LFB} add-genesis-account $(${LFB} keys show evelyn -a --keyring-backend=test) 1000link,1000000000000stake

${LFB} gentx jack 100000000stake --keyring-backend=test --chain-id=lfb

${LFB} collect-gentxs

${LFB} validate-genesis

# ${LFB} start --log_level *:debug --rpc.laddr=tcp://0.0.0.0:26657 --p2p.laddr=tcp://0.0.0.0:26656

