<!--
order: 6
-->

# Deploy Your Own Gaia Testnet

This document describes 3 ways to setup a network of `lfb` nodes, each serving a different usecase:

1. Single-node, local, manual testnet
2. Multi-node, local, automated testnet
3. Multi-node, remote, automated testnet

Supporting code can be found in the [networks directory](https://github.com/line/lfb/tree/main/networks) and additionally the `local` or `remote` sub-directories.

> NOTE: The `remote` network bootstrapping may be out of sync with the latest releases and is not to be relied upon.

## Available Docker images

TBD

## Single-node, Local, Manual Testnet

This guide helps you create a single validator node that runs a network locally for testing and other development related uses.

### Requirements

- [Install lfb](./installation.md)
- [Install `jq`](https://stedolan.github.io/jq/download/) (optional)

### Create Genesis File and Start the Network

```bash
# You can run all of these commands from your home directory
cd $HOME

# Initialize the genesis.json file that will help you to bootstrap the network
lfb init --chain-id=testing testing

# Create a key to hold your validator account
lfb keys add validator

# Add that key into the genesis.app_state.accounts array in the genesis file
# NOTE: this command lets you set the number of coins. Make sure this account has some coins
# with the genesis.app_state.staking.params.bond_denom denom, the default is staking
lfb add-genesis-account $(lfb keys show validator -a) 1000000000stake,1000000000validatortoken

# Generate the transaction that creates your validator
lfb gentx validator 1000000000stake --chain-id testing

# Add the generated bonding transaction to the genesis file
lfb collect-gentxs

# Now its safe to start `lfb`
lfb start
```

This setup puts all the data for `lfb` in `~/.lfb`. You can examine the genesis file you created at `~/.lfb/config/genesis.json`. With this configuration `lfb` is also ready to use and has an account with tokens (both staking and custom).

## Multi-node, Local, Automated Testnet

From the [networks/local directory](https://github.com/line/lfb/tree/main/networks/local):

### Requirements

- [Install lfb](./installation.md)
- [Install docker](https://docs.docker.com/engine/installation/)
- [Install docker-compose](https://docs.docker.com/compose/install/)

### Build

Build the `lfb` binary (linux) and the `line/lfbnode` docker image required for running the `localnet` commands. This binary will be mounted into the container and can be updated rebuilding the image, so you only need to build the image once.

```bash
# Clone the lfb repo
git clone https://github.com/line/lfb.git

# Work from the lfb repo
cd lfb

# Build the linux binary in ./build
make build-linux

# Build line/lfbnode image
make build-docker-lfbnode
```

### Run Your Testnet

To start a 4 node testnet run:

```
make localnet-start
```

This command creates a 4-node network using the lfbnode image.
The ports for each node are found in this table:

| Node ID    | P2P Port | RPC Port |
| ---------- | -------- | -------- |
| `lfbnode0` | `26656`  | `26657`  |
| `lfbnode1` | `26659`  | `26660`  |
| `lfbnode2` | `26661`  | `26662`  |
| `lfbnode3` | `26663`  | `26664`  |

To update the binary, just rebuild it and restart the nodes:

```
make build-linux localnet-start
```

To stop docker lfb nodes:

```
make localnet-stop
```

### Configuration

The `make localnet-start` creates files for a 4-node testnet in `./build` by
calling the `lfb testnet` command. This outputs a handful of files in the
`./build` directory:

```bash
$ tree -L 2 build/
build/
├── gentxs
│   ├── node0.json
│   ├── node1.json
│   ├── node2.json
│   └── node3.json
├── lfb
├── node0
│   └── lfb
│       ├── config
│       ├── data
│       ├── key_seed.json
│       ├── keyring-test
│       └── lfb.log
├── node1
│   └── lfb
│       ├── config
│       ├── data
│       ├── key_seed.json
│       ├── keyring-test
│       └── lfb.log
├── node2
│   └── lfb
│       ├── config
│       ├── data
│       ├── key_seed.json
│       ├── keyring-test
│       └── lfb.log
└── node3
    └── lfb
        ├── config
        ├── data
        ├── key_seed.json
        ├── keyring-test
        └── lfb.log
```

Each `./build/nodeN` directory is mounted to the `/lfb` directory in each container.

### Logging

Logs are saved under each `./build/nodeN/lfb/lfb.log`. You can also watch logs
directly via Docker, for example:

```
docker logs -f lfbnode0
```

### Keys & Accounts

To interact with `lfb` and start querying state or creating txs, you use the
`lfb` directory of any given node as your `home`, for example:

```bash
lfb keys list --home ./build/node0/lfb
```

Now that accounts exists, you may create new accounts and send those accounts
funds!

::: tip
**Note**: Each node's seed is located at `./build/nodeN/lfb/key_seed.json` and can be restored to the CLI using the `lfb keys add --restore` command
:::

### Special Binaries

If you have multiple binaries with different names, you can specify which one to run with the BINARY environment variable. The path of the binary is relative to the attached volume. For example:

```
# Run with custom binary
BINARY=lfbfoo make localnet-start
```

## Multi-Node, Remote, Automated Testnet

The following should be run from the [networks directory](https://github.com/line/lfb/tree/main/networks).

### Terraform & Ansible

Automated deployments are done using [Terraform](https://www.terraform.io/) to create servers on AWS then
[Ansible](http://www.ansible.com/) to create and manage testnets on those servers.

### Prerequisites

- Install [Terraform](https://www.terraform.io/downloads.html) and [Ansible](http://docs.ansible.com/ansible/latest/installation_guide/intro_installation.html) on a Linux machine.
- Create an [AWS API token](https://docs.aws.amazon.com/general/latest/gr/managing-aws-access-keys.html) with EC2 create capability.
- Create SSH keys

```
export AWS_ACCESS_KEY_ID="2345234jk2lh4234"
export AWS_SECRET_ACCESS_KEY="234jhkg234h52kh4g5khg34"
export TESTNET_NAME="remotenet"
export CLUSTER_NAME= "remotenetvalidators"
export SSH_PRIVATE_FILE="$HOME/.ssh/id_rsa"
export SSH_PUBLIC_FILE="$HOME/.ssh/id_rsa.pub"
```

These will be used by both `terraform` and `ansible`.

### Create a Remote Network

```
SERVERS=1 REGION_LIMIT=1 make validators-start
```

The testnet name is what's going to be used in --chain-id, while the cluster name is the administrative tag in AWS for the servers. The code will create SERVERS amount of servers in each availability zone up to the number of REGION_LIMITs, starting at us-east-2. (us-east-1 is excluded.) The below BaSH script does the same, but sometimes it's more comfortable for input.

```
./new-testnet.sh "$TESTNET_NAME" "$CLUSTER_NAME" 1 1
```

### Quickly see the /status Endpoint

```
make validators-status
```

### Delete Servers

```
make validators-stop
```

### Logging

You can ship logs to Logz.io, an Elastic stack (Elastic search, Logstash and Kibana) service provider. You can set up your nodes to log there automatically. Create an account and get your API key from the notes on [this page](https://app.logz.io/#/dashboard/data-sources/Filebeat), then:

```
yum install systemd-devel || echo "This will only work on RHEL-based systems."
apt-get install libsystemd-dev || echo "This will only work on Debian-based systems."

go get github.com/mheese/journalbeat
ansible-playbook -i inventory/digital_ocean.py -l remotenet logzio.yml -e LOGZIO_TOKEN=ABCDEFGHIJKLMNOPQRSTUVWXYZ012345
```

### Monitoring

You can install the DataDog agent with:

```
make datadog-install
```
