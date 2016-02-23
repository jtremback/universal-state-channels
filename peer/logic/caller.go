package logic

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	core "github.com/jtremback/usc/core/peer"
	"github.com/jtremback/usc/core/wire"
	"github.com/jtremback/usc/peer/access"
	"github.com/jtremback/usc/peer/clients"
)

type CallerAPI struct {
	DB             *bolt.DB
	CounterpartyCl *clients.Counterparty
	JudgeCl        *clients.Judge
}

func (a *CallerAPI) ProposeChannel(state []byte, mpk []byte, tpk []byte, hold uint32) error {
	var err error
	cpt := &core.Counterparty{}
	acct := &core.Account{}
	err = a.DB.Update(func(tx *bolt.Tx) error {
		acct, err = access.GetAccount(tx, mpk)
		if err != nil {
			return err
		}

		cpt, err = access.GetCounterparty(tx, tpk)
		if err != nil {
			return err
		}

		otx, err := acct.NewOpeningTx(cpt, state, hold)
		if err != nil {
			return errors.New("server error")
		}

		ev, err := core.SerializeOpeningTx(otx)
		if err != nil {
			return errors.New("server error")
		}

		acct.AppendSignature(ev)

		ch, err := core.NewChannel(ev, otx, acct, cpt)
		if err != nil {
			return errors.New("server error")
		}

		err = a.CounterpartyCl.Send(ev, cpt.Address)
		if err != nil {
			return err
		}

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *CallerAPI) ConfirmChannel(chID string) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		ch, err = access.GetChannel(tx, chID)
		if err != nil {
			return err
		}

		ch.Account.AppendSignature(ch.OpeningTxEnvelope)

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		err = a.JudgeCl.Send(ch.OpeningTxEnvelope, ch.Judge.Address)
		if err != nil {
			return err
		}

		return nil
	})
}

func (a *CallerAPI) OpenChannel(ev *wire.Envelope) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		otx := &wire.OpeningTx{}
		err = proto.Unmarshal(ev.Payload, otx)
		if err != nil {
			return err
		}

		ch, err = access.GetChannel(tx, otx.ChannelId)
		if err != nil {
			return err
		}

		ch.Open(ev, otx)
		if err != nil {
			return err
		}

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		return nil
	})
}

func (a *CallerAPI) SendUpdateTx(state []byte, chID string, fast bool) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		ch, err = access.GetChannel(tx, chID)
		if err != nil {
			return err
		}

		utx := ch.NewUpdateTx(state, fast)

		ev, err := core.SerializeUpdateTx(utx)
		if err != nil {
			return errors.New("server error")
		}

		err = a.CounterpartyCl.Send(ev, ch.Counterparty.Address)
		if err != nil {
			return err
		}

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		return nil
	})
}

func (a *CallerAPI) ConfirmUpdateTx(chID string) error {
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch, err := access.GetChannel(tx, chID)
		if err != nil {
			return err
		}

		ev := ch.CosignProposedUpdateTx()

		err = a.CounterpartyCl.Send(ev, ch.Counterparty.Address)
		if err != nil {
			return err
		}

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		return nil
	})
}

func (a *CallerAPI) CheckFinalUpdateTx(ev *wire.Envelope) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		utx := &wire.UpdateTx{}
		err = proto.Unmarshal(ev.Payload, utx)
		if err != nil {
			return err
		}
		ch, err := access.GetChannel(tx, utx.ChannelId)
		if err != nil {
			return err
		}

		ev2, err := ch.CheckFinalUpdateTx(ev, utx)
		if err != nil {
			return err
		}
		if ev2 != nil {
			err = a.JudgeCl.Send(ev2, ch.Judge.Address)
			if err != nil {
				return err
			}
		}

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		return nil
	})
}

func (a *CallerAPI) AddJudge() {

}

func (a *CallerAPI) NewAccount() {

}

func (a *CallerAPI) AddCounterparty() {

}
