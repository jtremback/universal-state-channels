package wallet

import (
	"bytes"
	"errors"
	"github.com/jtremback/upc-core/wallet/schema"
	"testing"
)

var account1 = &Account{
	Name:    "alfred",
	Pubkey:  []byte{71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	Privkey: []byte{147, 131, 100, 59, 112, 77, 196, 211, 124, 170, 199, 79, 190, 194, 175, 244, 1, 9, 48, 255, 200, 168, 138, 165, 187, 46, 251, 28, 183, 13, 214, 5, 71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
}

var account2 = &Account{
	Name:    "billary",
	Pubkey:  []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Privkey: []byte{184, 174, 56, 197, 104, 10, 100, 13, 194, 229, 111, 227, 49, 49, 126, 232, 117, 100, 207, 170, 154, 36, 118, 153, 143, 150, 182, 228, 98, 161, 144, 112, 166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
}

var escrowProvider = &schema.EscrowProvider{}

func TestConfirmUpdateTxProposal(t *testing.T) {
	otx, err := account1.NewOpeningTxProposal([]*Account{account2}, []byte{166, 179}, 86400)
	if err != nil {
		t.Error(err)
	}

	ev, err := SignOpeningTxProposal(otx, account1)
	if err != nil {
		t.Error(err)
	}

	// --- Send to second party ---

	ev, err = ConfirmOpeningTx(ev, account2, func(ost []byte) error {
		if bytes.Compare(ost, []byte{166, 179}) != 0 {
			return errors.New("opening tx state is incorrect")
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	ch2, err := NewChannel(ev, account2, peer1)
	if err != nil {
		t.Error(err)
	}

	// --- Send back to first party ---

	ch1, err := NewChannel(ev, account1, peer2)
	if err != nil {
		t.Error(err)
	}

	// --- Send to escrow provider ---

	otx, err = VerifyOpeningTx(ev, func(ost []byte) error {
		if bytes.Compare(ost, []byte{166, 179}) != 0 {
			return errors.New("opening tx state is incorrect")
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	utx, err := ch1.NewUpdateTxProposal([]byte{164, 179})
	if err != nil {
		t.Error(err)
	}

	ev, err = ch1.SignUpdateTxProposal(utx)
	if err != nil {
		t.Error(err)
	}

	err = ch2.ConfirmUpdateTx(ev, func(ost []byte, ust []byte) error {
		if bytes.Compare(ost, []byte{166, 179}) != 0 {
			return errors.New("opening tx state is incorrect")
		}
		if bytes.Compare(ust, []byte{164, 179}) != 0 {
			return errors.New("update txstate is incorrect")
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}

	// --- Send to escrow provider ---

	err = ep
}
