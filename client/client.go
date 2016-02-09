package client

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

	LastUpdateTx         *wire.UpdateTx
	LastUpdateTxEnvelope *wire.Envelope

	LastFullUpdateTx         *wire.UpdateTx
	LastFullUpdateTxEnvelope *wire.Envelope

	Me           uint32
	Fulfillments [][]byte

	Judge        *Judge
	MyAccount    *MyAccount
	TheirAccount *TheirAccount
}

type MyAccount struct {
	Name    string
	Pubkey  []byte
	Privkey []byte
	Judge   *Judge
}

type TheirAccount struct {
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
func NewAccount(name string, address string, ep *Judge) (*MyAccount, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &MyAccount{
		Name:    name,
		Judge:   ep,
		Pubkey:  pub[:],
		Privkey: priv[:],
	}, nil
}

// NewOpeningTx assembles an OpeningTx
func (acct *MyAccount) NewOpeningTx(tAcct *TheirAccount, state []byte, holdPeriod uint32) (*wire.OpeningTx, error) {
	b, err := randomBytes(32)
	chID := string(b)
	if err != nil {
		return nil, err
	}

	pubkeys := [][]byte{acct.Pubkey, tAcct.Pubkey}

	return &wire.OpeningTx{
		ChannelId:  chID,
		Pubkeys:    pubkeys,
		State:      state,
		HoldPeriod: holdPeriod,
	}, nil
}

// SignOpeningTx signs and serializes an opening transaction
func (acct *MyAccount) SignOpeningTx(otx *wire.OpeningTx) (*wire.Envelope, error) {
	// Serialize opening transaction
	data, err := proto.Marshal(otx)
	if err != nil {
		return nil, err
	}

	// Make new envelope
	return &wire.Envelope{
		Type:    wire.Envelope_OpeningTx,
		Payload: data,
		Signatures: [][]byte{
			ed25519.Sign(sliceTo64Byte(acct.Privkey), data)[:],
		},
	}, nil
}

func (acct *MyAccount) ConfirmOpeningTx(ev *wire.Envelope) (*wire.Envelope, *wire.OpeningTx, error) {
	otx := &wire.OpeningTx{}
	err := proto.Unmarshal(ev.Payload, otx)
	if err != nil {
		return nil, nil, err
	}

	if !ed25519.Verify(sliceTo32Byte(otx.Pubkeys[0]), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
		return nil, nil, errors.New("signature 0 invalid")
	}

	ev.Signatures = append(ev.Signatures, [][]byte{ed25519.Sign(sliceTo64Byte(acct.Privkey), ev.Payload)[:]}...)

	return ev, otx, nil
}

// NewChannel creates a new Channel from an Envelope containing an opening transaction,
// an Account and a Peer.
func (acct *MyAccount) NewChannel(ev *wire.Envelope, mAcct *MyAccount, tAcct *TheirAccount) (*Channel, error) {
	if !ed25519.Verify(sliceTo32Byte(acct.Judge.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[len(ev.Signatures)-1])) {
		return nil, errors.New("judge signature invalid")
	}
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
		MyAccount:         mAcct,
		TheirAccount:      tAcct,
		Phase:             OPEN,
	}

	return ch, nil
}

// NewUpdateTx makes a new UpdateTx on Channel with NetTransfer changed by amount.
func (ch *Channel) NewUpdateTx(state []byte, fast bool) (*wire.UpdateTx, error) {
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

// SignUpdateTx signs an update proposal and puts it in an envelope
func (ch *Channel) SignUpdateTx(utx *wire.UpdateTx) (*wire.Envelope, error) {
	// Serialize update transaction
	data, err := proto.Marshal(utx)
	if err != nil {
		return nil, err
	}
	// Make new envelope
	ev := wire.Envelope{
		Type:       wire.Envelope_UpdateTx,
		Payload:    data,
		Signatures: [][]byte{ed25519.Sign(sliceTo64Byte(ch.MyAccount.Privkey), data)[:]},
	}

	return &ev, nil
}

// ConfirmUpdateTx takes an Envelope containing a UpdateTx with one
// signature and checks the signature, as well as the sequence number. It also
// calls fn to verify the state. If all factors are correct, it signs the UpdateTx
// and puts it in an envelope.
func (ch *Channel) ConfirmUpdateTx(ev *wire.Envelope) (*wire.Envelope, *wire.UpdateTx, error) {
	if ch.Phase != OPEN {
		return nil, nil, errors.New("channel must be open")
	}

	// Read signature from correct slot
	// Copy signature and pubkey
	switch ch.Me {
	case 0:
		if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkeys[1]), ev.Payload, sliceTo64Byte(ev.Signatures[1])) {
			return nil, nil, errors.New("invalid signature")
		}
	case 1:
		if !ed25519.Verify(sliceTo32Byte(ch.OpeningTx.Pubkeys[0]), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
			return nil, nil, errors.New("invalid signature")
		}
	}

	utx := &wire.UpdateTx{}
	err := proto.Unmarshal(ev.Payload, utx)
	if err != nil {
		return nil, nil, err
	}

	// Check ChannelId
	if utx.ChannelId != ch.OpeningTx.ChannelId {
		return nil, nil, errors.New("ChannelId does not match")
	}

	lst := ch.LastUpdateTx

	if lst != nil {
		// Check last sequence number
		if lst.SequenceNumber+1 != utx.SequenceNumber {
			return nil, nil, errors.New("invalid sequence number")
		}
	}

	// Sign envelope
	ev.Signatures = append(ev.Signatures, [][]byte{ed25519.Sign(sliceTo64Byte(ch.MyAccount.Privkey), ev.Payload)[:]}...)

	return ev, utx, nil
}

func (ch *Channel) StartHoldPeriod(utx *wire.UpdateTx) error {
	if ch.Phase == PENDING_CLOSED {
		if ch.LastFullUpdateTx.SequenceNumber > utx.SequenceNumber {
			return errors.New("update tx with higher sequence number exists")
		}
	}
	ch.Phase = PENDING_CLOSED
	ch.LastFullUpdateTx = utx
	return nil
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
