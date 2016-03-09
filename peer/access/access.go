package access

import (
	"bytes"
	"encoding/json"
	"errors"

	"github.com/boltdb/bolt"
	core "github.com/jtremback/usc/core/peer"
	"github.com/tv42/compound"
)

// compound index types
type ssb struct {
	A string
	B string
	C []byte
}

var (
	Indexes        []byte = []byte("Indexes")
	Channels       []byte = []byte("Channels")
	Judges         []byte = []byte("Judges")
	Accounts       []byte = []byte("Accounts")
	Counterparties []byte = []byte("Counterparties")
)

func MakeBuckets(db *bolt.DB) error {
	err := db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(Indexes)
		_, err = tx.CreateBucketIfNotExists(Channels)
		_, err = tx.CreateBucketIfNotExists(Judges)
		_, err = tx.CreateBucketIfNotExists(Accounts)
		_, err = tx.CreateBucketIfNotExists(Counterparties)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func SetJudge(tx *bolt.Tx, jd *core.Judge) error {
	b, err := json.Marshal(jd)
	if err != nil {
		return err
	}

	err = tx.Bucket(Judges).Put(jd.Pubkey, b)
	if err != nil {
		return err
	}

	return nil
}

func GetJudge(tx *bolt.Tx, key []byte) (*core.Judge, error) {
	jd := &core.Judge{}
	jason := tx.Bucket(Judges).Get([]byte(key))

	if bytes.Compare(jason, []byte{}) == 0 {
		return nil, nil
	}

	err := json.Unmarshal(jason, jd)
	if err != nil {
		return nil, errors.New("channel not found")
	}

	return jd, nil
}

func SetAccount(tx *bolt.Tx, acct *core.Account) error {
	b, err := json.Marshal(acct)
	if err != nil {
		return err
	}

	err = tx.Bucket(Accounts).Put([]byte(acct.Pubkey), b)
	if err != nil {
		return err
	}

	// Relations

	b, err = json.Marshal(acct.Judge)
	if err != nil {
		return err
	}

	err = tx.Bucket(Judges).Put(acct.Judge.Pubkey, b)
	if err != nil {
		return err
	}

	return nil
}

func GetAccount(tx *bolt.Tx, key []byte) (*core.Account, error) {
	acct := &core.Account{}

	jason := tx.Bucket(Accounts).Get([]byte(key))

	if bytes.Compare(jason, []byte{}) == 0 {
		return nil, nil
	}

	err := json.Unmarshal(jason, acct)
	if err != nil {
		return nil, errors.New("channel not found")
	}

	err = PopulateAccount(tx, acct)
	if err != nil {
		return nil, errors.New("database error")
	}
	return acct, nil
}

func PopulateAccount(tx *bolt.Tx, acct *core.Account) error {
	jd := &core.Judge{}
	err := json.Unmarshal(tx.Bucket(Judges).Get([]byte(acct.Judge.Pubkey)), jd)
	if err != nil {
		return err
	}
	acct.Judge = jd

	return nil
}

func SetCounterparty(tx *bolt.Tx, cpt *core.Counterparty) error {
	b, err := json.Marshal(cpt)
	if err != nil {
		return err
	}

	err = tx.Bucket(Counterparties).Put([]byte(cpt.Pubkey), b)
	if err != nil {
		return err
	}

	// Relations

	b, err = json.Marshal(cpt.Judge)
	if err != nil {
		return err
	}

	err = tx.Bucket(Judges).Put(cpt.Judge.Pubkey, b)
	if err != nil {
		return err
	}

	return nil
}

func GetCounterparty(tx *bolt.Tx, key []byte) (*core.Counterparty, error) {
	cpt := &core.Counterparty{}
	jason := tx.Bucket(Counterparties).Get([]byte(key))

	if bytes.Compare(jason, []byte{}) == 0 {
		return nil, nil
	}

	err := json.Unmarshal(jason, cpt)
	if err != nil {
		return nil, errors.New("channel not found")
	}

	err = PopulateCounterparty(tx, cpt)
	if err != nil {
		return nil, errors.New("database error")
	}
	return cpt, nil
}

func PopulateCounterparty(tx *bolt.Tx, cpt *core.Counterparty) error {
	jd := &core.Judge{}
	err := json.Unmarshal(tx.Bucket(Judges).Get([]byte(cpt.Judge.Pubkey)), jd)
	if err != nil {
		return err
	}

	cpt.Judge = jd

	return nil
}

func SetChannel(tx *bolt.Tx, ch *core.Channel) error {
	b, err := json.Marshal(ch)
	if err != nil {
		return err
	}

	err = tx.Bucket(Channels).Put([]byte(ch.ChannelId), b)
	if err != nil {
		return err
	}

	// Relations

	// Escrow Provider

	b, err = json.Marshal(ch.Judge)
	if err != nil {
		return err
	}

	tx.Bucket(Judges).Put(ch.Judge.Pubkey, b)

	// My Account

	b, err = json.Marshal(ch.Account)
	if err != nil {
		return err
	}

	tx.Bucket(Accounts).Put(ch.Account.Pubkey, b)

	// Their Account

	b, err = json.Marshal(ch.Counterparty)
	if err != nil {
		return err
	}

	tx.Bucket(Counterparties).Put(ch.Counterparty.Pubkey, b)

	// Indexes

	// Escrow Provider Pubkey

	err = tx.Bucket(Indexes).Put(compound.Key(ssb{
		"Judge",
		"Pubkey",
		ch.Judge.Pubkey}), []byte(ch.ChannelId))
	if err != nil {
		return err
	}

	return nil
}

func GetChannel(tx *bolt.Tx, key string) (*core.Channel, error) {
	ch := &core.Channel{}
	jason := tx.Bucket(Channels).Get([]byte(key))

	if bytes.Compare(jason, []byte{}) == 0 {
		return nil, nil
	}

	err := json.Unmarshal(jason, ch)
	if err != nil {
		return nil, errors.New("channel not found")
	}

	err = PopulateChannel(tx, ch)
	if err != nil {
		return nil, errors.New("database error")
	}
	return ch, nil
}

func GetChannels(tx *bolt.Tx) ([]*core.Channel, error) {
	var err error
	chs := []*core.Channel{}

	err = tx.Bucket(Channels).ForEach(func(k, v []byte) error {
		ch := &core.Channel{}
		err = json.Unmarshal(v, ch)
		if err != nil {
			return err
		}

		err = PopulateChannel(tx, ch)
		if err != nil {
			return errors.New("error populating channel")
		}

		chs = append(chs, ch)

		return nil
	})
	if err != nil {
		return nil, errors.New("database error")
	}
	return chs, nil
}

// func GetProposedUpdateTxs(tx *bolt.Tx) ([]*core.Channel, error) {
// 	var err error
// 	chs := []*core.Channel{}
// 	i := 0
// 	err = tx.Bucket(Channels).ForEach(func(k, v []byte) error {
// 		ch := &core.Channel{}
// 		err = json.Unmarshal(v, ch)
// 		if err != nil {
// 			return err
// 		}
// 		if bytes.Compare(ch.ProposedUpdateTxEnvelope.Signatures[ch.Me], []byte{}) == 0 {
// 			chs[i] = ch
// 			i++
// 		}
// 		return nil
// 	})
// 	if err != nil {
// 		return nil, errors.New("database error")
// 	}
// 	return chs, nil
// }

func PopulateChannel(tx *bolt.Tx, ch *core.Channel) error {
	acct := &core.Account{}
	err := json.Unmarshal(tx.Bucket(Accounts).Get([]byte(ch.Account.Pubkey)), acct)
	if err != nil {
		return err
	}
	err = PopulateAccount(tx, acct)
	if err != nil {
		return err
	}

	cpt := &core.Counterparty{}
	err = json.Unmarshal(tx.Bucket(Counterparties).Get([]byte(ch.Counterparty.Pubkey)), cpt)
	if err != nil {
		return err
	}
	err = PopulateCounterparty(tx, cpt)
	if err != nil {
		return err
	}

	jd := &core.Judge{}
	err = json.Unmarshal(tx.Bucket(Judges).Get([]byte(ch.Judge.Pubkey)), jd)
	if err != nil {
		return err
	}

	ch.Account = acct
	ch.Counterparty = cpt
	ch.Judge = jd

	return nil
}
