package peer

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"

	"github.com/agl/ed25519"
	"github.com/golang/protobuf/proto"
	"github.com/jtremback/usc-core/wire"
)

// Phases of a Tx
// Created
// Confirmed
// Verified

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

type Phase int

const (
	PENDING_OPEN   Phase = 1
	OPEN           Phase = 2
	PENDING_CLOSED Phase = 3
	CLOSED         Phase = 4
)

type Channel struct {
	ChannelId string
	Phase     Phase

	OpeningTx         *wire.OpeningTx
	OpeningTxEnvelope *wire.Envelope

	ProposedUpdateTx         *wire.UpdateTx
	ProposedUpdateTxEnvelope *wire.Envelope

	LastFullUpdateTx         *wire.UpdateTx
	LastFullUpdateTxEnvelope *wire.Envelope

	Me          uint32
	FollowOnTxs []*wire.Envelope

	Judge        *Judge
	Account      *Account
	Counterparty *Counterparty
}

type Account struct {
	Name    string
	Pubkey  []byte
	Privkey []byte
	Judge   *Judge
}

type Counterparty struct {
	Name    string
	Pubkey  []byte
	Address string
	Judge   *Judge
}

type Judge struct {
	Name    string
	Pubkey  []byte
	Address string
}

// NewAccount makes a new account
func NewAccount(name string, address string, ep *Judge) (*Account, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Account{
		Name:    name,
		Judge:   ep,
		Pubkey:  pub[:],
		Privkey: priv[:],
	}, nil
}

// NewOpeningTx assembles an OpeningTx
func (acct *Account) NewOpeningTx(cpt *Counterparty, state []byte, holdPeriod uint32) (*wire.OpeningTx, error) {
	b, err := randomBytes(32)
	chID := string(b)
	if err != nil {
		return nil, err
	}

	pubkeys := [][]byte{acct.Pubkey, cpt.Pubkey}

	return &wire.OpeningTx{
		ChannelId:  chID,
		Pubkeys:    pubkeys,
		State:      state,
		HoldPeriod: holdPeriod,
	}, nil
}

// SignOpeningTx signs and serializes an opening transaction
func (acct *Account) SerializeOpeningTx(otx *wire.OpeningTx) (*wire.Envelope, error) {
	// Serialize opening transaction
	data, err := proto.Marshal(otx)
	if err != nil {
		return nil, err
	}

	// Make new envelope
	return &wire.Envelope{
		Payload: data,
	}, nil
}

func (acct *Account) CheckOpeningTx(ev *wire.Envelope, cpt *Counterparty) error {
	if bytes.Compare(acct.Judge.Pubkey, cpt.Judge.Pubkey) != 0 {
		return errors.New("accounts do not have matching judges")
	}
	if ed25519.Verify(sliceTo32Byte(cpt.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[1])) {
		return errors.New("counterparty signature invalid")
	}

	return nil
}

func (acct *Account) SignEnvelope(ev *wire.Envelope) {
	ev.Signatures = append(ev.Signatures, [][]byte{ed25519.Sign(sliceTo64Byte(acct.Privkey), ev.Payload)[:]}...)
}

func NewChannel(ev *wire.Envelope, otx *wire.OpeningTx, acct *Account, cpt *Counterparty) (*Channel, error) {
	// Who is Me?
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
		Account:           acct,
		Counterparty:      cpt,
		Phase:             PENDING_OPEN,
	}

	return ch, nil
}

func (ch *Channel) Confirm(ev *wire.Envelope, otx *wire.OpeningTx) error {
	var them int
	switch ch.Me {
	case 0:
		them = 1
	case 1:
		them = 0
	}

	if ed25519.Verify(sliceTo32Byte(ch.Account.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[ch.Me])) {
		return errors.New("my account signature invalid")
	}
	if ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[them])) {
		return errors.New("their account signature invalid")
	}
	if ed25519.Verify(sliceTo32Byte(ch.Judge.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[2])) {
		return errors.New("judge signature invalid")
	}

	if bytes.Compare(ev.Payload, ch.OpeningTxEnvelope.Payload) != 0 {
		return errors.New("opening tx invalid")
	}

	ch.Phase = OPEN

	return nil
}

