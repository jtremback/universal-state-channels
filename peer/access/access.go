package access

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	core "github.com/jtremback/usc/core/peer"
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

type NilError struct {
	s string
}

func (e *NilError) Error() string {
	return fmt.Sprintf("%s", e.s)
}

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

	return tx.Bucket(Judges).Put(jd.Pubkey, b)
}

func GetJudge(tx *bolt.Tx, key []byte) (*core.Judge, error) {
	b := tx.Bucket(Judges).Get([]byte(key))

	if bytes.Compare(b, []byte{}) == 0 {
		return nil, &NilError{"judge not found"}
	}

	jd := &core.Judge{}
	err := json.Unmarshal(b, jd)
	if err != nil {
		return nil, err
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

	err = SetJudge(tx, acct.Judge)
	if err != nil {
		return err
	}
	return nil
}

func GetAccount(tx *bolt.Tx, key []byte) (*core.Account, error) {
	acct := &core.Account{}

	b := tx.Bucket(Accounts).Get([]byte(key))

	if bytes.Compare(b, []byte{}) == 0 {
		return nil, &NilError{"account not found"}
	}

	err := json.Unmarshal(b, acct)
	if err != nil {
		return nil, err
	}

	err = PopulateAccount(tx, acct)
	if err != nil {
		return nil, err
	}
	return acct, nil
}

func PopulateAccount(tx *bolt.Tx, acct *core.Account) error {
	judge, err := GetJudge(tx, acct.Judge.Pubkey)
	if err != nil {
		return err
	}

	acct.Judge = judge

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

	err = SetJudge(tx, cpt.Judge)
	if err != nil {
		return err
	}
	return nil
}

func GetCounterparty(tx *bolt.Tx, key []byte) (*core.Counterparty, error) {
	b := tx.Bucket(Counterparties).Get([]byte(key))

	if bytes.Compare(b, []byte{}) == 0 {
		return nil, &NilError{"counterparty not found"}
	}

	cpt := &core.Counterparty{}
	err := json.Unmarshal(b, cpt)
	if err != nil {
		return nil, err
	}

	// Populate

	err = PopulateCounterparty(tx, cpt)
	if err != nil {
		return nil, err
	}

	return cpt, nil
}

func PopulateCounterparty(tx *bolt.Tx, cpt *core.Counterparty) error {
	judge, err := GetJudge(tx, cpt.Judge.Pubkey)
	if err != nil {
		return err
	}

	cpt.Judge = judge

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

	// Judge

	err = SetJudge(tx, ch.Judge)
	if err != nil {
		return err
	}

	// Accounts

	err = SetAccount(tx, ch.Account)
	if err != nil {
		return err
	}

	err = SetCounterparty(tx, ch.Counterparty)
	if err != nil {
		return err
	}

	return nil
}

func GetChannel(tx *bolt.Tx, key string) (*core.Channel, error) {
	b := tx.Bucket(Channels).Get([]byte(key))

	if bytes.Compare(b, []byte{}) == 0 {
		return nil, &NilError{"channel not found"}
	}

	ch := &core.Channel{}
	err := json.Unmarshal(b, ch)
	if err != nil {
		return nil, err
	}

	err = PopulateChannel(tx, ch)
	if err != nil {
		return nil, err
	}
	return ch, nil
}

func PopulateChannel(tx *bolt.Tx, ch *core.Channel) error {
	acct, err := GetAccount(tx, ch.Account.Pubkey)
	if err != nil {
		return err
	}

	cpt, err := GetCounterparty(tx, ch.Counterparty.Pubkey)
	if err != nil {
		return err
	}

	jd, err := GetJudge(tx, ch.Judge.Pubkey)
	if err != nil {
		return err
	}

	ch.Account = acct
	ch.Counterparty = cpt
	ch.Judge = jd

	return nil
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
