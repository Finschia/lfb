<!--
order: 2
-->

# Delegator FAQ

## What is a delegator?

People that cannot or do not want to operate [validator nodes](..//validators/overview.md) can still participate in the staking process as delegators. Indeed, validators are not chosen based on their self-delegated stake but based on their total stake, which is the sum of their self-delegated stake and of the stake that is delegated to them. This is an important property, as it makes delegators a safeguard against validators that exhibit bad behavior. If a validator misbehaves, their delegators will move their Atoms away from them, thereby reducing their stake. Eventually, if a validator's stake falls under the top 125 addresses with highest stake, they will exit the validator set.

**Delegators share the revenue of their validators, but they also share the risks.** In terms of revenue, validators and delegators differ in that validators can apply a commission on the revenue that goes to their delegator before it is distributed. This commission is known to delegators beforehand and can only change according to predefined constraints (see [section](#choosing-a-validator) below). In terms of risk, delegators' base coins can be slashed if their validator misbehaves. For more, see [Risks](#risks) section.

To become delegators, base coin holders need to send a ["Delegate transaction"](./delegator-guide-cli.md#sending-transactions) where they specify how many base coins they want to bond and to which validator. A list of validator candidates will be displayed in Cosmos Hub explorers. Later, if a delegator wants to unbond part or all of their stake, they needs to send an "Unbond transaction". From there, the delegator will have to wait 3 weeks to retrieve their base coins. Delegators can also send a "Rebond Transaction" to switch from one validator to another, without having to go through the 3 weeks waiting period. 

For a practical guide on how to become a delegator, click [here](./delegator-guide-cli.md).

<--
It is necessary to write when a public blockchain is released.
-->
TBD
