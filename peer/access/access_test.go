package access

import (
	"os"
	"reflect"
	"testing"

	"github.com/boltdb/bolt"
	core "github.com/jtremback/usc/core/peer"
	"github.com/jtremback/usc/core/wire"
)

func TestJudge(t *testing.T) {
	db, err := bolt.Open("/tmp/test.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("/tmp/test.db")

	err = MakeBuckets(db)
	if err != nil {
		t.Fatal(err)
	}

	jd := &core.Judge{
		Name:    "joe",
		Pubkey:  []byte{40, 40, 40},
		Address: "stoops.com:3004",
	}

	db.Update(func(tx *bolt.Tx) error {
		err := SetJudge(tx, jd)
		if err != nil {
			t.Fatal(err)
		}

		return nil
	})

	db.View(func(tx *bolt.Tx) error {
		jd2, err := GetJudge(tx, jd.Pubkey)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(jd, jd2) {
			t.Fatal("Account incorrect")
		}
		return nil
	})
}

func TestAccount(t *testing.T) {
	db, err := bolt.Open("/tmp/test.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("/tmp/test.db")

	err = MakeBuckets(db)
	if err != nil {
		t.Fatal(err)
	}

	acct := &core.Account{
		Name:    "boogie",
		Privkey: []byte{30, 30, 30},
		Pubkey:  []byte{40, 40, 40},
		Judge: &core.Judge{
			Name:    "wrong",
			Pubkey:  []byte{40, 40, 40},
			Address: "stoops.com:3004",
		},
	}

	jd := &core.Judge{
		Name:    "joe",
		Pubkey:  []byte{40, 40, 40},
		Address: "stoops.com:3004",
	}

	db.Update(func(tx *bolt.Tx) error {
		err := SetAccount(tx, acct)
		if err != nil {
			t.Fatal(err)
		}

		err = SetJudge(tx, jd)
		if err != nil {
			t.Fatal(err)
		}

		return nil
	})

	acct.Judge = jd

	db.View(func(tx *bolt.Tx) error {
		acct2, err := GetAccount(tx, acct.Pubkey)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(acct, acct2) {
			t.Fatal("Account incorrect")
		}

		fooba, err := GetAccount(tx, []byte("fooba"))
		if err != nil {
			t.Fatal(err)
		}
		if fooba != nil {
			t.Fatal("nonexistant counterparty should be nil")
		}

		return nil
	})
}

func TestCounterparty(t *testing.T) {
	db, err := bolt.Open("/tmp/test.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("/tmp/test.db")

	err = MakeBuckets(db)
	if err != nil {
		t.Fatal(err)
	}

	cpt := &core.Counterparty{
		Name:   "boogie",
		Pubkey: []byte{40, 40, 40},
		Judge: &core.Judge{
			Name:    "wrong",
			Pubkey:  []byte{40, 40, 40},
			Address: "stoops.com:3004",
		},
	}

	jd := &core.Judge{
		Name:    "joe",
		Pubkey:  []byte{40, 40, 40},
		Address: "stoops.com:3004",
	}

	db.Update(func(tx *bolt.Tx) error {
		err := SetCounterparty(tx, cpt)
		if err != nil {
			t.Fatal(err)
		}

		err = SetJudge(tx, jd)
		if err != nil {
			t.Fatal(err)
		}

		return nil
	})

	cpt.Judge = jd

	db.View(func(tx *bolt.Tx) error {
		cpt2, err := GetCounterparty(tx, cpt.Pubkey)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(cpt, cpt2) {
			t.Fatal("Counterparty incorrect")
		}

		fooba, err := GetCounterparty(tx, []byte("fooba"))
		if err != nil {
			t.Fatal(err)
		}
		if fooba != nil {
			t.Fatal("nonexistant counterparty should be nil")
		}

		return nil
	})
}

func TestChannel(t *testing.T) {
	db, err := bolt.Open("/tmp/test.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("/tmp/test.db")

	err = MakeBuckets(db)
	if err != nil {
		t.Fatal(err)
	}

	ch := &core.Channel{
		ChannelId: "xyz23",
		Phase:     2,

		OpeningTx:         &wire.OpeningTx{},
		OpeningTxEnvelope: &wire.Envelope{},

		MyProposedUpdateTx:         &wire.UpdateTx{},
		MyProposedUpdateTxEnvelope: &wire.Envelope{},

		TheirProposedUpdateTx:         &wire.UpdateTx{},
		TheirProposedUpdateTxEnvelope: &wire.Envelope{},

		LastFullUpdateTx:         &wire.UpdateTx{},
		LastFullUpdateTxEnvelope: &wire.Envelope{},

		Me:          0,
		FollowOnTxs: []*wire.Envelope{},

		Judge: &core.Judge{
			Name:    "wrong",
			Pubkey:  []byte{40, 40, 40},
			Address: "stoops.com:3004",
		},

		Account: &core.Account{
			Name:    "wrong",
			Pubkey:  []byte{40, 40, 40},
			Privkey: []byte{40, 40, 40},
			Judge: &core.Judge{
				Name:    "wrong",
				Pubkey:  []byte{40, 40, 40},
				Address: "stoops.com:3004",
			},
		},

		Counterparty: &core.Counterparty{
			Name:    "wrong",
			Pubkey:  []byte{40, 40, 40},
			Address: "stoops.com:3004",
			Judge: &core.Judge{
				Name:    "wrong",
				Pubkey:  []byte{40, 40, 40},
				Address: "stoops.com:3004",
			},
		},
	}

	jd := &core.Judge{
		Name:    "joe",
		Pubkey:  []byte{40, 40, 40},
		Address: "stoops.com:3004",
	}

	acct := &core.Account{
		Name:    "bob",
		Pubkey:  []byte{40, 40, 40},
		Privkey: []byte{40, 40, 40},
		Judge: &core.Judge{
			Name:    "joe",
			Pubkey:  []byte{40, 40, 40},
			Address: "stoops.com:3004",
		},
	}

	cpt := &core.Counterparty{
		Name:    "crunk",
		Pubkey:  []byte{40, 40, 40},
		Address: "stoops.com:3002",
		Judge: &core.Judge{
			Name:    "joe",
			Pubkey:  []byte{40, 40, 40},
			Address: "stoops.com:3004",
		},
	}

	db.Update(func(tx *bolt.Tx) error {
		err := SetChannel(tx, ch)
		if err != nil {
			t.Fatal(err)
		}

		err = SetJudge(tx, jd)
		if err != nil {
			t.Fatal(err)
		}

		err = SetAccount(tx, acct)
		if err != nil {
			t.Fatal(err)
		}

		err = SetCounterparty(tx, cpt)
		if err != nil {
			t.Fatal(err)
		}

		return nil
	})

	ch.Judge = jd
	ch.Account = acct
	ch.Counterparty = cpt

	db.View(func(tx *bolt.Tx) error {
		ch2, err := GetChannel(tx, ch.ChannelId)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(ch, ch2) {
			t.Fatal("Counterparty incorrect")
		}

		fooba, err := GetChannel(tx, "fooba")
		if err != nil {
			t.Fatal(err)
		}
		if fooba != nil {
			t.Fatal("nonexistant channel should be nil")
		}

		return nil
	})
}

func TestGetChannels(t *testing.T) {
	db, err := bolt.Open("/tmp/test.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	defer os.Remove("/tmp/test.db")

	err = MakeBuckets(db)
	if err != nil {
		t.Fatal(err)
	}

	ch := &core.Channel{
		ChannelId: "xyz23",
		Phase:     2,

		OpeningTx:         &wire.OpeningTx{},
		OpeningTxEnvelope: &wire.Envelope{},

		MyProposedUpdateTx:         &wire.UpdateTx{},
		MyProposedUpdateTxEnvelope: &wire.Envelope{},

		TheirProposedUpdateTx:         &wire.UpdateTx{},
		TheirProposedUpdateTxEnvelope: &wire.Envelope{},

		LastFullUpdateTx:         &wire.UpdateTx{},
		LastFullUpdateTxEnvelope: &wire.Envelope{},

		Me:          0,
		FollowOnTxs: []*wire.Envelope{},

		Judge: &core.Judge{
			Name:    "wrong",
			Pubkey:  []byte{40, 40, 40},
			Address: "stoops.com:3004",
		},

		Account: &core.Account{
			Name:    "wrong",
			Pubkey:  []byte{40, 40, 40},
			Privkey: []byte{40, 40, 40},
			Judge: &core.Judge{
				Name:    "wrong",
				Pubkey:  []byte{40, 40, 40},
				Address: "stoops.com:3004",
			},
		},

		Counterparty: &core.Counterparty{
			Name:    "wrong",
			Pubkey:  []byte{40, 40, 40},
			Address: "stoops.com:3004",
			Judge: &core.Judge{
				Name:    "wrong",
				Pubkey:  []byte{40, 40, 40},
				Address: "stoops.com:3004",
			},
		},
	}

	db.Update(func(tx *bolt.Tx) error {
		err := SetChannel(tx, ch)
		if err != nil {
			t.Fatal(err)
		}

		chs, err := GetChannels(tx)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(*chs[0], *ch) {
			t.Fatal(err)
		}

		return nil
	})

}
