package wallet

import (
	"bytes"
	"crypto/rand"
	"errors"
	"github.com/agl/ed25519"
	"github.com/golang/protobuf/proto"
	"github.com/jtremback/upc-core/wallet/schema"
	"github.com/jtremback/upc-core/wire"
	"io"
)

// Phases of a Tx
// Created
// Confirmed
// Verified

// Channel is used to allow us to attach methods to schema.Channel
type Channel schema.Channel
type Account schema.Account
type EscrowProvider schema.EscrowProvider

func sliceTo64Byte(slice []byte) *[64]byte {
	var array [64]byte
	copy(array[:], slice[:64])
	return &array
}

func sliceTo32Byte(slice []byte) *[32]byte {
	var array [32]byte
	copy(array[:], slice[:32])
	return &array
}

func randomBytes(c uint) ([]byte, error) {
	b := make([]byte, c)
	n, err := io.ReadFull(rand.Reader, b)
	if n != len(b) || err != nil {
		return nil, err
	}
	return b, nil
}

// NewOpeningTxProposal assembles an OpeningTx
func (acct *Account) NewOpeningTxProposal(
	counterparties []*Account,
	state []byte,
	holdPeriod uint32,
) (*wire.OpeningTx, error) {
	b, err := randomBytes(32)
	chID := string(b)
	if err != nil {
		return nil, err
	}

	var pubkeys [][]byte
	for i, a := range append([]*Account{acct}, counterparties...) {
		pubkeys[i] = a.Pubkey
	}

	return &wire.OpeningTx{
		ChannelId:  chID,
		Pubkeys:    pubkeys,
		State:      state,
		HoldPeriod: holdPeriod,
	}, nil
}

// SignOpeningTxProposal signs and serializes an opening transaction
func SignOpeningTxProposal(otx *wire.OpeningTx, acct *schema.Account) (*wire.Envelope, error) {
	// Serialize opening transaction
	data, err := proto.Marshal(otx)
	if err != nil {
		return nil, err
	}

	// Make new envelope
	return &wire.Envelope{
		Type:       wire.Envelope_OpeningTxProposal,
		Payload:    data,
		Signatures: [][]byte{ed25519.Sign(sliceTo64Byte(acct.Privkey), data)[:]},
	}, nil
}

// ConfirmOpeningTx checks if a partially-signed OpeningTxProposal has the correct
// signature from Pubkeys[0], and calls fn to check that the state is correct. If both are ok, it signs the OpeningTx.
func (acct *Account) ConfirmOpeningTx(ev *wire.Envelope, fn func([][]byte, []byte) error) (*wire.Envelope, error) {
	otx := &wire.OpeningTx{}
	err := proto.Unmarshal(ev.Payload, otx)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(sliceTo32Byte(otx.Pubkeys[0]), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
		return nil, errors.New("signature 0 invalid")
	}

	ev.Signatures[1] = ed25519.Sign(sliceTo64Byte(acct.Privkey), ev.Payload)[:]
	ev.Type = wire.Envelope_OpeningTx

	return ev, nil
}

// NewChannel creates a new Channel from an Envelope containing an opening transaction,
// an Account and a Peer.
func (acct *Account) NewChannel(ev *wire.Envelope, counterparties []*Account) (*Channel, error) {
	otx := &wire.OpeningTx{}
	err := proto.Unmarshal(ev.Payload, otx)
	if err != nil {
		return nil, err
	}

	// var me uint32
	// if bytes.Compare(acct.Pubkey, otx.Pubkey1) == 0 &&
	// 	bytes.Compare(peer.Pubkey, otx.Pubkey2) == 0 {
	// 	me = 1
	// } else if bytes.Compare(acct.Pubkey, otx.Pubkey2) == 0 &&
	// 	bytes.Compare(peer.Pubkey, otx.Pubkey1) == 0 {
	// 	me = 2
	// } else {
	// 	return nil, errors.New("peer or account public keys do not match opening transaction")
	// }

	var me uint32
	for i, k := range otx.Pubkeys {
		if bytes.Compare(acct.Pubkey, k) == 0 {
			me = uint32(i)
		}
	}

	ch := &Channel{
		ChannelId:         otx.ChannelId,
		OpeningTx:         otx,
		OpeningTxEnvelope: ev,
		Me:                me,
		Accounts:          append([]*Account{acct}, counterparties...),
		Phase:             schema.Channel_Open,
	}

	return ch, nil
}

