package peer

import (
	"bytes"
	"crypto/rand"
	"errors"
	"io"

	"github.com/agl/ed25519"
	"github.com/golang/protobuf/proto"
	"github.com/jtremback/usc/core/wire"
)

// Phases of a Tx
// Created
// Confirmed
// Verified

func sliceTo64Byte(slice []byte) *[64]byte {
	if len(slice) == 64 {
		var array [64]byte
		copy(array[:], slice[:64])
		return &array
	}
	return &[64]byte{}
}

func sliceTo32Byte(slice []byte) *[32]byte {
	if len(slice) == 32 {
		var array [32]byte
		copy(array[:], slice[:32])
		return &array
	}
	return &[32]byte{}
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

var swap = []int{1, 0}

type Channel struct {
	ChannelId string
	Phase     Phase

	OpeningTx         *wire.OpeningTx
	OpeningTxEnvelope *wire.Envelope

	LastFullUpdateTx         *wire.UpdateTx
	LastFullUpdateTxEnvelope *wire.Envelope

	MyProposedUpdateTx         *wire.UpdateTx
	MyProposedUpdateTxEnvelope *wire.Envelope

	TheirProposedUpdateTx         *wire.UpdateTx
	TheirProposedUpdateTxEnvelope *wire.Envelope

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

func NewAccount(name string, jd *Judge) (*Account, error) {
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}

	return &Account{
		Name:    name,
		Judge:   jd,
		Pubkey:  pub[:],
		Privkey: priv[:],
	}, nil
}

func (acct *Account) NewOpeningTx(channelId string, cpt *Counterparty, state []byte, holdPeriod uint64) (*wire.OpeningTx, error) {
	pubkeys := [][]byte{acct.Pubkey, cpt.Pubkey}

	return &wire.OpeningTx{
		ChannelId:  channelId,
		Pubkeys:    pubkeys,
		State:      state,
		HoldPeriod: holdPeriod,
	}, nil
}

func SerializeOpeningTx(otx *wire.OpeningTx) (*wire.Envelope, error) {
	data, err := proto.Marshal(otx)
	if err != nil {
		return nil, err
	}
	return &wire.Envelope{
		Payload: data,
	}, nil
}

func (acct *Account) CheckOpeningTx(ev *wire.Envelope, cpt *Counterparty) error {
	if len(ev.Signatures) != 1 {
		return errors.New("wrong number of signatures")
	}
	if bytes.Compare(acct.Judge.Pubkey, cpt.Judge.Pubkey) != 0 {
		return errors.New("accounts do not have the same judge")
	}
	if !ed25519.Verify(sliceTo32Byte(cpt.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
		return errors.New("counterparty signature not valid")
	}

	return nil
}

func (acct *Account) AppendSignature(ev *wire.Envelope) {
	ev.Signatures = append(ev.Signatures, [][]byte{ed25519.Sign(sliceTo64Byte(acct.Privkey), ev.Payload)[:]}...)
}

func NewChannel(ev *wire.Envelope, otx *wire.OpeningTx, acct *Account, cpt *Counterparty) (*Channel, error) {
	if bytes.Compare(acct.Judge.Pubkey, cpt.Judge.Pubkey) != 0 {
		return nil, errors.New("accounts do not have the same judge")
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
		Judge:             acct.Judge,
		Counterparty:      cpt,
		Phase:             PENDING_OPEN,
	}

	return ch, nil
}

func (ch *Channel) Open(ev *wire.Envelope, otx *wire.OpeningTx) error {
	if ch.Phase != PENDING_OPEN {
		return errors.New("channel not PENDING_OPEN")
	}
	if len(ev.Signatures) != 3 {
		return errors.New("wrong number of signatures")
	}
	if !ed25519.Verify(sliceTo32Byte(ch.Account.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[ch.Me])) {
		return errors.New("my account signature not valid")
	}
	if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[swap[ch.Me]])) {
		return errors.New("counterparty signature not valid")
	}
	if !ed25519.Verify(sliceTo32Byte(ch.Judge.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[2])) {
		return errors.New("judge signature not valid")
	}

	if bytes.Compare(ev.Payload, ch.OpeningTxEnvelope.Payload) != 0 {
		return errors.New("opening tx not valid")
	}

	ch.Phase = OPEN
	ch.OpeningTx = otx
	ch.OpeningTxEnvelope = ev

	return nil
}

func (ch *Channel) HighestSeq() uint32 {
	var num uint32
	if ch.MyProposedUpdateTx != nil {
		if ch.MyProposedUpdateTx.SequenceNumber > num {
			num = ch.MyProposedUpdateTx.SequenceNumber
		}
	}

	if ch.TheirProposedUpdateTx != nil {
		if ch.TheirProposedUpdateTx.SequenceNumber > num {
			num = ch.TheirProposedUpdateTx.SequenceNumber
		}
	}

	return num
}

func (ch *Channel) NewUpdateTx(state []byte, fast bool) *wire.UpdateTx {
	return &wire.UpdateTx{
		ChannelId:      ch.ChannelId,
		State:          state,
		SequenceNumber: ch.HighestSeq() + 1,
		Fast:           fast,
	}
}

func SerializeUpdateTx(utx *wire.UpdateTx) (*wire.Envelope, error) {
	data, err := proto.Marshal(utx)
	if err != nil {
		return nil, err
	}

	ev := wire.Envelope{
		Payload:    data,
		Signatures: [][]byte{[]byte{}, []byte{}},
	}

	return &ev, nil
}