// NewUpdateTx makes a new UpdateTx on Channel with NetTransfer changed by amount.
func (ch *Channel) NewUpdateTx(state []byte, fast bool) (*wire.UpdateTx, error) {
	lst := ch.ProposedUpdateTx
	var seq uint32
	if lst != nil {
		seq = lst.SequenceNumber + 1
	}

	// Make new update transaction
	return &wire.UpdateTx{
		ChannelId:      ch.ChannelId,
		State:          state,
		SequenceNumber: seq,
		Fast:           fast,
	}, nil
}

// SignUpdateTx signs an update proposal and puts it in an envelope
func (ch *Channel) SerializeUpdateTx(utx *wire.UpdateTx) (*wire.Envelope, error) {
	// Serialize update transaction
	data, err := proto.Marshal(utx)
	if err != nil {
		return nil, err
	}
	// Make new envelope
	ev := wire.Envelope{
		Payload: data,
	}

	return &ev, nil
}

func (ch *Channel) CheckUpdateTx(ev *wire.Envelope, utx *wire.UpdateTx) (uint32, error) {
	switch ch.Me {
	case 0:
		if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[1])) {
			return 0, errors.New("signature not valid")
		}
	case 1:
		if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
			return 0, errors.New("signature not valid")
		}
	}

	if utx.ChannelId != ch.OpeningTx.ChannelId {
		return 0, errors.New("channel id incorrect")
	}

	if ch.ProposedUpdateTx != nil {
		if !(ch.ProposedUpdateTx.SequenceNumber < utx.SequenceNumber) {
			return ch.ProposedUpdateTx.SequenceNumber, errors.New("sequence number too low")
		}
	}

	if ch.LastFullUpdateTx != nil {
		if !(ch.LastFullUpdateTx.SequenceNumber < utx.SequenceNumber) {
			return ch.LastFullUpdateTx.SequenceNumber, errors.New("sequence number too low")
		}
	}

	ch.ProposedUpdateTx = utx
	ch.ProposedUpdateTxEnvelope = ev

	return 0, nil
}

// ConfirmUpdateTx confirms a channel's proposed update tx, if it has one
func (ch *Channel) ConfirmUpdateTx() (*wire.Envelope, error) {
	if ch.Phase != OPEN {
		return nil, errors.New("channel must be open")
	}

	ev := &wire.Envelope{}

	if ch.ProposedUpdateTxEnvelope == ev {
		return nil, errors.New("no proposed update tx")
	}

	if bytes.Compare(ch.ProposedUpdateTxEnvelope.Signatures[ch.Me], []byte{}) == 0 {
		return nil, errors.New("proposed update tx belongs to me")
	}

	utx := ch.ProposedUpdateTx
	ev = ch.ProposedUpdateTxEnvelope

	ev.Signatures = append(
		ev.Signatures,
		[][]byte{ed25519.Sign(sliceTo64Byte(ch.Account.Privkey), ev.Payload)[:]}...,
	)

	ch.LastFullUpdateTx = utx
	ch.LastFullUpdateTxEnvelope = ev

	return ev, nil
}

func (ch *Channel) CheckFinalUpdateTx(ev *wire.Envelope, utx *wire.UpdateTx) (*wire.Envelope, error) {
	if ed25519.Verify(sliceTo32Byte(ch.Account.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
		return nil, errors.New("my account signature invalid")
	}
	if ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[1])) {
		return nil, errors.New("their account signature invalid")
	}
	if ed25519.Verify(sliceTo32Byte(ch.Judge.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[2])) {
		return nil, errors.New("judge signature invalid")
	}

	if ch.Phase == PENDING_CLOSED {
		if ch.LastFullUpdateTx.SequenceNumber > utx.SequenceNumber {
			return ch.LastFullUpdateTxEnvelope, nil
		}
	}
	ch.Phase = PENDING_CLOSED
	ch.LastFullUpdateTx = utx
	ch.LastFullUpdateTxEnvelope = ev

	return nil, nil
}

// AddFollowOnTx verifies a FollowOnTx's signature and adds it to the Channel's
// FollowOnTxs array.
func (ch *Channel) AddFollowOnTx(ev *wire.Envelope) error {
	if ch.Phase != PENDING_CLOSED {
		return errors.New("channel must be pending closed")
	}

	if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
		return errors.New("signature invalid")
	}

	ch.FollowOnTxs = append(ch.FollowOnTxs, ev)
	return nil
}
