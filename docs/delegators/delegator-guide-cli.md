<!--
order: 1
-->

# Delegator Guide (CLI)

This document contains all the necessary information for delegators to interact with the LFB through the Command-Line Interface (CLI).

It also contains instructions on how to manage accounts, restore accounts from the fundraiser and use a ledger nano device.

::: danger
**Very Important**: Please assure that you follow the steps described hereinafter
carefully, as negligence in this significant process could lead to an indefinite
loss of your base coin. Therefore, read through the following instructions in their 
entirety prior to proceeding and reach out to us in case you need support.

Please also note that you are about to interact with the LFB, a
blockchain technology containing highly experimental software. While the
blockchain has been developed in accordance to the state of the art and audited
with utmost care, we can nevertheless expect to have issues, updates and bugs.
Furthermore, interaction with blockchain technology requires
advanced technical skills and always entails risks that are outside our control.
By using the software, you confirm that you understand the inherent risks
associated with cryptographic software and that the Line Blockchain Lab Team 
may not be held liable for potential damages arising out of the use of the
software. Any use of this open source software released under the Apache 2.0 license is
done at your own risk and on a "AS IS" basis, without warranties or conditions
of any kind.
:::

Please exercise extreme caution!

## Table of Contents

- [Installing `lfb`](#installing-lfb)
- [LFB Accounts](#lfb-accounts)
    + [Restoring an Account from the Fundraiser](#restoring-an-account-from-the-fundraiser)
    + [Creating an Account](#creating-an-account)
- [Accessing a LINE Financial Blockchain](#accessing-a-line-blockchain-mainnet)
    + [Running Your Own Full-Node](#running-your-own-full-node)
    + [Connecting to a Remote Full-Node](#connecting-to-a-remote-full-node)
- [Setting Up `lfb`](#setting-up-lfb)
- [Querying the State](#querying-the-state)
- [Sending Transactions](#sending-transactions)
    + [A Note on Gas and Fees](#a-note-on-gas-and-fees)
    + [Bonding base coin and Withdrawing Rewards](#bonding-base-coin-and-withdrawing-rewards)
    + [Participating in Governance](#participating-in-governance)
    + [Signing Transactions from an Offline Computer](#signing-transactions-from-an-offline-computer)

## Installing `lfb` 

`lfb`: This is the command-line interface to interact with a `lfb` full-node. 

::: warning
**Please check that you download the latest stable release of `lfb` that is available**
:::

[**Download the binaries**]
Not available yet.

[**Install from source**](../lfb-tutorials/installation.md)

::: tip
`lfb` is used from a terminal. To open the terminal, follow these steps:
- **Windows**: `Start` > `All Programs` > `Accessories` > `Command Prompt`
- **MacOS**: `Finder` > `Applications` > `Utilities` > `Terminal`
- **Linux**: `Ctrl` + `Alt` + `T`
:::

## LFB Accounts

At the core of every LFB account, there is a seed, which takes the form of a 12 or 24-words mnemonic. From this mnemonic, it is possible to create any number of LFB accounts, i.e. pairs of private key/public key. This is called an HD wallet (see [BIP32](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki) for more information on the HD wallet specification).

```
     Account 0                         Account 1                         Account 2

+------------------+              +------------------+               +------------------+
|                  |              |                  |               |                  |
|    Address 0     |              |    Address 1     |               |    Address 2     |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Public key 0    |              |  Public key 1    |               |  Public key 2    |
|        ^         |              |        ^         |               |        ^         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        |         |              |        |         |               |        |         |
|        +         |              |        +         |               |        +         |
|  Private key 0   |              |  Private key 1   |               |  Private key 2   |
|        ^         |              |        ^         |               |        ^         |
+------------------+              +------------------+               +------------------+
         |                                 |                                  |
         |                                 |                                  |
         |                                 |                                  |
         +--------------------------------------------------------------------+
                                           |
                                           |
                                 +---------+---------+
                                 |                   |
                                 |  Mnemonic (Seed)  |
                                 |                   |
                                 +-------------------+
```

The funds stored in an account are controlled by the private key. This private key is generated using a one-way function from the mnemonic. If you lose the private key, you can retrieve it using the mnemonic. However, if you lose the mnemonic, you will lose access to all the derived private keys. Likewise, if someone gains access to your mnemonic, they gain access to all the associated accounts. 

::: danger
**Do not lose or share your 12 words with anyone. To prevent theft or loss of funds, it is best to ensure that you keep multiple copies of your mnemonic, and store it in a safe, secure place and that only you know how to access. If someone is able to gain access to your mnemonic, they will be able to gain access to your private keys and control the accounts associated with them.**
:::

The address is a public string with a human-readable prefix (e.g. `link1twsfmuj28ndph54k4nw8crwu8h9c8mh3rtx705`) that identifies your account. When someone wants to send you funds, they send it to your address. It is computationally infeasible to find the private key associated with a given address. 

### Restoring an Account from the Fundraiser

::: tip
*NOTE: This section only concerns fundraiser participants*
:::

If you participated in the fundraiser, you should be in possession of a 12-words mnemonic. Newly generated mnemonics use 24 words, but 12-word mnemonics are also compatible with all the LFB tools. 

#### On a Ledger Device

At the core of a ledger device, there is a mnemonic used to generate accounts on multiple blockchains (including a LINE Financial Blockchain). Usually, you will create a new mnemonic when you initialize your ledger device. However, it is possible to tell the ledger device to use a mnemonic provided by the user instead. Let us go ahead and see how you can input the mnemonic you obtained during the fundraiser as the seed of your ledger device. 

::: warning
*NOTE: To do this, **it is preferable to use a brand new ledger device.**. Indeed, there can be only one mnemonic per ledger device. If, however, you want to use a ledger that is already initialized with a seed, you can reset it by going in `Settings`>`Device`>`Reset All`. **Please note that this will wipe out the seed currently stored on the device. If you have not properly secured the associated mnemonic, you could lose your funds!!!***
:::

The following steps need to be performed on an un-initialized ledger device:

1. Connect your ledger device to the computer via USB
2. Press both buttons
3. Do **NOT** choose the "Config as a new device" option. Instead, choose "Restore Configuration"
4. Choose a PIN
5. Choose the 12 words option
6. Input each of the words you got during the fundraiser, in the correct order. 

Your ledger is now correctly set up with your fundraiser mnemonic! Do not lose this mnemonic! If your ledger is compromised, you can always restore a new device again using the same mnemonic.

Next, click [here](#using-a-ledger-device) to learn how to generate an account. 

#### On a Computer

::: warning
**NOTE: It is more secure to perform this action on an offline computer**
::: 

To restore an account using a fundraiser mnemonic and store the associated encrypted private key on a computer, use the following command:

```bash
lfb keys add <yourKeyName> --recover
```

- `<yourKeyName>` is the name of the account. It is a reference to the account number used to derive the key pair from the mnemonic. You will use this name to identify your account when you want to send a transaction.
- You can add the optional `--account` flag to specify the path (`0`, `1`, `2`, ...) you want to use to generate your account. By default, account `0` is generated. 

The private key of account `0` will be saved in your operating system's credentials storage.
Each time you want to send a transaction, you will need to unlock your system's credentials store.
If you lose access to your credentials storage, you can always recover the private key with the
mnemonic.

::: tip
**You may not be prompted for password each time you send a transaction since most operating systems
unlock user's credentials store upon login by default. If you want to change your credentials
store security policies please refer to your operating system manual.**
:::

### Creating an Account

To create an account, you just need to have `lfb` installed. Before creating it, you need to know where you intend to store and interact with your private keys. The best options are to store them in an offline dedicated computer or a ledger device. Storing them on your regular online computer involves more risk, since anyone who infiltrates your computer through the internet could exfiltrate your private keys and steal your funds.

#### Using a Ledger Device

::: warning
**Only use Ledger devices that you bought factory new or trust fully**
:::

When you initialize your ledger, a 24-word mnemonic is generated and stored in the device. This mnemonic is compatible with LFB and LFB accounts can be derived from it. Therefore, all you have to do is make your ledger compatible with `lfb`. To do so, you need to go through the following steps:

1. Download the Ledger Live app [here](https://www.ledger.com/pages/ledger-live). 
2. Connect your ledger via USB and update to the latest firmware
3. Go to the ledger live app store, and download the "LFB" application (this can take a while). **Note: You may have to enable `Dev Mode` in the `Settings` of Ledger Live to be able to download the "LFB" application**. 
4. Navigate to the LFB app on your ledger device

Then, to create an account, use the following command:

```bash
lfb keys add <yourAccountName> --ledger 
```

::: warning
**This command will only work while the Ledger is plugged in and unlocked**
:::

- `<yourKeyName>` is the name of the account. It is a reference to the account number used to derive the key pair from the mnemonic. You will use this name to identify your account when you want to send a transaction.
- You can add the optional `--account` flag to specify the path (`0`, `1`, `2`, ...) you want to use to generate your account. By default, account `0` is generated. 

#### Using a Computer 

::: warning
**NOTE: It is more secure to perform this action on an offline computer**
:::

To generate an account, just use the following command:

```bash
lfb keys add <yourKeyName>
```

The command will generate a 24-words mnemonic and save the private and public keys for account `0`
at the same time.
Each time you want to send a transaction, you will need to unlock your system's credentials store.
If you lose access to your credentials storage, you can always recover the private key with the
mnemonic.

::: tip
**You may not be prompted for password each time you send a transaction since most operating systems
unlock user's credentials store upon login by default. If you want to change your credentials
store security policies please refer to your operating system manual.**
:::

::: danger
**Do not lose or share your 12 words with anyone. To prevent theft or loss of funds, it is best to ensure that you keep multiple copies of your mnemonic, and store it in a safe, secure place and that only you know how to access. If someone is able to gain access to your mnemonic, they will be able to gain access to your private keys and control the accounts associated with them.**
::: 

::: warning
After you have secured your mnemonic (triple check!), you can delete bash history to ensure no one can retrieve it:

```bash
history -c
rm ~/.bash_history
```
:::

- `<yourKeyName>` is the name of the account. It is a reference to the account number used to derive the key pair from the mnemonic. You will use this name to identify your account when you want to send a transaction.
- You can add the optional `--account` flag to specify the path (`0`, `1`, `2`, ...) you want to use to generate your account. By default, account `0` is generated. 


You can generate more accounts from the same mnemonic using the following command:

```bash
lfb keys add <yourKeyName> --recover --account 1
```

This command will prompt you to input a passphrase as well as your mnemonic. Change the account number to generate a different account. 


## Accessing a LINE Financial Blockchain

In order to query the state and send transactions, you need a way to access the network. To do so, you can either run your own full-node, or connect to someone else's.

::: danger
**NOTE: Do not share your mnemonic (12 or 24 words) with anyone. The only person who should ever need to know it is you. This is especially important if you are ever approached via email or direct message by someone requesting that you share your mnemonic for any kind of blockchain services or support. No one from Line Blockchain Lab team will ever send an email that asks for you to share any kind of account credentials or your mnemonic."**.
::: 

### Running Your Own Full-Node

This is the most secure option, but comes with relatively high resource requirements. In order to run your own full-node, you need good bandwidth and at least 1TB of disk space. 

You will find the tutorial on how to install `lfb` [here](../lfb-tutorials/installation.md), and the guide to run a full-node [here](../lfb-tutorials/join-mainnet.md).

### Connecting to a Remote Full-Node

If you do not want or cannot run your own node, you can connect to someone else's full-node. You should pick an operator you trust, because a malicious operator could return  incorrect query results or censor your transactions. However, they will never be able to steal your funds, as your private keys are stored locally on your computer or ledger device. Possible options of full-node operators include validators, wallet providers or exchanges. 

In order to connect to the full-node, you will need an address of the following form: `https://77.87.106.33:26657` (*Note: This is a placeholder*). This address has to be communicated by the full-node operator you choose to trust. You will use this address in the [following section](#setting-up-lfb).

## Setting Up `lfb`

::: tip
**Before setting up `lfb`, make sure you have set up a way to [access a LINE Financial Blockchain](#accessing-a-line-blockchain-mainnet)**
:::

::: warning
**Please check that you are always using the latest stable release of `lfb`**
:::

`lfb` is the tool that enables you to interact with the node that runs on a LINE Financial Blockchain, whether you run it yourself or not. Let us set it up properly.

In order to set up `lfb`, use the following command:

```bash
lfb config <flag> <value>
```

It allows you to set a default value for each given flag. 

First, set up the address of the full-node you want to connect to:

```bash
lfb config node <host>:<port

// example: lfb config node https://77.87.106.33:26657
```

If you run your own full-node, just use `tcp://localhost:26657` as the address. 

Then, let us set the default value of the `--trust-node` flag:

```bash
lfb config trust-node false

// Set to true if you run a light-client node, false otherwise
```

Finally, let us set the `chain-id` of the blockchain we want to interact with:

```bash
lfb config chain-id lfb-2
```

## Querying the State

::: tip
**Before you can bond base coins and withdraw rewards, you need to [set up `lfb`](#setting-up-lfb)**
:::

`lfb` lets you query all relevant information from the blockchain, like account balances, amount of bonded tokens, outstanding rewards, governance proposals and more. Next is a list of the most useful commands for delegator. 

```bash
// query account balances and other account-related information
lfb query account <yourAddress>

// query the list of validators
lfb query staking validators

// query the information of a validator given their address (e.g. linkvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnjz9hqc)
lfb query staking validator <validatorAddress>

// query all delegations made from a delegator given their address (e.g. link1qyqszqgpqyqszqgpqyqszqgpqyqszqgp8apuk5)
lfb query staking delegations <delegatorAddress>

// query a specific delegation made from a delegator (e.g. link1qyqszqgpqyqszqgpqyqszqgpqyqszqgp8apuk5) to a validator (e.g. linkvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnjz9hqc) given their addresses
lfb query staking delegation <delegatorAddress> <validatorAddress>

// query the rewards of a delegator given a delegator address (e.g. link1qyqszqgpqyqszqgpqyqszqgpqyqszqgp8apuk5)
lfb query distribution rewards <delegatorAddress> 

// query all proposals currently open for depositing
lfb query gov proposals --status deposit_period

// query all proposals currently open for voting
lfb query gov proposals --status voting_period

// query a proposal given its proposalID
lfb query gov proposal <proposalID>
```

For more commands, just type:

```bash
lfb query
```

For each command, you can use the `-h` or `--help` flag to get more information.

## Sending Transactions

::: warning
On a LINE Financial Blockchain, the accepted denom is `ulink`, where `1link = 1,000,000ulink`
:::

### A Note on Gas and Fees

Transactions on a LINE Financial Blockchain need to include a transaction fee in order to be processed. This fee pays for the gas required to run the transaction. The formula is the following:

```
fees = ceil(gas * gasPrices)
```

The `gas` is dependent on the transaction. Different transaction require different amount of `gas`. The `gas` amount for a transaction is calculated as it is being processed, but there is a way to estimate it beforehand by using the `auto` value for the `gas` flag. Of course, this only gives an estimate. You can adjust this estimate with the flag `--gas-adjustment` (default `1.0`) if you want to be sure you provide enough `gas` for the transaction. For the remainder of this tutorial, we will use a `--gas-adjustment` of `1.5`.

The `gasPrice` is the price of each unit of `gas`. Each validator sets a `min-gas-price` value, and will only include transactions that have a `gasPrice` greater than their `min-gas-price`. 

The transaction `fees` are the product of `gas` and `gasPrice`. As a user, you have to input 2 out of 3. The higher the `gasPrice`/`fees`, the higher the chance that your transaction will get included in a block. 

::: tip
For mainnet, the recommended `gas-prices` is `0.025ulink`. 
::: 

### Sending Tokens

::: tip
**Before you can bond atbase coins and withdraw rewards, you need to [set up `lfb`](#setting-up-lfb) and [create an account](#creating-an-account)**
:::

::: warning
**Note: These commands need to be run on an online computer. It is more secure to perform them commands using a Ledger Nano S device. For the offline procedure, click [here](#signing-transactions-from-an-offline-computer).**
::: 

```bash
// Send a certain amount of tokens to an address
// Ex value for parameters (do not actually use these values in your tx!!): <to_address>=link1ghekyjucln7y67ntx7cf27m9dpuxxemnqk82wt <amount>=1000000ulink 
// Ex value for flags: <gasPrice>=0.025ulink

lfb tx send <to_address> <amount> --from <yourKeyName> --gas auto --gas-adjustment 1.5 --gas-prices <gasPrice>
```

### Bonding base coin and Withdrawing Rewards

::: tip
**Before you can bond base coins and withdraw rewards, you need to [set up `lfb`](#setting-up-lfb) and [create an account](#creating-an-account)**
:::

::: warning
**Before bonding base coin, please read the [delegator faq](delegator-faq.md) to understand the risk and responsibilities involved with delegating**
:::

::: warning
**Note: These commands need to be run on an online computer. It is more secure to perform them commands using a ledger device. For the offline procedure, click [here](#signing-transactions-from-an-offline-computer).**
::: 

```bash
// Bond a certain amount of base coin to a given validator
// ex value for flags: <validatorAddress>=linkvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnjz9hqc, <amountToBound>=10000000ulink, <gasPrice>=0.025ulink

lfb tx staking delegate <validatorAddress> <amountToBond> --from <delegatorKeyName> --gas auto --gas-adjustment 1.5 --gas-prices <gasPrice>


// Redelegate a certain amount of base coin from a validator to another
// Can only be used if already bonded to a validator
// Redelegation takes effect immediately, there is no waiting period to redelegate
// After a redelegation, no other redelegation can be made from the account for the next 3 weeks
// ex value for flags: <stcValidatorAddress>=linkvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnjz9hqc, <amountToRedelegate>=100000000ulink, <gasPrice>=0.025ulink

lfb tx staking redelegate <srcValidatorAddress> <destValidatorAddress> <amountToRedelegate> --from <delegatorKeyName> --gas auto --gas-adjustment 1.5 --gas-prices <gasPrice>

// Withdraw all rewards
// ex value for flag: <gasPrice>=0.025ulink

lfb tx distribution withdraw-all-rewards --from <delegatorKeyName> --gas auto --gas-adjustment 1.5 --gas-prices <gasPrice>


// Unbond a certain amount of base coin from a given validator 
// You will have to wait 3 weeks before your base coin are fully unbonded and transferrable 
// ex value for flags: <validatorAddress>=linkvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnjz9hqc, <amountToUnbound>=10000000ulink, <gasPrice>=0.025ulink

lfb tx staking unbond <validatorAddress> <amountToUnbond> --from <delegatorKeyName> --gas auto --gas-adjustment 1.5 --gas-prices <gasPrice>
```

::: warning
**If you use a connected Ledger, you will be asked to confirm the transaction on the device before it is signed and broadcast to the network. Note that the command will only work while the Ledger is plugged in and unlocked.**
::: 

To confirm that your transaction went through, you can use the following queries:

```bash
// your balance should change after you bond base coin or withdraw rewards
lfb query account

// you should have delegations after you bond base coin
lfb query staking delegations <delegatorAddress>

// this returns your tx if it has been included
// use the tx hash that was displayed when you created the tx
lfb query tx <txHash>

```

Double check with a block explorer if you interact with the network through a trusted full-node. 

## Participating in Governance

#### Primer on Governance

A LINE Financial Blockchain has a built-in governance system that lets bonded base coin holders vote on proposals. There are three types of proposal:

- `Text Proposals`: These are the most basic type of proposals. They can be used to get the opinion of the network on a given topic. 
- `Parameter Proposals`: These are used to update the value of an existing parameter.
- `Software Upgrade Proposal`: These are used to propose an upgrade of the Hub's software.

Any base coin holder can submit a proposal. In order for the proposal to be open for voting, it needs to come with a `deposit` that is greater than a parameter called `minDeposit`. The `deposit` need not be provided in its entirety by the submitter. If the initial proposer's `deposit` is not sufficient, the proposal enters the `deposit_period` status. Then, any base coin holder can increase the deposit by sending a `depositTx`. 

Once the `deposit` reaches `minDeposit`, the proposal enters the `voting_period`, which lasts 2 weeks. Any **bonded** base coin holder can then cast a vote on this proposal. The options are `Yes`, `No`, `NoWithVeto` and `Abstain`. The weight of the vote is based on the amount of bonded base coin of the sender. If they don't vote, delegator inherit the vote of their validator. However, delegators can override their validator's vote by sending a vote themselves. 

At the end of the voting period, the proposal is accepted if there are more than 50% `Yes` votes (excluding `Abstain ` votes) and less than 33.33% of `NoWithVeto` votes (excluding `Abstain` votes).

#### In Practice

::: tip
**Before you can bond base coins and withdraw rewards, you need to [bond base coin](#bonding-base-coin-and-withdrawing-rewards)**
:::

::: warning
**Note: These commands need to be run on an online computer. It is more secure to perform them commands using a ledger device. For the offline procedure, click [here](#signing-transactions-from-an-offline-computer).**
::: 

```bash
// Submit a Proposal
// <type>=text/parameter_change/software_upgrade
// ex value for flag: <gasPrice>=0.025ulink

lfb tx gov submit-proposal --title "Test Proposal" --description "My awesome proposal" --type <type> --deposit=10000000ulink --gas auto --gas-adjustment 1.5 --gas-prices <gasPrice> --from <delegatorKeyName>

// Increase deposit of a proposal
// Retrieve proposalID from lfb query gov proposals --status deposit_period
// ex value for parameter: <deposit>=10000000ulink

lfb tx gov deposit <proposalID> <deposit> --gas auto --gas-adjustment 1.5 --gas-prices <gasPrice> --from <delegatorKeyName>

// Vote on a proposal
// Retrieve proposalID from lfb query gov proposals --status voting_period 
// <option>=yes/no/no_with_veto/abstain

lfb tx gov vote <proposalID> <option> --gas auto --gas-adjustment 1.5 --gas-prices <gasPrice> --from <delegatorKeyName>
```

### Signing Transactions From an Offline Computer

If you do not have a ledger device and want to interact with your private key on an offline computer, you can use the following procedure. First, generate an unsigned transaction on an **online computer** with the following command (example with a bonding transaction):

```bash
// Bond base coin 
// ex value for flags: <amountToBound>=10000000ulink, <bech32AddressOfValidator>=linkvaloper1ghekyjucln7y67ntx7cf27m9dpuxxemnjz9hqc, <gasPrice>=0.025ulink, <delegatorAddress>=link1qyqszqgpqyqszqgpqyqszqgpqyqszqgp8apuk5

lfb tx staking delegate <validatorAddress> <amountToBond> --from <delegatorAddress> --gas auto --gas-adjustment 1.5 --gas-prices <gasPrice> --generate-only > unsignedTX.json
```

In order to sign, you will also need the `chain-id`, `account-number` and `sequence`. The `chain-id` is a unique identifier for the blockchain on which you are submitting the transaction. The `account-number` is an identifier generated when your account first receives funds. The `sequence` number is used to keep track of the number of transactions you have sent and prevent replay attacks.

Get the chain-id from the genesis file (`lfb-2`), and the two other fields using the account query:

```bash
lfb query account <yourAddress> --chain-id lfb-2
```

Then, copy `unsignedTx.json` and transfer it (e.g. via USB) to the offline computer. If it is not done already, [create an account on the offline computer](#using-a-computer). For additional security, you can double check the parameters of your transaction before signing it using the following command:

```bash
cat unsignedTx.json
```

Now, sign the transaction using the following command. You will need the `chain-id`, `sequence` and `account-number` obtained earlier:

```bash
lfb tx sign unsignedTx.json --from <delegatorKeyName> --offline --chain-id lfb-2 --sequence <sequence> --account-number <account-number> > signedTx.json
```

Copy `signedTx.json` and transfer it back to the online computer. Finally, use the following command to broadcast the transaction:

```bash
lfb tx broadcast signedTx.json
```
