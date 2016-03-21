package access

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/boltdb/bolt"
	core "github.com/jtremback/usc/core/judge"
)

// compound index types
type ssb struct {
	A string
	B string
	C []byte
}

var (
	Indexes  []byte = []byte("Indexes")
	Channels []byte = []byte("Channels")
	Judges   []byte = []byte("Judges")
	Accounts []byte = []byte("Accounts")
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
	b := tx.Bucket(Accounts).Get([]byte(key))

	if bytes.Compare(b, []byte{}) == 0 {
		return nil, &NilError{"account not found"}
	}

	acct := &core.Account{}
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
	jd, err := GetJudge(tx, acct.Judge.Pubkey)
	if err != nil {
		return err
	}

	acct.Judge = jd

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

	err = SetAccount(tx, ch.Accounts[0])
	if err != nil {
		return err
	}

	err = SetAccount(tx, ch.Accounts[1])
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
	acct0, err := GetAccount(tx, ch.Accounts[0].Pubkey)
	if err != nil {
		return err
	}

	acct1, err := GetAccount(tx, ch.Accounts[1].Pubkey)
	if err != nil {
		return err
	}

	jd, err := GetJudge(tx, ch.Judge.Pubkey)
	if err != nil {
		return err
	}

	ch.Accounts[0] = acct0
	ch.Accounts[1] = acct1
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
