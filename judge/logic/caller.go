package logic

import (
	"github.com/boltdb/bolt"
	core "github.com/jtremback/usc/core/judge"
	"github.com/jtremback/usc/judge/access"
)

type CallerAPI struct {
	DB *bolt.DB
}

func (a *CallerAPI) NewJudge(
	name string,
) (*core.Judge, error) {
	var err error
	jd := &core.Judge{}
	a.DB.Update(func(tx *bolt.Tx) error {
		jd, err = core.NewJudge(name)
		if err != nil {
			return err
		}

		err = access.SetJudge(tx, jd)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return jd, nil
}

func (a *CallerAPI) AddAccount(
	name string,
	judge []byte,
	pubkey []byte,
	address string,
) error {
	return a.DB.Update(func(tx *bolt.Tx) error {
		jd, err := access.GetJudge(tx, judge)
		if err != nil {
			return err
		}

		acct := &core.Account{
			Name:    name,
			Judge:   jd,
			Pubkey:  pubkey,
			Address: address,
		}

		err = access.SetAccount(tx, acct)
		if err != nil {
			return err
		}

		return nil
	})
}

func (a *CallerAPI) AcceptChannel(chID string) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		ch, err = access.GetChannel(tx, chID)
		if err != nil {
			return err
		}

		ch.Judge.AppendSignature(ch.OpeningTxEnvelope)
		ch.Phase = core.OPEN

		access.SetChannel(tx, ch)
		if err != nil {
			return err
		}

		return nil
	})
}

func (a *CallerAPI) ViewChannels() ([]*core.Channel, error) {
	var chs []*core.Channel
	var err error
	err = a.DB.View(func(tx *bolt.Tx) error {
		chs, err = access.GetChannels(tx)

		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return chs, nil
}

func (a *CallerAPI) CloseChannel(chID string, i int) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		ch, err = access.GetChannel(tx, chID)
		if err != nil {
			return err
		}

		err = ch.Close(i)
		if err != nil {
			return err
		}

		access.SetChannel(tx, ch)
		if err != nil {
			return err
		}

		return nil
	})
}
