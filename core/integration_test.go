package wallet

import (
	"fmt"
	"testing"

	j "github.com/jtremback/usc/core/judge"
	c "github.com/jtremback/usc/core/peer"
)

// Judge's computer

var j_judge = &j.Judge{
	Name:    "sffcu",
	Pubkey:  []byte{197, 198, 13, 156, 213, 181, 160, 15, 105, 7, 66, 222, 66, 15, 212, 8, 172, 55, 20, 47, 34, 182, 117, 106, 213, 203, 6, 172, 119, 66, 87, 170},
	Privkey: []byte{244, 9, 180, 60, 13, 13, 60, 215, 158, 30, 236, 128, 111, 107, 44, 54, 75, 151, 209, 13, 20, 19, 58, 42, 162, 147, 207, 0, 189, 188, 4, 136, 197, 198, 13, 156, 213, 181, 160, 15, 105, 7, 66, 222, 66, 15, 212, 8, 172, 55, 20, 47, 34, 182, 117, 106, 213, 203, 6, 172, 119, 66, 87, 170},
}

var j_c1 = &j.Account{
	Name:   "alfred",
	Pubkey: []byte{71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	Judge:  j_judge,
}

var j_c2 = &j.Account{
	Name:   "billary",
	Pubkey: []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Judge:  j_judge,
}

// Client 1's computer

var c1_judge = &c.Judge{
	Name:   "sffcu",
	Pubkey: []byte{197, 198, 13, 156, 213, 181, 160, 15, 105, 7, 66, 222, 66, 15, 212, 8, 172, 55, 20, 47, 34, 182, 117, 106, 213, 203, 6, 172, 119, 66, 87, 170},
}

var c1_Account = &c.Account{
	Name:    "alfred",
	Pubkey:  []byte{71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	Privkey: []byte{147, 131, 100, 59, 112, 77, 196, 211, 124, 170, 199, 79, 190, 194, 175, 244, 1, 9, 48, 255, 200, 168, 138, 165, 187, 46, 251, 28, 183, 13, 214, 5, 71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	Judge:   c1_judge,
}

var c1_Counterparty = &c.Counterparty{
	Name:   "billary",
	Pubkey: []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Judge:  c1_judge,
}

// Client 2's computer

var c2_judge = &c.Judge{
	Name:   "sffcu",
	Pubkey: []byte{197, 198, 13, 156, 213, 181, 160, 15, 105, 7, 66, 222, 66, 15, 212, 8, 172, 55, 20, 47, 34, 182, 117, 106, 213, 203, 6, 172, 119, 66, 87, 170},
}

var c2_Account = &c.Account{
	Name:    "billary",
	Pubkey:  []byte{166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Privkey: []byte{184, 174, 56, 197, 104, 10, 100, 13, 194, 229, 111, 227, 49, 49, 126, 232, 117, 100, 207, 170, 154, 36, 118, 153, 143, 150, 182, 228, 98, 161, 144, 112, 166, 179, 85, 111, 208, 182, 235, 76, 4, 45, 157, 209, 98, 106, 201, 245, 59, 25, 255, 99, 66, 25, 135, 20, 5, 86, 82, 72, 97, 212, 177, 132},
	Judge:   c2_judge,
}

var c2_Counterparty = &c.Counterparty{
	Name:   "alfred",
	Pubkey: []byte{71, 153, 85, 86, 207, 54, 51, 205, 34, 228, 234, 81, 223, 175, 82, 180, 154, 154, 29, 46, 181, 45, 223, 143, 205, 48, 159, 75, 237, 51, 200, 0},
	Judge:  c2_judge,
}

func Test(t *testing.T) {
	otx, err := c1_Account.NewOpeningTx("shibby", c1_Counterparty, []byte{166, 179}, 86400)
	if err != nil {
		t.Fatal(err)
	}
	ev, err := c.SerializeOpeningTx(otx)
	if err != nil {
		t.Fatal(err)
	}

	c1_Account.AppendSignature(ev)

	ch1, err := c.NewChannel(ev, otx, c1_Account, c1_Counterparty)
	if err != nil {
		t.Fatal(err)
	}

	// --- Send to second party ---

	err = c2_Account.CheckOpeningTx(ev, c2_Counterparty)
	if err != nil {
		t.Fatal(err)
	}

	ch2, err := c.NewChannel(ev, otx, c2_Account, c2_Counterparty)
	if err != nil {
		t.Fatal(err)
	}
	c2_Account.AppendSignature(ev)

	// --- Send to judge ---

	jch, err := j_judge.AddChannel(ev, otx, j_c1, j_c2)
	if err != nil {
		t.Fatal(err)
	}

	jch.Confirm()

	// --- Send back to accounts ---

	ch1.Open(jch.OpeningTxEnvelope, jch.OpeningTx)
	if err != nil {
		t.Fatal(err)
	}

	ch2.Open(jch.OpeningTxEnvelope, jch.OpeningTx)
	if err != nil {
		t.Fatal(err)
	}

	// --- Make update tx

	utx := ch1.NewUpdateTx([]byte{164, 179}, false)

	utxEv, err := c.SerializeUpdateTx(utx)
	if err != nil {
		t.Fatal(err)
	}

	ch1.SignProposedUpdateTx(utxEv, utx)

	// --- Send to second party ---
	fmt.Println(utxEv.Signatures[0])
	err = ch2.AddProposedUpdateTx(utxEv, utx)
	if err != nil {
		t.Fatal(err)
	}

	ch2.CosignProposedUpdateTx()
	fmt.Println(utxEv.Signatures[1])

	//, --- Make follow on tx

	ftx := ch2.NewFollowOnTx([]byte{0, 4})

	ftxEv, err := c.SerializeFollowOnTx(ftx)
	if err != nil {
		t.Fatal(err)
	}

	ch2.AddFollowOnTx(ftxEv)

	// // --- Send to judge ---

	// err = jch.AddProposedUpdateTx(utxEv, utx)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// jch.Judge.AppendSignature(utxEv)

	// // --- Back to participants ---

	// _, err = ch1.AddFullUpdateTx(utxEv, utx)
	// if err != nil {
	// 	t.Fatal(err)
	// }

	// _, err = ch2.AddFullUpdateTx(utxEv, utx)
	// if err != nil {
	// 	t.Fatal(err)
	// }
}

// Extra keys
// &[197 198 13 156 213 181 160 15 105 7 66 222 66 15 212 8 172 55 20 47 34 182 117 106 213 203 6 172 119 66 87 170] &[244 9 180 60 13 13 60 215 158 30 236 128 111 107 44 54 75 151 209 13 20 19 58 42 162 147 207 0 189 188 4 136 197 198 13 156 213 181 160 15 105 7 66 222 66 15 212 8 172 55 20 47 34 182 117 106 213 203 6 172 119 66 87 170]
// &[236 129 33 67 119 101 27 246 101 161 109 184 246 50 2 214 184 162 40 197 194 196 212 210 163 136 39 229 123 204 82 25] &[97 111 164 221 195 25 249 6 17 161 159 191 252 118 241 114 92 113 7 100 234 111 160 131 230 22 181 67 197 183 9 99 236 129 33 67 119 101 27 246 101 161 109 184 246 50 2 214 184 162 40 197 194 196 212 210 163 136 39 229 123 204 82 25]
// &[118 97 30 186 23 231 51 77 244 88 148 216 9 177 104 120 183 209 212 48 44 133 220 62 24 92 165 7 153 68 194 83] &[117 54 222 53 77 11 219 41 154 161 185 104 208 248 30 59 132 230 116 108 150 60 215 9 221 101 210 53 150 159 129 174 118 97 30 186 23 231 51 77 244 88 148 216 9 177 104 120 183 209 212 48 44 133 220 62 24 92 165 7 153 68 194 83]
