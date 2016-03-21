package access

import (
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/boltdb/bolt"
	core "github.com/jtremback/usc/core/judge"
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
		Name:   "joe",
		Pubkey: []byte{40, 40, 40},
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

		_, err = GetJudge(tx, []byte("fooba"))
		err, ok := err.(*NilError)
		if !ok {
			t.Fatal("nonexistant judge should return NilError")
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
		Name:   "boogie",
		Pubkey: []byte{40, 40, 40},
		Judge: &core.Judge{
			Name:    "wrong",
			Pubkey:  []byte{40, 40, 40},
			Privkey: []byte{4, 20},
		},
	}

	jd := &core.Judge{
		Name:    "joe",
		Pubkey:  []byte{40, 40, 40},
		Privkey: []byte{4, 20},
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

		_, err = GetAccount(tx, []byte("fooba"))
		err, ok := err.(*NilError)
		if !ok {
			t.Fatal("nonexistant account should return NilError")
		}

		return nil
	})
}

var ch = &core.Channel{
	ChannelId: "xyz23",
	Phase:     2,

	OpeningTx:         &wire.OpeningTx{},
	OpeningTxEnvelope: &wire.Envelope{},

	LastFullUpdateTx:         &wire.UpdateTx{},
	LastFullUpdateTxEnvelope: &wire.Envelope{},

	FollowOnTxs: []*wire.Envelope{},

	Judge: &core.Judge{
		Name:    "wrong",
		Pubkey:  []byte{40, 40, 40},
		Privkey: []byte{4, 20},
	},

	Accounts: []*core.Account{
		&core.Account{
			Name:   "wrong",
			Pubkey: []byte{0, 0, 0},
			Judge: &core.Judge{
				Name:    "wrong",
				Pubkey:  []byte{40, 40, 40},
				Privkey: []byte{4, 20},
			},
		},

		&core.Account{
			Name:   "wrong",
			Pubkey: []byte{1, 1, 1},
			Judge: &core.Judge{
				Name:    "wrong",
				Pubkey:  []byte{40, 40, 40},
				Privkey: []byte{4, 20},
			},
		},
	},
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

	jd := &core.Judge{
		Name:    "joe",
		Pubkey:  []byte{40, 40, 40},
		Privkey: []byte{4, 20},
	}

	acct0 := &core.Account{
		Name:   "bob",
		Pubkey: []byte{0, 0, 0},
		Judge: &core.Judge{
			Name:    "joe",
			Pubkey:  []byte{40, 40, 40},
			Privkey: []byte{4, 20},
		},
	}

	acct1 := &core.Account{
		Name:   "bob",
		Pubkey: []byte{1, 1, 1},
		Judge: &core.Judge{
			Name:    "joe",
			Pubkey:  []byte{40, 40, 40},
			Privkey: []byte{4, 20},
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

		err = SetAccount(tx, acct0)
		if err != nil {
			t.Fatal(err)
		}

		err = SetAccount(tx, acct1)
		if err != nil {
			t.Fatal(err)
		}

		return nil
	})

	ch.Judge = jd
	ch.Accounts[0] = acct0
	ch.Accounts[1] = acct1

	db.View(func(tx *bolt.Tx) error {
		ch2, err := GetChannel(tx, ch.ChannelId)
		if err != nil {
			t.Fatal(err)
		}

		ch2.CloseTime = time.Time{}

		if !reflect.DeepEqual(ch.CloseTime, ch2.CloseTime) {
			t.Fatal("Channel incorrect")
		}

		_, err = GetChannel(tx, "fooba")
		err, ok := err.(*NilError)
		if !ok {
			t.Fatal("nonexistant account should return NilError")
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

	db.Update(func(tx *bolt.Tx) error {
		err := SetChannel(tx, ch)
		if err != nil {
			t.Fatal(err)
		}

		chs, err := GetChannels(tx)
		if err != nil {
			t.Fatal(err)
		}

		chs[0].CloseTime = time.Time{}

		if !reflect.DeepEqual(*chs[0], *ch) {
			t.Fatal(err)
		}

		return nil
	})

}
