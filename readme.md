# Universal State Channels

Universal State Channels (or USC) is a platform for building state channels using centralized or decentralized judges.

*What is a state channel?* A state channel is a way for two or more parties to maintain a state (any sequence of bytes) between them, being able to keep it updated without having to trust each other.

*What are state channels used for?* The most well-known use of state channels is in payment channels, a technique which allows two parties to pay each other by adjusting the balance of an escrow account or a special blockchain transaction. Payment channels can also be linked together, allowing payments to be securely transmitted across multiple untrusted parties. Participants in the channel exchanged signed state update transactions, and can close the channel at any time with confidence that the last valid update transaction will be honored. Because only the last update transaction is sent to the judge, they can be used to create very scalable systems. [Here's](http://www.jeffcoleman.ca/state-channels/) an easy explanation of state channels.

Payment channels, and any other state chanel, can be implemented easily on top of USC with a minimum of effort. [Here's](https://github.com/jtremback/universal-state-channels/blob/master/examples/payment-channel/app.js) a 200 line Javascript example app implementing a simple payment channel.

USC Peer handles communication between the participants in the channel, and cryptographically checks all transactions for validity. The developer writing the channel's business logic interacts with a friendly HTTP API over localhost. No specialized knowledge of cryptography is required.

USC Judge is run by a trusted third party and checks the validity of a state channel's final update transaction. It also has an easy API, which the developer can use to write the business logic which reacts to the state contained in the final update transaction. For example, in a payment channel, the USC Judge could be run by a bank and permanently transfer money when the channel closed.

USC will also have blockchain adapters, which run alongside the USC Peer. Blockchain adapters relay transactions to a judge contract on a blockchain. This allows the trusted third party to be replaced by trust in a programmable blockchain, such as Ethereum, Tendermint, or Hyperledger.

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

>#### HOLD phase
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