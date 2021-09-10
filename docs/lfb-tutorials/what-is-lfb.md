<!--
order: 1
-->

# What is lfb?

`lfb` is the name of the LBM SDK application for a LINE Financial Blockchain. It comes with 2 main entrypoints:

- `lfb`: The LFB Daemon and command-line interface (CLI). runs a full-node of the `lfb` application.

`lfb` is built on the LBM SDK using the following modules:

- `x/auth`: Accounts and signatures.
- `x/bank`: Token transfers.
- `x/staking`: Staking logic.
- `x/mint`: Inflation logic.
- `x/distribution`: Fee distribution logic.
- `x/slashing`: Slashing logic.
- `x/gov`: Governance logic.
- `x/ibc`: Inter-blockchain transfers.
- `x/params`: Handles app-level parameters.

About a LINE Financial Blockchain: A LINE Financial Blockchain is a blockchain mainnet network using LFB. Any LINE Financial Blockchain can connects to each other via IBC, it automatically gains access to all the other blockchains that are connected to it. A LINE Financial Blockchain is a public Proof-of-Stake chain. Its staking token is called the Link.

Next, learn how to [install LFB](./installation.md).