// NewUpdateTxProposal makes a new UpdateTx on Channel with NetTransfer changed by amount.
func (ch *Channel) NewUpdateTxProposal(state []byte) (*wire.UpdateTx, error) {
	lst := ch.LastUpdateTx
	var seq uint32
	if lst != nil {
		seq = lst.SequenceNumber + 1
	}

	// Make new update transaction
	return &wire.UpdateTx{
		ChannelId:      ch.ChannelId,
		State:          state,
		SequenceNumber: seq,
		Fast:           false,
	}, nil
}

// SignUpdateTxProposal signs an update proposal and puts it in an envelope
func (ch *Channel) SignUpdateTxProposal(utx *wire.UpdateTx) (*wire.Envelope, error) {
	// Serialize update transaction
	data, err := proto.Marshal(utx)
	if err != nil {
		return nil, err
	}

	// Make new envelope
	ev := wire.Envelope{
		Type:       wire.Envelope_UpdateTxProposal,
		Payload:    data,
		Signature1: ed25519.Sign(sliceTo64Byte(ch.Account.Privkey), data)[:],
	}

	return &ev, nil
}

// ConfirmUpdateTx takes an Envelope containing a UpdateTx with one
// signature and checks the signature, as well as the sequence number. It also
// calls fn to verify the state. If all factors are correct, it signs the UpdateTx
// and puts it in an envelope.
func (ch *Channel) ConfirmUpdateTx(ev *wire.Envelope, fn func([]byte, []byte) error) error {
	if ch.Phase != schema.Channel_Open {
		return errors.New("channel must be open")
	}

	// Read signature from correct slot
	// Copy signature and pubkey
	switch ch.Me {
	case 1:
		if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey2), ev.Payload, sliceTo64Byte(ev.Signature2)) {
			return errors.New("invalid signature")
		}
	case 2:
		if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey1), ev.Payload, sliceTo64Byte(ev.Signature1)) {
			return errors.New("invalid signature")
		}
	}

	utx := wire.UpdateTx{}
	err := proto.Unmarshal(ev.Payload, &utx)
	if err != nil {
		return err
	}

	// Check ChannelId
	if utx.ChannelId != ch.OpeningTx.ChannelId {
		return errors.New("ChannelId does not match")
	}

	lst := ch.LastUpdateTx

	if lst != nil {
		// Check last sequence number
		if lst.SequenceNumber+1 != utx.SequenceNumber {
			return errors.New("invalid sequence number")
		}
	}

	// Check state
	err = fn(ch.OpeningTx.State, utx.State)
	if err != nil {
		return err
	}

	return nil
}

// VerifyOpeningTx checks the signatures and state of a fully-signed OpeningTx,
// unmarshals it and returns it.
func VerifyOpeningTx(ev *wire.Envelope, fn func([]byte) error) (*wire.OpeningTx, error) {
	otx := wire.OpeningTx{}
	err := proto.Unmarshal(ev.Payload, &otx)
	if err != nil {
		return nil, err
	}

	// Check signatures
	if !ed25519.Verify(sliceTo32Byte(otx.Pubkey1), ev.Payload, sliceTo64Byte(ev.Signature1)) {
		return nil, errors.New("signature 1 invalid")
	}
	if !ed25519.Verify(sliceTo32Byte(otx.Pubkey2), ev.Payload, sliceTo64Byte(ev.Signature2)) {
		return nil, errors.New("signature 2 invalid")
	}

	// Check state
	err = fn(otx.State)
	if err != nil {
		return nil, err
	}

	return &otx, nil
}

