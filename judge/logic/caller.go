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

		ch.Judge.AppendSignature(ch.OpeningTxEnvelope)

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		return nil
	})
}

func (a *CallerAPI) CloseChannel(chID string, i int) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		ch, err = access.GetChannel(tx, chID)
		if err != nil {
			return err
		}

		ev := ch.UpdateTxEnvelopes[i]
		utx := ch.UpdateTxs[i]

		ch.Judge.AppendSignature(ev)

		ch.LastFullUpdateTx = utx
		ch.LastFullUpdateTxEnvelope = ev

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		return nil
	})
}
