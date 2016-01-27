package wallet

import (
	"fmt"
	"github.com/agl/ed25519"
	"github.com/jtremback/upc/wire"
	"reflect"
	"testing"
)

var ident = &Identity{
	Nickname: "alfred",
	Pubkey:   []byte{71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	Privkey:  []byte{147, 131, 100, 59, 112, 77, 196, 211, 124, 170, 199, 79, 190, 194, 175, 244, 1, 9, 48, 255, 200, 168, 138, 165, 187, 46, 251, 28, 183, 13, 214, 5, 71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	Channels: [][]byte{[]byte{4, 20}},
}

var ident2 = &Identity{
	Nickname: "billary",
	Pubkey:   []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Privkey:  []byte{184, 174, 56, 197, 104, 10, 100, 13, 194, 229, 111, 227, 49, 49, 126, 232, 117, 100, 207, 170, 154, 36, 118, 153, 143, 150, 182, 228, 98, 161, 144, 112, 166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Channels: [][]byte{[]byte{4, 20}},
}

var ch = &Channel{
	ChannelID: []byte{4, 20},
	OpeningTx: &wire.OpeningTx{
		Pubkey1: ident.Pubkey,
		Pubkey2: ident2.Pubkey,
		Amount1: 100,
		Amount2: 100,
	},
	LastUpdateTx: &wire.UpdateTx{
		ChannelID:      []byte{4, 20},
		NetTransfer:    -24,
		SequenceNumber: 1,
	},
	Me: 1,
}

var ch2 = &Channel{
	ChannelID: []byte{4, 20},
	OpeningTx: &wire.OpeningTx{
		Pubkey1: ident.Pubkey,
		Pubkey2: ident2.Pubkey,
		Amount1: 100,
		Amount2: 100,
	},
	LastUpdateTx: &wire.UpdateTx{
		ChannelID:      []byte{4, 20},
		NetTransfer:    -24,
		SequenceNumber: 1,
	},
	Me: 2,
}

// func TestNewChannel(t *testing.T) {
// 	NewChannel(ident, ident2, 100, 100, holdPeriod)
// }

func TestNewUpdateTxProposal(t *testing.T) {
	ideal := &wire.UpdateTx{
		ChannelID:      []byte{4, 20},
		NetTransfer:    -12,
		SequenceNumber: 2,
		Fast:           false,
	}
	actual, err := ch.NewUpdateTxProposal(12)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(ideal, actual) {
		t.Error("incorrect output")
	}
}

func TestSignUpdateTxProposal(t *testing.T) {
	utx, err := ch.NewUpdateTxProposal(12)
	if err != nil {
		t.Error(err)
	}

	ev, err := ch.SignUpdateTxProposal(utx)
	if err != nil {
		t.Error(err)
	}

	if !ed25519.Verify(sliceTo32Byte(ident.Pubkey), ev.Payload, sliceTo64Byte(ev.Signature1)) {
		t.Error("invalid signature")
	}
}

func TestVerifyUpdateTxProposal(t *testing.T) {
	utx, err := ch2.NewUpdateTxProposal(12)
	if err != nil {
		t.Error(err)
	}

	ev, err := ch2.SignUpdateTxProposal(utx)
	if err != nil {
		t.Error(err)
	}

	amt, err := ch.VerifyUpdateTxProposal(ev)
	if err != nil {
		t.Error(err)
	}

	if amt != 12 {
		t.Error("wrong amt:", amt)
	}

	fmt.Println(amt, err)
}
