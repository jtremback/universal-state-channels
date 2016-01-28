package wallet

import (
	"fmt"
	// "github.com/agl/ed25519"
	"github.com/jtremback/upc-core/wallet/schema"
	// "github.com/jtremback/upc-core/wire"
	// "reflect"
	"testing"
)

var account1 = &schema.Account{
	Name:    "alfred",
	Pubkey:  []byte{71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	Privkey: []byte{147, 131, 100, 59, 112, 77, 196, 211, 124, 170, 199, 79, 190, 194, 175, 244, 1, 9, 48, 255, 200, 168, 138, 165, 187, 46, 251, 28, 183, 13, 214, 5, 71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
}
var peer1 = &schema.Peer{
	Name:   "alfred",
	Pubkey: []byte{71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
}

var account2 = &schema.Account{
	Name:    "billary",
	Pubkey:  []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Privkey: []byte{184, 174, 56, 197, 104, 10, 100, 13, 194, 229, 111, 227, 49, 49, 126, 232, 117, 100, 207, 170, 154, 36, 118, 153, 143, 150, 182, 228, 98, 161, 144, 112, 166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
}
var peer2 = &schema.Peer{
	Name:   "billary",
	Pubkey: []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
}

// var ch = &Channel{
// 	ChannelId: "snoop",
// 	OpeningTx: &wire.OpeningTx{
// 		ChannelId: "snoop",
// 		Pubkey1:   account.Pubkey,
// 		Pubkey2:   account2.Pubkey,
// 		Amount1:   100,
// 		Amount2:   100,
// 	},
// 	LastUpdateTx: &wire.UpdateTx{
// 		ChannelId:      "poop",
// 		NetTransfer:    -24,
// 		SequenceNumber: 1,
// 	},
// 	Me: 1,
// }

// var ch2 = &Channel{
// 	ChannelId: "snoop",
// 	OpeningTx: &wire.OpeningTx{
// 		Pubkey1: account.Pubkey,
// 		Pubkey2: account2.Pubkey,
// 		Amount1: 100,
// 		Amount2: 100,
// 	},
// 	LastUpdateTx: &wire.UpdateTx{
// 		ChannelId:      "snoop",
// 		NetTransfer:    -24,
// 		SequenceNumber: 1,
// 	},
// 	Me: 2,
// }

// func TestNewChannel(t *testing.T) {
// 	NewChannel(account, account2, 100, 100, holdPeriod)
// }

// func TestNewChannel(t *testing.T) {
// 	ch, err := NewChannel(account1, peer2, 100, 100, 86400)
// 	if err != nil {
// 		t.Error(err)
// 	}
// 	fmt.Println(ch)
// }

// func TestNewUpdateTxProposal(t *testing.T) {
// 	ideal := &wire.UpdateTx{
// 		ChannelId:      "snoop",
// 		NetTransfer:    -12,
// 		SequenceNumber: 2,
// 		Fast:           false,
// 	}
// 	actual, err := ch.NewUpdateTxProposal(12)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if !reflect.DeepEqual(ideal, actual) {
// 		t.Error("incorrect output", ideal, actual)
// 	}
// }

// func TestNewUpdateTxProposal(t *testing.T) {
// 	ideal := &wire.UpdateTx{
// 		ChannelId:      "snoop",
// 		NetTransfer:    -12,
// 		SequenceNumber: 2,
// 		Fast:           false,
// 	}
// 	actual, err := ch.NewUpdateTxProposal(12)
// 	if err != nil {
// 		t.Error(err)
// 	}

// 	if !reflect.DeepEqual(ideal, actual) {
// 		t.Error("incorrect output", ideal, actual)
// 	}
// }

func TestVerifyUpdateTxProposal(t *testing.T) {
	ch1, err := NewChannel(account1, peer2, 100, 100, 86400)
	if err != nil {
		t.Error(err)
	}

	ch2, err := NewChannel(account2, peer1, 100, 100, 86400)
	if err != nil {
		t.Error(err)
	}

	utx, err := ch2.NewUpdateTxProposal(12)
	if err != nil {
		t.Error(err)
	}

	ev, err := ch2.SignUpdateTxProposal(utx)
	if err != nil {
		t.Error(err)
	}

	amt, err := ch1.VerifyUpdateTxProposal(ev)
	if err != nil {
		t.Error(err)
	}

	if amt != 12 {
		t.Error("wrong amt:", amt)
	}

	fmt.Println(amt, err)
}
