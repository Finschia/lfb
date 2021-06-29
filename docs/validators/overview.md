<!--
order: 1
-->

# Validators Overview

## Introduction

The [LINE Financial Blockchain](../README.md) is based on [Ostracon](https://github.com/line/ostracon/tree/main/docs/introduction), which relies on a set of validators that are responsible for committing new blocks in the blockchain. These validators participate in the consensus protocol by broadcasting votes which contain cryptographic signatures signed by each validator's private key.

Validator candidates can bond their own base coins and have base coins ["delegated"](../delegators/delegator-guide-cli.md), or staked, to them by token holders. A LINE Financial Blockchain can have over 100 validators. The validators are determined by who has the most stake delegated to them — the top N validator candidates with the most stake will become LFB validators.

Validators and their delegators will earn base coins as block provisions and tokens as transaction fees through execution of the Ostracon consensus protocol. Initially, transaction fees will be paid in base coins but in the future, any token in the LFB ecosystem will be valid as fee tender if it is whitelisted by governance. Note that validators can set commission on the fees their delegators receive as additional incentive.

If validators double sign, are frequently offline or do not participate in governance, their staked base coins (including base coins of users that delegated to them) can be slashed. The penalty depends on the severity of the violation.

## Hardware

There currently exists no appropriate cloud solution for validator key management. This may change in 2018 when cloud SGX becomes more widely available. For this reason, validators must set up a physical operation secured with restricted access. A good starting place, for example, would be co-locating in secure data centers.

Validators should expect to equip their datacenter location with redundant power, connectivity, and storage backups. Expect to have several redundant networking boxes for fiber, firewall and switching and then small servers with redundant hard drive and failover. Hardware can be on the low end of datacenter gear to start out with.

We anticipate that network requirements will be low initially. The current testnet requires minimal resources. Then bandwidth, CPU and memory requirements will rise as the network grows. Large hard drives are recommended for storing years of blockchain history.

## Seek Legal Advice

Seek legal advice if you intend to run a Validator.
