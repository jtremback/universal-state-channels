# Lifecycle of a channel

## Opening


*peer/caller/propose_channel* - When a peer wants to start a channel:

- create and sign opening tx
- send it to counterparty
- make a new channel in a PENDING_OPEN phase


*peer/counterparty/add_channel* - When the counterparty receives the opening tx:

- check that there is not already a channel with that ID.
- verify the signature
- verify that the account and the counterparty have the same judge
- make a new channel in a PENDING_OPEN phase


*peer/caller/accept_channel* - When a peer wants to confirm a proposed channel:

- if the channel is PENDING_OPEN:
  - sign the opening tx
  - send it to the judge


*judge/peer/add_channel* - When a judge receives an opening tx:

- check the signatures
- check that there is not already a channel with that ID
- make channel in PENDING_OPEN phase with opening tx and save


*judge/caller/accept_channel* - When a judge wants to confirm a proposed channel:

- if the channel is PENDING_OPEN:
  - change channel phase to OPEN
  - sign opening tx


*peer/caller/open_channel* - When a peer finds a fully-signed opening tx from the judge:

- if the channel is PENDING_OPEN:
  - check signatures of account, counterparty, and judge
  - change channel phase to OPEN


## Updating


*peer/caller/new_update_tx* - When a peer wants to update a channel:

- if the channel is OPEN or PENDING_CLOSED
  - passed in state data
  - set sequence number to sequence number to (highest sequence number of MyProposedUpdateTx and TheirProposedUpdateTx) + 1
  - sign update tx
  - send to counterparty


*peer/counterparty/add_update_tx* - When the counterparty receives an update tx:

- if the channel is OPEN or PENDING_CLOSED
  - check that it is signed by counterparty
  - check that the sequence number is higher than (highest sequence number of MyProposedUpdateTx and TheirProposedUpdateTx)
  - save update as TheirProposedUpdateTx


*peer/caller/accept_update_tx* - When a peer wants to approve an update tx:

- if the channel is OPEN or PENDING_CLOSED
  - check if channel has an update tx to be approved
  - sign it and save it in LastFullUpdateTx
  - send it to the counterparty


## Closing

*peer/caller/close_channel* -

- if the channel is OPEN or PENDING CLOSED:
  - send the LastFullUpdateTx to the judge.


*judge/peer/add_closing_tx* - When a judge receives a closing tx:

- if the channel is OPEN:
  - check that it is signed by one of the channel's accounts
  - place channel in PENDING_CLOSE


## *judge/peer/add_update_tx* - When a judge receives an UpdateTx:

## - if the channel is OPEN:
##   - check that the UpdateTx is signed by both of the channel's accounts

##   - if the UpdateTx has a higher SequenceNumber than the channel's LastFullUpdateTx and ProposedUpdateTx
##     - replace the ProposedUpdateTx

## - if the channel is PENDING_CLOSED:
##   - check that the LastFullUpdateTx is signed by both of the channel's accounts

##   - if the channel does not have a LastFullUpdateTx or if peer's LastFullUpdateTx SequenceNumber is equal to or higher than the judge's own:
##     - save the peer's LastFullUpdateTx with the channel, overwriting any existing LastFullUpdateTx


*judge/peer/add_update_tx* - When a judge receives an UpdateTx:

- if the channel is OPEN:
  - check that the UpdateTx is signed by both of the channel's accounts

  - if the UpdateTx has a higher SequenceNumber than the channel's LastFullUpdateTx
    - replace the LastFullUpdateTx

*peer/caller/new_follow_on_tx* - When a peer wants to submit a follow-on tx:

- if the channel is OPEN:
  - add to the channel.

- if the channel is PENDING_CLOSED:
  - sign it and send it to the judge.


*judge/peer/add_follow_on_tx* - When a judge receives a follow-on tx:

- if the channel is OPEN or PENDING_CLOSED:
  - check that it is signed by one of the channel's accounts
  - save it in the follow on txs array


*peer/caller/check_final_update_tx* - When a peer sees that a judge has posted a last LastFullUpdateTx:

- if channel is OPEN or PENDING_CLOSED:
  - check that the last LastFullUpdateTx is signed by the peer, the counterparty, and the judge.

  - if judge's last LastFullUpdateTx SequenceNumber is equal to or higher than the peer's own:
    - place channel into PENDING_CLOSED if it isn't already
  - else:
    - send peer's last LastFullUpdateTx to judge


## Daemon

The usc daemon checks with the judge of every channel every once in a while. If it finds that an update tx has been posted, it calls peer/caller/check_final_update_tx



When a peer sends an update tx to the judge, the judge appends it to the UpdateTxs array and sets the ClosingTime if it isn't already set.

When a judge caller checks in and sees that there is a channel with a ClosingTime + HoldPeriod that is earlier than the current time, they can run their verification code on the UpdateTxs array and the FollowOnTxs array and close the channel if desired. If the UpdateTx with the highest SequenceNumber is not valid the judge caller is able to check an earlier one, or not.