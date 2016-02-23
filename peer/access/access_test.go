package access

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/boltdb/bolt"
	core "github.com/jtremback/usc-core/peer"
	"github.com/jtremback/usc-core/wire"
)

func TestSetJudge(t *testing.T) {
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
	ju2 := &core.Judge{}

	db.Update(func(tx *bolt.Tx) error {
		err := SetJudge(tx, jd)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})

	db.View(func(tx *bolt.Tx) error {
		fmt.Println(string(tx.Bucket([]byte("Judges")).Get(jd.Pubkey)))
		err := json.Unmarshal(tx.Bucket([]byte("Judges")).Get(jd.Pubkey), ju2)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})

	if !reflect.DeepEqual(jd, ju2) {
		t.Fatal("structs not equal :(")
	}
}

func TestSetMyAccount(t *testing.T) {
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

	db.Update(func(tx *bolt.Tx) error {
		err := SetAccount(tx, acct)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})

	db.View(func(tx *bolt.Tx) error {
		acct2 := &core.Account{}
		json.Unmarshal(tx.Bucket([]byte("Accounts")).Get(acct.Pubkey), acct2)
		if !reflect.DeepEqual(acct, acct2) {
			t.Fatal("MyAccount incorrect")
		}

		fromDB := tx.Bucket([]byte("Judges")).Get(acct.Judge.Pubkey)
		jd := &core.Judge{}
		json.Unmarshal(fromDB, jd)

		if !reflect.DeepEqual(acct.Judge, jd) {
			t.Fatal("Judge incorrect", acct.Judge, jd, string(tx.Bucket([]byte("Judges")).Get(acct.Judge.Pubkey)))
		}
		return nil
	})
}

func TestPopulateMyAccount(t *testing.T) {
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

	db.View(func(tx *bolt.Tx) error {
		err := PopulateAccount(tx, acct)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(acct.Judge, jd) {
			t.Fatal("Judge incorrect", acct.Judge, jd)
		}
		return nil
	})
}

func TestSetTheirAccount(t *testing.T) {
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

	db.Update(func(tx *bolt.Tx) error {
		err := SetCounterparty(tx, cpt)
		if err != nil {
			t.Fatal(err)
		}
		return nil
	})

	db.View(func(tx *bolt.Tx) error {
		cpt2 := &core.Counterparty{}
		json.Unmarshal(tx.Bucket([]byte("Counterparties")).Get(cpt.Pubkey), cpt2)
		if !reflect.DeepEqual(cpt, cpt2) {
			t.Fatal("TheirAccount incorrect")
		}

		fromDB := tx.Bucket([]byte("Judges")).Get(cpt.Judge.Pubkey)
		jd := &core.Judge{}
		json.Unmarshal(fromDB, jd)

		if !reflect.DeepEqual(cpt.Judge, jd) {
			t.Fatal("Judge incorrect", cpt.Judge, jd, string(tx.Bucket([]byte("Judges")).Get(cpt.Judge.Pubkey)))
		}
		return nil
	})
}

func TestPopulateTheirAccount(t *testing.T) {
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

	db.View(func(tx *bolt.Tx) error {
		err := PopulateCounterparty(tx, cpt)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(cpt.Judge, jd) {
			t.Fatal("Judge incorrect", cpt.Judge, jd)
		}
		return nil
	})
}

func TestSetChannel(t *testing.T) {
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

		ProposedUpdateTx:         &wire.UpdateTx{},
		ProposedUpdateTxEnvelope: &wire.Envelope{},

		LastFullUpdateTx:         &wire.UpdateTx{},
		LastFullUpdateTxEnvelope: &wire.Envelope{},

		Me:          0,
		FollowOnTxs: [][]byte{[]byte{80, 80}},

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
		return nil
	})

	db.View(func(tx *bolt.Tx) error {
		ch2 := &core.Channel{}
		json.Unmarshal(tx.Bucket([]byte("Channels")).Get([]byte(ch.ChannelId)), ch2)
		if !reflect.DeepEqual(ch, ch2) {
			t.Fatal("Channel incorrect")
		}
		juJson := tx.Bucket([]byte("Judges")).Get(ch.Judge.Pubkey)
		jd := &core.Judge{}
		json.Unmarshal(juJson, jd)

		if !reflect.DeepEqual(ch.Judge, jd) {
			t.Fatal("Judge incorrect")
		}

		maJson := tx.Bucket([]byte("Accounts")).Get(ch.Account.Pubkey)
		acct := &core.Account{}
		json.Unmarshal(maJson, acct)

		if !reflect.DeepEqual(ch.Account, acct) {
			t.Fatal("MyAccount incorrect")
		}

		taJson := tx.Bucket([]byte("Counterparties")).Get(ch.Counterparty.Pubkey)
		cpt := &core.Counterparty{}
		json.Unmarshal(taJson, cpt)

		if !reflect.DeepEqual(ch.Counterparty, cpt) {
			t.Fatal("TheirAccount incorrect")
		}
		return nil
	})
}

func TestPopulateChannel(t *testing.T) {
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

		ProposedUpdateTx:         &wire.UpdateTx{},
		ProposedUpdateTxEnvelope: &wire.Envelope{},

		LastFullUpdateTx:         &wire.UpdateTx{},
		LastFullUpdateTxEnvelope: &wire.Envelope{},

		Me:          0,
		FollowOnTxs: [][]byte{[]byte{80, 80}},

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
		Name:    "crow",
		Pubkey:  []byte{40, 40, 40},
		Privkey: []byte{40, 40, 40},
		Judge: &core.Judge{
			Name:    "joe",
			Pubkey:  []byte{40, 40, 40},
			Address: "stoops.com:3004",
		},
	}

	cpt := &core.Counterparty{
		Name:    "flerb",
		Pubkey:  []byte{40, 40, 40},
		Address: "stoops.com:3004",
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

		err = SetAccount(tx, acct)
		if err != nil {
			t.Fatal(err)
		}

		err = SetJudge(tx, jd)
		if err != nil {
			t.Fatal(err)
		}

		err = SetCounterparty(tx, cpt)
		if err != nil {
			t.Fatal(err)
		}

		return nil
	})

	db.View(func(tx *bolt.Tx) error {
		err = PopulateChannel(tx, ch)
		if err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(ch.Judge, jd) {
			t.Fatal("Judge incorrect", ch.Judge, jd)
		}

		if !reflect.DeepEqual(ch.Account, acct) {
			t.Fatal("MyAccount incorrect", ch.Account, acct)
		}

		if !reflect.DeepEqual(ch.Counterparty, cpt) {
			t.Fatal("TheirAccount incorrect", ch.Counterparty, cpt)
		}

		return nil
	})
}
