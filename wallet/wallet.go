package wallet

import (
	"crypto/rand"
	"errors"
	// "fmt"
	"github.com/agl/ed25519"
	"github.com/golang/protobuf/proto"
	"github.com/jtremback/upc-core/wallet/schema"
	"github.com/jtremback/upc-core/wire"
	"io"
)

// Channel
type Channel schema.Channel
type OpeningTx wire.OpeningTx

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

func calcConditionalTransfer(lst *wire.UpdateTx) int64 {
	// Sum up all conditional transfer amounts and add to net transfer
	var ct int64
	for _, v := range lst.Conditions {
		ct += v.ConditionalTransfer
	}

	return ct
}

// NewChannel makes a new Channel with the supplied information. `Pubkey1` of
// the created Channel will correspond to `account`.
func NewChannel(
	account *schema.Account,
	peer *schema.Peer,
	myAmount uint32,
	theirAmount uint32,
	holdPeriod uint32,
) (*Channel, error) {
	b, err := randomBytes(32)
	chID := string(b)
	if err != nil {
		return nil, err
	}

	ch := &Channel{
		ChannelId: chID,
		OpeningTx: &wire.OpeningTx{
			ChannelId:  chID,
			Pubkey1:    account.Pubkey,
			Pubkey2:    peer.Pubkey,
			Amount1:    myAmount,
			Amount2:    theirAmount,
			HoldPeriod: holdPeriod,
		},
		Me:      1,
		Account: account,
		Peer:    peer,
		State:   schema.Channel_PendingOpen,
	}

	// Serialize opening transaction
	data, err := proto.Marshal(ch.OpeningTx)
	if err != nil {
		return nil, err
	}

	// Make new envelope
	ch.OpeningTxEnvelope = &wire.Envelope{
		Type:       wire.Envelope_OpeningTx,
		Payload:    data,
		Signature1: ed25519.Sign(sliceTo64Byte(account.Privkey), data)[:],
	}

	return ch, nil
}

// VerifyOpeningTxProposal checks if a partially-signed OpeningTx has the correct
// signature
func VerifyOpeningTxProposal(ev *wire.Envelope) (*wire.OpeningTx, error) {
	otx := &wire.OpeningTx{}
	err := proto.Unmarshal(ev.Payload, otx)
	if err != nil {
		return nil, err
	}

	if !ed25519.Verify(sliceTo32Byte(otx.Pubkey1), ev.Payload, sliceTo64Byte(ev.Signature1)) {
		return nil, errors.New("signature invalid")
	}

	return otx, nil
}

// NewChannelFromOpeningTxProposal
func NewChannelFromOpeningTxProposal(otx *wire.OpeningTx, acct *schema.Account, peer *schema.Peer) (*Channel, error) {
	ch := &Channel{
		ChannelId: otx.ChannelId,
		OpeningTx: otx,
		Me:        2,
		Account:   acct,
		Peer:      peer,
		State:     schema.Channel_Open,
	}

	return ch, nil
}

// NewUpdateTxProposal makes a new UpdateTx on Channel with NetTransfer changed by amount.
func (ch *Channel) NewUpdateTxProposal(amount int64) (*wire.UpdateTx, error) {
	lst := ch.LastUpdateTx
	var nt int64
	var seq uint32
	if lst != nil {
		nt = lst.NetTransfer
		nt += calcConditionalTransfer(lst)
		seq = lst.SequenceNumber + 1
	}
	// Check if we are pubkey1 or pubkey2 and add or subtract amount from net transfer
	switch ch.Me {
	case 1:
		nt += amount
	case 2:
		nt -= amount
	}

	// Check if the net transfer amount is still valid
	if nt > int64(ch.OpeningTx.Amount1) || nt < -int64(ch.OpeningTx.Amount2) {
		return nil, errors.New("invalid amount")
	}

	// Make new update transaction
	return &wire.UpdateTx{
		ChannelId:      ch.ChannelId,
		NetTransfer:    nt,
		SequenceNumber: seq,
		Fast:           false,
	}, nil
}

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

// VerifyUpdateTxProposal takes an Envelope containing a UpdateTx with one
// signature and checks the signature, as well as the sequence number and
// returns the amount that it add or subtract from our account.
func (ch *Channel) VerifyUpdateTxProposal(ev *wire.Envelope) (uint32, error) {
	if ch.State != schema.Channel_Open {
		return 0, errors.New("channel must be open")
	}

	var pubkey [32]byte
	var sig [64]byte

	// Read signature from correct slot
	// Copy signature and pubkey
	switch ch.Me {
	case 1:
		pubkey = *sliceTo32Byte(ch.OpeningTx.Pubkey2)
		sig = *sliceTo64Byte(ev.Signature2)
	case 2:
		pubkey = *sliceTo32Byte(ch.OpeningTx.Pubkey1)
		sig = *sliceTo64Byte(ev.Signature1)
	}

	// Check signature
	if !ed25519.Verify(&pubkey, ev.Payload, &sig) {
		return 0, errors.New("invalid signature")
	}

	utx := wire.UpdateTx{}
	err := proto.Unmarshal(ev.Payload, &utx)
	if err != nil {
		return 0, err
	}

	// Check ChannelId
	if utx.ChannelId != ch.OpeningTx.ChannelId {
		return 0, errors.New("")
	}

	lst := ch.LastUpdateTx

	// Check last sequence number
	if lst.SequenceNumber+1 != utx.SequenceNumber {
		return 0, errors.New("invalid sequence number")
	}

	// Get amount depending on if we are party 1 or party 2
	var amt uint32
	switch ch.Me {
	case 1:
		amt = uint32(lst.NetTransfer - utx.NetTransfer)
	case 2:
		amt = uint32(utx.NetTransfer - lst.NetTransfer)
	}

	return amt, nil
}

// StartClose changes the Channel to pending closed and signs the LastFullUpdateTx
func (ch *Channel) StartClose() (*wire.Envelope, error) {
	if ch.State != schema.Channel_Open {
		return nil, errors.New("channel must be open")
	}
	ch.State = schema.Channel_PendingClosed
	return ch.LastFullUpdateTxEnvelope, nil
}

// ConfirmClose is called when we receive word from the bank that the channel is permanently closed
func (ch *Channel) ConfirmClose(utx *wire.UpdateTx) error {
	if ch.State != schema.Channel_PendingClosed {
		return errors.New("channel must be pending closed")
	}
	ch.LastUpdateTx = utx
	ch.LastFullUpdateTx = utx
	// Change channel state to closed
	ch.State = schema.Channel_Closed
	return nil
}
