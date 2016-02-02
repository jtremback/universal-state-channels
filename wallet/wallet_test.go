package wallet

import (
	// "fmt"
	"testing"
)

var escrow = &EscrowProvider{
	Name:    "holding",
	Pubkey:  []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Privkey: []byte{184, 174, 56, 197, 104, 10, 100, 13, 194, 229, 111, 227, 49, 49, 126, 232, 117, 100, 207, 170, 154, 36, 118, 153, 143, 150, 182, 228, 98, 161, 144, 112, 166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
}

var escrowp = &EscrowProvider{
	Name:   "holding",
	Pubkey: []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
}

var account1 = &Account{
	Name:           "alfred",
	Pubkey:         []byte{71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	Privkey:        []byte{147, 131, 100, 59, 112, 77, 196, 211, 124, 170, 199, 79, 190, 194, 175, 244, 1, 9, 48, 255, 200, 168, 138, 165, 187, 46, 251, 28, 183, 13, 214, 5, 71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	EscrowProvider: escrowp,
}

var account1p = &Account{
	Name:           "alfred",
	Pubkey:         []byte{71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	EscrowProvider: escrowp,
}

var account2 = &Account{
	Name:           "billary",
	Pubkey:         []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Privkey:        []byte{184, 174, 56, 197, 104, 10, 100, 13, 194, 229, 111, 227, 49, 49, 126, 232, 117, 100, 207, 170, 154, 36, 118, 153, 143, 150, 182, 228, 98, 161, 144, 112, 166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	EscrowProvider: escrowp,
}

var account2p = &Account{
	Name:           "billary",
	Pubkey:         []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	EscrowProvider: escrowp,
}

func Test(t *testing.T) {
	otx, err := account1.NewOpeningTx([]*Account{account2}, []byte{166, 179}, 86400)
	if err != nil {
		t.Error(err)
	}

	ev, err := account1.SignOpeningTx(otx)
	if err != nil {
		t.Error(err)
	}

	// --- Send to second party ---

	ev, otx, err = account2.ConfirmOpeningTx(ev)
	if err != nil {
		t.Error(err)
	}

	// --- Send to escrow provider ---

	ev, otx, err = escrow.VerifyOpeningTx(ev)
	if err != nil {
		t.Error(err)
	}

	ech, err := escrow.NewChannel(ev, []*Account{account1p, account2p})
	if err != nil {
		t.Error(err)
	}

	// --- Back to accounts ---

	ch1, err := account1.NewChannel(ev, []*Account{account2p})
	if err != nil {
		t.Error(err)
	}

	ch2, err := account2.NewChannel(ev, []*Account{account1p})
	if err != nil {
		t.Error(err)
	}

	// Make update tx

	utx, err := ch1.NewUpdateTx([]byte{164, 179}, false)
	if err != nil {
		t.Error(err)
	}

	ev, err = ch1.SignUpdateTx(utx)
	if err != nil {
		t.Error(err)
	}

	// --- Send to second party ---

	ev, utx, err = ch2.ConfirmUpdateTx(ev)
	if err != nil {
		t.Error(err)
	}

	// --- Send to escrow provider ---

	ev, utx, err = ech.VerifyUpdateTx(ev)
	if err != nil {
		t.Error(err)
	}
}
