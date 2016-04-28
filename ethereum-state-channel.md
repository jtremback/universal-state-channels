The ethereum blockchain must do a few things with a state channel. 

1. Opening the channel. To open the channel, some state is changed on the blockchain (in a payment channel, this consists of putting some tokens in escrow). A channel oracle is set up to determine report on the state in the channel and whether administer the channel closing rules.

2. Closing the channel. When a state channel oracle receives a valid state update from one of the channel participants, it enters a challenge period during which another channel participant can submit a state update with a higher sequence number. After the challenge period ends, the valid state update with the highest sequence number is accepted as the final state. What constitutes a valid state update will be discussed below.  

3. Taking action on the channel. After the channel is closed, contracts on the blockchain call the oracle and learn that it is closed. They can then change some other state on the blockchain (in a payment channel, this consists of sending the tokens in escrow back to the participants according to the channel's last valid state).

The standardized part of the channel that we describe here is the channel oracle responsible for step 2. The contract using the state channel oracle is responsible for steps 1 and 3.

A state channel oracle is a channel that gives a result when called:

```
{
    phase: 'open' | 'challenge' | 'closed'
    closingBlock: <integer>
    stateUpdate: {
        state: <bytes>
        sequence: <integer>
    }
}
```

A state channel oracle reports whether a channel is open, in the challenge period, or closed. If the channel is in the challenge period or closed, it also reports the block that the channel closes on, and the last valid state update.

## State update validity
During the `open` and `challenge` phases the state channel oracle accepts state updates. The state channel oracle may have several requirements to consider a submitted state update valid. The simplest requirements are:

- State updates must be signed by both channel participants.
- Each successive state update must have a higher `sequence` than the last.
- State updates cannot be submitted after the channel is closed.

However, there will always be state updates which are valid according to the oracle, but contain state which is not valid to the contract using the state channel. For example, a payment channel whose balance goes below zero.

What happens if the last valid state as reported by the state channel oracle is invalid for an application-specific reason? Presumably, the application would simply need to roll back the state that existed before opening the channel (in a payment channel, the escrow amounts would be refunded in full). This may be a good solution. It's the responsibility of the channel participants to ensure that they do not accidently sign an invalid state update.

## Added features
These things are key to the functionality and security of state channels, but might not be built into the core of the standard. 

### Hashlock
A hashlocked state update includes a hash. With a hashlocked state update, all or some of the state in the update is not considered valid unless the channel also receives the preimage for a given hash. 

We could ignore this use in our specification and allow this abstraction to remain in the channel state. The contract consulting the state channel oracle would determine whether a hashlock is valid, having also received the preimage from one of the participants.

Another way to do this is with evidence: unsigned data stored by the state channel oracle. The contract consulting the state channel oracle would determine whether a hashlock is valid, after consulting the preimage from one of the participants.

Hashlocks could also be built into the state channel specification. In this case, we'd want to allow the state channel oracle to call out to an external contract as part of evaluating a hashlock to allow the use of different hashing algorithms or other conditions such as signatures.

### Expiration block
State updates can include be given an expiration block after which they are no longer valid. If Alice receives such a state update, she can send her signature to Bob. In exchange Bob sends her a new update with the same state but a later expiration block. 

Without this, Alice can refuse to submit the state update, forcing Bob to submit the previous state update to close the channel. This allows us to attach some penalty to submitting an old state update.