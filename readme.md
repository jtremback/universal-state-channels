# Universal State Channels

Universal State Channels (or USC) is a platform for building state channels using centralized or decentralized judges.

*What is a state channel?* A state channel is a way for two or more parties to maintain a state (any sequence of bytes) between them, being able to keep it updated without having to trust each other. A "judge", either a third party trusted by all channel participants, or a contract on a blockchain, can verify this state later and take actions on it.

*What are state channels used for?* The most well-known use of state channels is in payment channels, a technique which allows two parties to pay each other by adjusting the balance of an escrow account or a special blockchain transaction. Payment channels can also be linked together, allowing payments to be securely transmitted across multiple untrusted parties. Participants in the channel exchanged signed state update transactions, and can close the channel at any time with confidence that the last valid update transaction will be honored. Because only the last update transaction is sent to the judge, they can be used to create very scalable systems. [Here's](http://www.jeffcoleman.ca/state-channels/) an easy explanation of state channels.

Payment channels, and any other state chanel, can be implemented easily on top of USC with a minimum of effort. [Here's](https://github.com/jtremback/universal-state-channels/blob/master/examples/payment-channel/app.js) a 200 line Javascript example app implementing a simple payment channel.

USC Peer handles communication between the participants in the channel, and cryptographically checks all transactions for validity. The developer writing the channel's business logic interacts with a friendly HTTP API over localhost. No specialized knowledge of cryptography is required.

USC Judge is run by a trusted third party and checks the validity of a state channel's final update transaction. It also has an easy API, which the developer can use to write the business logic which reacts to the state contained in the final update transaction. For example, in a payment channel, the USC Judge could be run by a bank and permanently transfer money when the channel closed.

USC will also have blockchain adapters, which run alongside the USC Peer. Blockchain adapters relay transactions to a judge contract on a blockchain. This allows the trusted third party to be replaced by trust in a programmable blockchain, such as Ethereum, Tendermint, or Hyperledger. 

## Lifecycle

Let's take the imaginary example of a state channel being used by Alice and Bob to decide which color to paint the bike shed. 

Peers: Alice, Bob
Judge: Acme Shed Painting inc.

First, Alice prepares an `opening transaction`. It tells Acme to paint the shed a color that Alice and Bob will agree on in a future `update transaction`.

```
OpeningTx
  channel_id: "foo"
  pubkeys: [<Alice's pubkey>, <Bob's pubkey>]
  state: "We want you to paint the bikeshed this color:"
  hold_period: 4
```
She signs it, and sends it to Bob. If Bob thinks it's OK, he signs it as well, and sends it to Acme.

Acme looks at the `OpeningTx`, and if Acme wishes to act as judge, they sign it. Now the channel is open.

Alice and Bob haggle over what color the shed should be. Each time they change their minds, they sign a new `UpdateTx` reflecting the current agreement on shed color. Note that these `UpdateTx`s do not necessarily need to be sent to Acme. 

```
UpdateTx
  channel_id: "foo"
  sequence_number: 1
  final: false
  state: "Red"
```

When Alice or Bob want to close the channel, one of them sends an `UpdateTx` and a `ClosingTx` to Acme. Acme records the  closing time, and continues to accept further `UpdateTx`s for the channel. After the hold period has elapsed, Acme accepts the valid `UpdateTx` with the highest sequence number.

Lets say Bob tries to close the channel with an old `UpdateTx`.
As long as Alice finds out about this before the hold period is over, she can send the latest `UpdateTx` to correct the record.

If Alice and Bob both agree that they would like to close the channel and get the shed painted, they set `final` to `true`. This means that either of them can submit the transaction and the channel will close without waiting for the hold period to be over.

Let's say that Alice wants to give Bob permission to paint the bike shed blue, but only if he can get Al Gore's sign off on the color choice. She signs this `UpdateTx` and sends it to Bob.

```
UpdateTx
  channel_id: "foo"
  sequence_number: 34
  final: false
  state: "Blue, if Al Gore is OK with it."
```

If Bob gets Al Gore's signature on the color choice, he can close the channel with this `UpdateTx` and then submit a `FollowOnTx` with the signature.

```
FollowOnTx
  channel_id: "foo"
  state: "I think the shed should be Blue" + <Al Gore's signature>
```

This `FollowOnTx` only needs to be signed by one of the channel participants, so Bob can submit it without Alice's help.


## Architecture

There are two main codebases, `Peer` and `Judge`. `Peer` is run by Alice and Bob, and handles the creation, signing, and exchange of `OpeningTx`s and `UpdateTx`s, etc. `Judge` is run by Acme Shed Painting and checks the validity of `OpeningTx`s and `UpdateTx`s and closes the channel after doing the hold period etc. Both of these codebases interact with other applications on the computer over an HTTP API. This way, USC can be used to create a variety of different channels, from a shed painting channel to a payment channel. 


## Blockchain adapter

With a blockchain adapter, and a judge contract, USC can be made to work on a programmable blockchain like Ethereum.

The blockchain adapter is a piece of software that lives on the channel participant's computers. It speaks to the `Peer` using the same protocol that a third party `Judge` would, but it relays the transactions onto a blockchain so that the judge contract can act on them.

Alice creates an `OpeningTx`, and sends it to Bob. This step is exactly the same as it is with a centralized judge. Bob gets the `OpeningTx`, checks that he is happy with the state, and signs it as well. Bob then sends the `OpeningTx` to his blockchain adapter, as if it was a third party judge. Bob's blockchain adapter packages the `OpeningTx` into a transaction for the blockchain in question. This transaction is addressed to the on-blockchain judge contract. When the judge contract receives the `OpeningTx`, it checks that the transaction and its signatures are valid, and then calls an executive contract with the state. The executive contract is not part of USC, it evaluates the state and takes some action upon it. For example, a contract that holds some currency in escrow to create a payment channel. If the executive contract finds that the state is valid, it performs whatever actions it is supposed to, and calls back to the judge contract to open the channel.

Once the channel is open, the judge contract accepts `UpdateTx`s from either Alice or Bob. It's important to note that while the judge contract will accept and store `UpdateTx`s, it does not close the channel when receiving one. The vast majority of `UpdateTx`s are exchanged only between Alice and Bob. When the judge contract gets an `UpdateTx`, it checks the signatures and the sequence number and saves the `UpdateTx`. Only `UpdateTx`s with a sequence number higher than the last are accepted.

`FollowOnTx`s are also accepted. `FollowOnTx`s contain some state and only need to have a valid signature from either Alice or Bob.

At some point, Alice or Bob submits their last `UpdateTx` and a `ClosingTx`. The `ClosingTx` contains no information, but serves to identify the time of closing. The judge contract contiues to accept and store valid `UpdateTx`s. At some point in the future, Alice or Bob send a `FinalizeTx`. When the judge contract receives the `FinalizeTx`, it checks that either the hold period has elapsed, or `final` is `true`, and calls the executive contract. The executive contract evaluates both the state of the `UpdateTx` with the highest sequence number, and any `FollowOnTx`s. The executive contract may take some action on it (i.e. moving tokens for a payment channel). If the executive contract finds that the state is not valid, it attempts to validate the `UpdateTx` state with the second highest sequence number, and so on. If the are no valid `UpdateTx`s, the channel is considered to be cancelled, and the executive contract may take some action to reverse the actions it performed when opening the channel. The judge contract also signs and stores the `UpdateTx`. This allows the blockchain adapter to easily tell the USC `Peer` and its client software that the channel is closed. 


## HTTP Peer API

### Accounts

Accounts correspond to identities known by a third party judge or a blockchain. Accounts embed the information for their judge.

#### List All Accounts

`accounts` returns a list of all accounts stored in the USC Peer.

GET `https://localhost:4456/accounts`

*Example response:*

```json
[
  {
    "name": "AC7739 at SFFCU",
    "pubkey": "R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=",
    "privkey": "k4NkO3BNxNN8qsdPvsKv9AEJMP_IqIqluy77HLcN1gVHmVVWzzYzzSLk6lHfr1K0mpodLrUt34_NMJ9L7TPIAA==",
    "judge": {
      "name": "San Francisco Federal Credit Union",
      "pubkey": "xcYNnNW1oA9pB0LeQg_UCKw3FC8itnVq1csGrHdCV6o=",
      "address": "https://sanfranciscofcu.com/channels/"
    }
  },
  ...
]
```

#### Accounts by Pubkey

`accounts_by_pubkey` returns account with the specified pubkey.

GET `https://localhost:4456/accounts_by_id/R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=`

Response: An account, see above.



### Counterparties

#### List all counterparties

`counterparties` returns a list of all accounts of counterparties known to the USC Peer. Counterparties are peers that USC can start channels with. Counterparties embed the information for their judge.

GET `https://localhost:4456/counterparties`

Response:

```json
[
  {
    "name": "AC2346 at SFFCU",
    "pubkey": "R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=",
    "judge": {
      "name": "San Francisco Federal Credit Union",
      "pubkey": "xcYNnNW1oA9pB0LeQg_UCKw3FC8itnVq1csGrHdCV6o=",
      "address": "https://sanfranciscofcu.com/channels/"
    }
  },
  ...
]
```

#### Counterparties by pubkey

`counterparties/pubkey/<pubkey>` returns the counterparty with the specified pubkey.

GET `https://localhost:4456/counterparties/pubkey/R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=`

Response: A counterparty, see above.



### Channels

Channels embed information about their account, their counterparty, and their judge.

#### List all channels

`channels` returns a list of all channels which accounts on the USC Peer participate in.

Request: GET `https://localhost:4456/channels`

Response:

```json
[
  {
    "channelId": "8789678",
    "phase": "OPEN",
    "openingTx": {
      "channelId": "8789678",
      "pubkeys": [
        "R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=",
        "prNVb9C260wELZ3RYmrJ9TsZ_2NCGYcUBVZSSGHUsYQ="
      ],
      "state": "{\"R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=\":100,\"prNVb9C260wELZ3RYmrJ9TsZ_2NCGYcUBVZSSGHUsYQ=\":100}"
    },
    "lastFullUpdateTx": {
      "channelId": "8789678",
      "sequenceNumber": 7,
      "fast": false,
      "state": "{\"R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=\":105,\"prNVb9C260wELZ3RYmrJ9TsZ_2NCGYcUBVZSSGHUsYQ=\":95}"
    },
    "myProposedUpdateTx": {
      "channelId": "8789678",
      "sequenceNumber": 7,
      "fast": false,
      "state": "{\"R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=\":105,\"prNVb9C260wELZ3RYmrJ9TsZ_2NCGYcUBVZSSGHUsYQ=\":95}"
    },
    "theirProposedUpdateTx": {
      "channelId": "8789678",
      "sequenceNumber": 6,
      "fast": false,
      "state": "{\"R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=\":102,\"prNVb9C260wELZ3RYmrJ9TsZ_2NCGYcUBVZSSGHUsYQ=\":98}"
    },
    "followOnTxs": [],
    "judge": {
      "name": "San Francisco Federal Credit Union",
      "pubkey": "xcYNnNW1oA9pB0LeQg_UCKw3FC8itnVq1csGrHdCV6o=",
      "address": "https://sanfranciscofcu.com/channels/"
    },
    "account": {
      "name": "AC7739 at SFFCU",
      "pubkey": "R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=",
      "privkey": "k4NkO3BNxNN8qsdPvsKv9AEJMP_IqIqluy77HLcN1gVHmVVWzzYzzSLk6lHfr1K0mpodLrUt34_NMJ9L7TPIAA==",
      "judge": {
        "name": "San Francisco Federal Credit Union",
        "pubkey": "xcYNnNW1oA9pB0LeQg_UCKw3FC8itnVq1csGrHdCV6o=",
        "address": "https://sanfranciscofcu.com/channels/"
      }
    },
    "counterparty": {
      "name": "AC2346 at SFFCU",
      "pubkey": "prNVb9C260wELZ3RYmrJ9TsZ_2NCGYcUBVZSSGHUsYQ=",
      "judge": {
        "name": "San Francisco Federal Credit Union",
        "pubkey": "xcYNnNW1oA9pB0LeQg_UCKw3FC8itnVq1csGrHdCV6o=",
        "address": "https://sanfranciscofcu.com/channels/"
      }
    }
  },
  ...
]
```

### Channels by Id

`channels/<channelId>` returns the channel with the specified channelId.

Request:

`GET https://localhost:4456/channels_by_id/8789678`

Response: A channel, see above.

## Channel lifecycle

[`propose_channel`](#propose-channel) ->

>#### PROPOSED phase
>- [`accept_channel`](#accept-channel)
>- [`reject_channel`](#reject-channel)

>#### OPEN phase
>- [`cancel_channel`](#cancel-channel)
>- [`propose_update_tx`](#propose-update-tx)
>- [`accept_update_tx`](#accept-update-tx)
>- [`reject_update_tx`](#reject-update-tx)
>- [`close_channel`](#close-channel)
>- [`check_channel`](#check-channel)

>#### CLOSED phase
>- (no actions possible)


### New Channel

`propose_channel` creates a new channel in PENDING_OPEN phase, signs it, and sends it to the counterparty.

Request:

```json
POST `https://localhost:4456/propose_channel`

{
  "channelId": "8789678",
  "accountPubkey": "R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=",
  "counterpartyPubkey": "prNVb9C260wELZ3RYmrJ9TsZ_2NCGYcUBVZSSGHUsYQ=",
  "state": "{\"R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=\":100,\"prNVb9C260wELZ3RYmrJ9TsZ_2NCGYcUBVZSSGHUsYQ=\":100}",
  "holdPeriod": 86400000
}
```

Response: A new channel, see above.


### Accept channel

`accept_channel` is called on a channel that is in PENDING_OPEN phase. The channel is signed, saved, and sent to the judge.

Request:

```json
POST `https://localhost:4456/accept_channel`

{
  "channelId": "8789678"
}
```

Response: A new channel, see above.


### Reject channel

`reject_channel` is called on a channel that is in PENDING_OPEN phase. The channel is deleted and the counterparty is notified of the rejection.

Request:

```json
POST `https://localhost:4456/reject_channel`

{
  "channelId": "8789678"
}
```

Response: `200 OK`


### Cancel channel

`cancel_channel` is called on a channel that is in OPEN phase, but has not yet had an update transaction posted. It sends a cancellation transaction to instruct the judge to close the channel with the state in the opening tx. The cancellation can be overridden during the hold period by any update tx.

Request:

```json
POST `https://localhost:4456/cancel_channel`

{
  "channelId": "8789678"
}
```

Response: `200 OK`


### Propose update tx

`propose_update_tx` is one of USC's key calls. It makes a transaction that updates the channel's state, signs it, and sends it to the counterparty.

Request:

```json
POST `https://localhost:4456/propose_update_tx`

{
  "channelId": "8789678",
  "state": "{\"R5lVVs82M80i5OpR369StJqaHS61Ld-PzTCfS-0zyAA=\":105,\"prNVb9C260wELZ3RYmrJ9TsZ_2NCGYcUBVZSSGHUsYQ=\":95}"
}
```

Response: `200 OK`


### Accept update tx

`accept_update_tx` is called to accept the update tx saved in a channel's `theirProposedUpdateTx`. The update tx is signed, replaces the channel's `lastFullUpdateTx` and is sent back to the counterparty.

```json
POST `https://localhost:4456/accept_update_tx`

{
  "channelId": "8789678"
}
```

Response: `200 OK`


### Reject update tx

`reject_update_tx` is called to reject an update tx. The counterparty is informed that the update tx saved in `theirProposedUpdateTx` will never be accepted.

```json
POST `https://localhost:4456/reject_update_tx`

{
  "channelId": "8789678"
}
```

Response: `200 OK`


### Close channel

`close_channel` sends the channel's `lastFullUpdateTx` to the judge, putting the channel into PENDING_CLOSE and starting the hold period.

```json
POST `https://localhost:4456/close_channel`

{
  "channelId": "8789678"
}
```

Response: `200 OK`


### Check channel for cheating

`check_channel` is possibly USC's most important call. This must be called at least once per hold period, the entire time the channel is open. It checks if the counterparty has tried to cheat by posting an old update tx. If so, it sends the judge the correct `lastFullUpdateTx`.

```json
POST `https://localhost:4456/check_channel`

{
  "channelId": "8789678"
}
```

Response: `200 OK`