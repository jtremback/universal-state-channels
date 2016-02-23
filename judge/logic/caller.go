package logic

import (
	"errors"

	"github.com/boltdb/bolt"
	core "github.com/jtremback/usc/core/judge"
	"github.com/jtremback/usc/judge/access"
)

type CallerAPI struct {
	DB *bolt.DB
}

func (a *CallerAPI) ConfirmChannel(chID string) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		ch, err = access.GetChannel(tx, chID)
		if err != nil {
			return err
		}

		ch.OpeningTxEnvelope = ch.Judge.SignEnvelope(ch.OpeningTxEnvelope)

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		return nil
	})
}

func (a *CallerAPI) CloseChannel(chID string) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		ch, err = access.GetChannel(tx, chID)
		if err != nil {
			return err
		}

		ch.CloseChannel()

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		return nil
	})
}