func (ch *Channel) SignProposedUpdateTx(ev *wire.Envelope, utx *wire.UpdateTx) {
	ev.Signatures[ch.Me] = ed25519.Sign(sliceTo64Byte(ch.Account.Privkey), ev.Payload)[:]
	ch.MyProposedUpdateTx = utx
	ch.MyProposedUpdateTxEnvelope = ev
}

func (ch *Channel) CosignProposedUpdateTx() *wire.Envelope {
	ev := ch.TheirProposedUpdateTxEnvelope
	ev.Signatures[ch.Me] = ed25519.Sign(sliceTo64Byte(ch.Account.Privkey), ev.Payload)[:]

	ch.LastFullUpdateTx = ch.TheirProposedUpdateTx
	ch.LastFullUpdateTxEnvelope = ch.TheirProposedUpdateTxEnvelope

	return ev
}

func (ch *Channel) AddProposedUpdateTx(ev *wire.Envelope, utx *wire.UpdateTx) error {
	if !(ch.Phase == OPEN || ch.Phase == PENDING_CLOSED) {
		return errors.New("channel not OPEN or PENDING_CLOSED")
	}
	if len(ev.Signatures) != 2 {
		return errors.New("wrong number of signatures")
	}
	if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[swap[ch.Me]])) {
		return errors.New("counterparty signature not valid")
	}
	if utx.ChannelId != ch.OpeningTx.ChannelId {
		return errors.New("channel id incorrect")
	}
	if !(utx.SequenceNumber > ch.HighestSeq()) {
		return errors.New("sequence number too low")
	}

	ch.TheirProposedUpdateTx = utx
	ch.TheirProposedUpdateTxEnvelope = ev

	return nil
}

func (ch *Channel) AddFullUpdateTx(ev *wire.Envelope, utx *wire.UpdateTx) error {
	if !(ch.Phase == OPEN || ch.Phase == PENDING_CLOSED) {
		return errors.New("channel not OPEN or PENDING_CLOSED")
	}
	if len(ev.Signatures) != 2 {
		return errors.New("wrong number of signatures")
	}
	if !ed25519.Verify(sliceTo32Byte(ch.Account.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[ch.Me])) {
		return errors.New("my account signature not valid")
	}
	if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[swap[ch.Me]])) {
		return errors.New("counterparty signature not valid")
	}
	if utx.ChannelId != ch.OpeningTx.ChannelId {
		return errors.New("channel id incorrect")
	}
	if ch.LastFullUpdateTx != nil {
		if utx.SequenceNumber > ch.LastFullUpdateTx.SequenceNumber {
			return errors.New("sequence number too low")
		}
	}

	ch.LastFullUpdateTx = utx
	ch.LastFullUpdateTxEnvelope = ev

	return nil
}

// func (ch *Channel) AddFullUpdateTx(ev *wire.Envelope, utx *wire.UpdateTx) (*wire.Envelope, error) {
// 	if !(ch.Phase == OPEN || ch.Phase == PENDING_CLOSED) {
// 		return nil, errors.New("channel not OPEN or PENDING_CLOSED")
// 	}
// 	if len(ev.Signatures) != 3 {
// 		return nil, errors.New("wrong number of signatures")
// 	}
// 	if !ed25519.Verify(sliceTo32Byte(ch.Account.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[ch.Me])) {
// 		return nil, errors.New("my account signature not valid")
// 	}
// 	if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[swap[ch.Me]])) {
// 		return nil, errors.New("counterparty signature not valid")
// 	}
// 	if !ed25519.Verify(sliceTo32Byte(ch.Judge.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[2])) {
// 		return nil, errors.New("judge signature not valid")
// 	}

// 	ch.Phase = PENDING_CLOSED
// 	ch.LastFullUpdateTx = utx
// 	ch.LastFullUpdateTxEnvelope = ev

// 	if ch.LastFullUpdateTx != nil {
// 		if ch.LastFullUpdateTx.SequenceNumber > utx.SequenceNumber {
// 			return ch.LastFullUpdateTxEnvelope, nil
// 		}
// 	}

// 	return nil, nil
// }

func (ch *Channel) NewClosingTx() *wire.ClosingTx {
	return &wire.ClosingTx{
		ChannelId: ch.ChannelId,
	}
}

func SerializeClosingTx(ctx *wire.ClosingTx) (*wire.Envelope, error) {
	data, err := proto.Marshal(ctx)
	if err != nil {
		return nil, err
	}

	return &wire.Envelope{
		Payload: data,
	}, nil
}

func (ch *Channel) NewFollowOnTx(state []byte) *wire.FollowOnTx {
	return &wire.FollowOnTx{
		ChannelId: ch.ChannelId,
		State:     state,
	}
}

func SerializeFollowOnTx(ftx *wire.FollowOnTx) (*wire.Envelope, error) {
	data, err := proto.Marshal(ftx)
	if err != nil {
		return nil, err
	}

	return &wire.Envelope{
		Payload: data,
	}, nil
}

func (ch *Channel) AddFollowOnTx(ev *wire.Envelope) error {
	if !(ch.Phase == OPEN || ch.Phase == PENDING_CLOSED) {
		return errors.New("channel not OPEN or PENDING_CLOSED")
	}
	if len(ev.Signatures) != 1 {
		return errors.New("wrong number of signatures")
	}
	if !ed25519.Verify(sliceTo32Byte(ch.Counterparty.Pubkey), ev.Payload, sliceTo64Byte(ev.Signatures[0])) {
		return errors.New("signature not valid")
	}

	ch.FollowOnTxs = append(ch.FollowOnTxs, ev)
	return nil
}
