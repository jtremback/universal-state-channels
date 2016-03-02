package logic

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	core "github.com/jtremback/usc/core/peer"
	"github.com/jtremback/usc/core/wire"
	"github.com/jtremback/usc/peer/access"
)

type CallerAPI struct {
	DB                 *bolt.DB
	CounterpartyClient CounterpartyClient
	JudgeClient        JudgeClient
}

type JudgeClient interface {
	GetFinalUpdateTx(string) (*wire.Envelope, error)
	AddChannel(*wire.Envelope, string) error
	AddCancellationTx(*wire.Envelope, string) error
	AddUpdateTx(*wire.Envelope, string) error
	AddFollowOnTx(*wire.Envelope, string) error
}

type CounterpartyClient interface {
	AddChannel(*wire.Envelope, string) error
	AddUpdateTx(*wire.Envelope, string) error
}

// ProposeChannel is called to propose a new channel. It creates and signs an
// OpeningTx, sends it to the Counterparty and saves it in a new Channel.
func (a *CallerAPI) ProposeChannel(state []byte, myPubkey []byte, theirPubkey []byte, holdPeriod uint32) error {
	var err error
	cpt := &core.Counterparty{}
	acct := &core.Account{}
	err = a.DB.Update(func(tx *bolt.Tx) error {
		acct, err = access.GetAccount(tx, myPubkey)
		if err != nil {
			return err
		}

		cpt, err = access.GetCounterparty(tx, theirPubkey)
		if err != nil {
			return err
		}

		otx, err := acct.NewOpeningTx(cpt, state, holdPeriod)
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

		err = a.CounterpartyClient.AddChannel(ev, cpt.Address)
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

// ConfirmChannel is called on Channels which are in phase PENDING_OPEN. It signs
// the Channel's OpeningTx, sends it to the Judge, and puts the Channel into
// phase OPEN.
func (a *CallerAPI) ConfirmChannel(channelID string) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		ch, err = access.GetChannel(tx, channelID)
		if err != nil {
			return err
		}

		ch.Account.AppendSignature(ch.OpeningTxEnvelope)

		access.SetChannel(tx, ch)
		if err != nil {
			return errors.New("database error")
		}

		err = a.JudgeClient.AddChannel(ch.OpeningTxEnvelope, ch.Judge.Address)
		if err != nil {
			return err
		}

		return nil
	})
}

// OpenChannel is called on Channels which are in phase PENDING_OPEN. It checks
// an OpeningTx signed by the Judge, and if everything is correct puts the Channel
// into phase OPEN.
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

// NewUpdateTx is called on Channels which are in phase OPEN. It makes a new UpdateTx,
// signs it, saves it as MyProposedUpdateTx, and sends it to the Counterparty.
func (a *CallerAPI) NewUpdateTx(state []byte, channelID string, fast bool) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch := &core.Channel{}
		ch, err = access.GetChannel(tx, channelID)
		if err != nil {
			return err
		}

		utx := ch.NewUpdateTx(state, fast)

		ev, err := core.SerializeUpdateTx(utx)
		if err != nil {
			return errors.New("server error")
		}

		ch.SignProposedUpdateTx(ev, utx)

		err = a.CounterpartyClient.AddChannel(ev, ch.Counterparty.Address)
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

// ConfirmUpdateTx cosigns the Channel's TheirProposedUpdateTx, saves it to
// LastFullUpdateTx, and sends it to the Counterparty.
func (a *CallerAPI) ConfirmUpdateTx(channelID string) error {
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch, err := access.GetChannel(tx, channelID)
		if err != nil {
			return err
		}

		ev := ch.CosignProposedUpdateTx()

		err = a.CounterpartyClient.AddUpdateTx(ev, ch.Counterparty.Address)
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

// CheckFinalUpdateTx checks with the Judge to see if the Counterparty has posted
// an UpdateTx. If the UpdateTx from the Judge has a lower SequenceNumber than
// LastFullUpdateTx, we send LastFullUpdateTx to the Judge.
func (a *CallerAPI) CheckFinalUpdateTx(channelID string) error {
	// var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ch, err := access.GetChannel(tx, channelID)
		if err != nil {
			return err
		}

		ev, err := a.JudgeClient.GetFinalUpdateTx(ch.Judge.Address)
		if err != nil {
			return err
		}

		utx := &wire.UpdateTx{}
		err = proto.Unmarshal(ev.Payload, utx)
		if err != nil {
			return err
		}

		newerUpdateTx, err := ch.AddFinalUpdateTx(ev, utx)
		if err != nil {
			return err
		}

		if newerUpdateTx != nil {
			err = a.JudgeClient.AddUpdateTx(newerUpdateTx, ch.Judge.Address)
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