// VerifyUpdateTx checks the signatures and the state of a fully-signed UpdateTx,
// unmarshals it and returns it.
func (ch *Channel) VerifyUpdateTx(ev *wire.Envelope, fn func([]byte, []byte) error) (*wire.UpdateTx, error) {
	// Check signatures
	if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey1), ev.Payload, sliceTo64Byte(ev.Signature1)) {
		return nil, errors.New("signature 1 invalid")
	}
	if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey2), ev.Payload, sliceTo64Byte(ev.Signature2)) {
		return nil, errors.New("signature 2 invalid")
	}

	utx := wire.UpdateTx{}
	err := proto.Unmarshal(ev.Payload, &utx)
	if err != nil {
		return nil, err
	}

	// Check state
	err = fn(ch.OpeningTx.State, utx.State)
	if err != nil {
		return nil, err
	}

	return &utx, nil
}

func (ch *Channel) StartHoldPeriod(utx *wire.UpdateTx) error {
	if ch.Phase != schema.Channel_PendingClosed {
		if ch.LastFullUpdateTx.SequenceNumber > utx.SequenceNumber {
			return errors.New("update tx with higher sequence number exists")
		}
	}

	ch.Phase = schema.Channel_PendingClosed
	ch.LastFullUpdateTx = utx
	return nil
}

// AddFulfillment verifies a fulfillment's signature and adds it to the Channel's
// Fulfillments array.
func (ch *Channel) AddFulfillment(ev *wire.Envelope) error {
	if ch.Phase != schema.Channel_PendingClosed {
		return errors.New("channel must be pending closed")
	}

	if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey1), ev.Payload, sliceTo64Byte(ev.Signature1)) ||
		!ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkey2), ev.Payload, sliceTo64Byte(ev.Signature1)) {
		return errors.New("signature invalid")
	}

	ful := wire.Fulfillment{}
	err := proto.Unmarshal(ev.Payload, &ful)
	if err != nil {
		return err
	}

	ch.Fulfillments = append(ch.Fulfillments, ful.State)
	return nil
}

func (ch *Channel) VerifyFulfillments(fn func([]byte, []byte, [][]byte) error) error {
	return fn(ch.OpeningTx.State, ch.LastFullUpdateTx.State, ch.Fulfillments)
}

// CheckState checks the state of a channel, evaluating OpeningTx state, UpdateTx state,
// and the state of any Fulfillments submitted

// // StartClose changes the Channel to pending closed and signs the LastFullUpdateTx
// func (ch *Channel) StartClose() (*wire.Envelope, error) {
// 	if ch.Phase != schema.Channel_Open {
// 		return nil, errors.New("channel must be open")
// 	}
// 	ch.Phase = schema.Channel_PendingClosed
// 	return ch.LastFullUpdateTxEnvelope, nil
// }

// // ConfirmClose is called when we receive word from the bank that the channel is permanently closed
// func (ch *Channel) ConfirmClose(utx *wire.UpdateTx) error {
// 	if ch.Phase != schema.Channel_PendingClosed {
// 		return errors.New("channel must be pending closed")
// 	}
// 	ch.LastUpdateTx = utx
// 	ch.LastFullUpdateTx = utx
// 	// Change channel state to closed
// 	ch.Phase = schema.Channel_Closed
// 	return nil
// }

// func (ch *Channel) StartClose(utx *wire.UpdateTx) error {
// 	if ch.State != schema.Channel_PendingClosed {
// 		if ch.LastFullUpdateTx.SequenceNumber > utx.SequenceNumber {
// 			return errors.New("update tx with higher sequence number exists")
// 		}
// 	}

// 	ch.State = schema.Channel_PendingClosed
// 	ch.LastFullUpdateTx = utx
// 	return nil
// }
