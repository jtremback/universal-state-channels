package access

import (
	"encoding/json"
	"errors"

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

	err = tx.Bucket(Judges).Put(jd.Pubkey, b)
	if err != nil {
		return err
	}

	return nil
}

func GetJudge(tx *bolt.Tx, key []byte) (*core.Judge, error) {
	jd := &core.Judge{}
	err := json.Unmarshal(tx.Bucket(Accounts).Get(key), jd)
	if err != nil {
		return nil, errors.New("database error")
	}
	if jd == nil {
		return nil, errors.New("judge not found")
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
	err := json.Unmarshal(tx.Bucket(Accounts).Get(key), acct)
	if err != nil {
		return nil, errors.New("database error")
	}
	if acct == nil {
		return nil, errors.New("account not found")
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

	b, err = json.Marshal(ch.Judge)
	if err != nil {
		return err
	}

	tx.Bucket(Judges).Put(ch.Judge.Pubkey, b)

	// Accounts

	b, err = json.Marshal(ch.Accounts[0])
	if err != nil {
		return err
	}

	tx.Bucket(Accounts).Put(ch.Accounts[0].Pubkey, b)

	b, err = json.Marshal(ch.Accounts[1])
	if err != nil {
		return err
	}

	tx.Bucket(Accounts).Put(ch.Accounts[1].Pubkey, b)

	return nil
}

func GetChannel(tx *bolt.Tx, key string) (*core.Channel, error) {
	ch := &core.Channel{}
	err := json.Unmarshal(tx.Bucket(Channels).Get([]byte(key)), ch)
	if err != nil {
		return nil, errors.New("database error")
	}
	if ch == nil {
		return nil, errors.New("channel not found")
	}
	err = PopulateChannel(tx, ch)
	if err != nil {
		return nil, errors.New("database error")
	}
	return ch, nil
}

func PopulateChannel(tx *bolt.Tx, ch *core.Channel) error {
	var err error
	acct0 := &core.Account{}
	err = json.Unmarshal(tx.Bucket(Accounts).Get([]byte(ch.Accounts[0].Pubkey)), acct0)
	if err != nil {
		return err
	}
	err = PopulateAccount(tx, acct0)
	if err != nil {
		return err
	}

	acct1 := &core.Account{}
	err = json.Unmarshal(tx.Bucket(Accounts).Get([]byte(ch.Accounts[1].Pubkey)), acct1)
	if err != nil {
		return err
	}
	err = PopulateAccount(tx, acct1)
	if err != nil {
		return err
	}

	jd := &core.Judge{}
	err = json.Unmarshal(tx.Bucket(Judges).Get([]byte(ch.Judge.Pubkey)), jd)
	if err != nil {
		return err
	}

	ch.Accounts[0] = acct0
	ch.Accounts[1] = acct1
	ch.Judge = jd

	return nil
}
