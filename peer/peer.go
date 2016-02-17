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

	Me           uint32
	Fulfillments [][]byte

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

// NewAccount makes a new my account
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
func (acct *Account) SignOpeningTx(otx *wire.OpeningTx) (*wire.Envelope, error) {
	// Serialize opening transaction
	data, err := proto.Marshal(otx)
	if err != nil {
		return nil, err
	}

	// Make new envelope
	return &wire.Envelope{
		Payload: data,
		Signatures: [][]byte{
			ed25519.Sign(sliceTo64Byte(acct.Privkey), data)[:],
		},
	}, nil
}

func UnpackageOpeningTx(ev *wire.Envelope) (*wire.OpeningTx, error) {
	otx := &wire.OpeningTx{}
	err := proto.Unmarshal(ev.Payload, otx)
	if err != nil {
		return nil, err
	}
	return otx, nil
}

func (acct *Account) SignEnvelope(ev *wire.Envelope) *wire.Envelope {
	ev.Signatures = append(ev.Signatures, [][]byte{ed25519.Sign(sliceTo64Byte(acct.Privkey), ev.Payload)[:]}...)

	return ev
}

func (ch *Channel) CheckSignatures(ev *wire.Envelope) int {
	res := 0
	if ed25519.Verify(sliceTo32Byte(ch.Account.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
		res++
	}
	if ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[1])) {
		res++
	}
	if ed25519.Verify(sliceTo32Byte(ch.Judge.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[2])) {
		res++
	}
	return res
}

// NewChannel creates a new Channel from an Envelope containing an opening transaction,
// an Account and a Counterparty.
func NewChannel(ev *wire.Envelope, acct *Account, cpt *Counterparty) (*Channel, error) {
	otx := &wire.OpeningTx{}
	err := proto.Unmarshal(ev.Payload, otx)
	if err != nil {
		return nil, err
	}

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

func (ch *Channel) Open(ev *wire.Envelope, otx *wire.OpeningTx) error {
	if ed25519.Verify(sliceTo32Byte(ch.Account.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
		return errors.New("my account signature invalid")
	}
	if ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[1])) {
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
func (ch *Channel) SignUpdateTx(utx *wire.UpdateTx) (*wire.Envelope, error) {
	// Serialize update transaction
	data, err := proto.Marshal(utx)
	if err != nil {
		return nil, err
	}
	// Make new envelope
	ev := wire.Envelope{
		Payload:    data,
		Signatures: [][]byte{ed25519.Sign(sliceTo64Byte(ch.Account.Privkey), data)[:]},
	}

	return &ev, nil
}

func (ch *Channel) CheckUpdateTx(ev *wire.Envelope, utx *wire.UpdateTx) error {
	switch ch.Me {
	case 0:
		if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[1])) {
			return errors.New("signature not valid")
		}
	case 1:
		if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
			return errors.New("signature not valid")
		}
	}

	// Check ChannelId
	if utx.ChannelId != ch.OpeningTx.ChannelId {
		return errors.New("channel id incorrect")
	}

	lst := ch.ProposedUpdateTx

	if lst != nil {
		if lst.SequenceNumber < utx.SequenceNumber {
			return errors.New("sequence number not valid")
		}
	}

	return nil
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

// AddFulfillment verifies a fulfillment's signature and adds it to the Channel's
// Fulfillments array.
func (ch *Channel) AddFulfillment(ev *wire.Envelope) error {
	if ch.Phase != PENDING_CLOSED {
		return errors.New("channel must be pending closed")
	}

	if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkeys[0]), ev.Payload, sliceTo64Byte(ev.Signatures[0])) ||
		!ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkeys[1]), ev.Payload, sliceTo64Byte(ev.Signatures[1])) {
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
