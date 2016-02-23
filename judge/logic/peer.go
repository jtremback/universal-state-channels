package logic

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	"github.com/jtremback/usc-core/wire"
	"github.com/jtremback/usc-judge/access"
)

type PeerAPI struct {
	DB *bolt.DB
}

func (a *PeerAPI) AddChannel(ev *wire.Envelope) error {
	var err error
	err = a.DB.Update(func(tx *bolt.Tx) error {

		otx := &wire.OpeningTx{}
		err = proto.Unmarshal(ev.Payload, otx)
		if err != nil {
			return err
		}

		_, err = access.GetChannel(tx, otx.ChannelId)
		if err != nil {
			return errors.New("channel already exists")
		}

		acct0, err := access.GetAccount(tx, otx.Pubkeys[0])
		if err != nil {
			return err
		}

		acct1, err := access.GetAccount(tx, otx.Pubkeys[1])
		if err != nil {
			return err
		}

		judge, err := access.GetJudge(tx, acct0.Judge.Pubkey)
		if err != nil {
			return err
		}

		ch, err := judge.AddChannel(ev, otx, acct0, acct1)

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

func (a *PeerAPI) AddUpdateTx(ev *wire.Envelope) error {
	var err error
	err = a.DB.Update(func(tx *bolt.Tx) error {
		utx := &wire.UpdateTx{}
		err = proto.Unmarshal(ev.Payload, utx)
		if err != nil {
			return err
		}

		ch, err := access.GetChannel(tx, utx.ChannelId)
		if err != nil {
			return err
		}

		err = ch.VerifyUpdateTx(ev, utx)
		if err != nil {
			return err
		}

		err = ch.StartHoldPeriod(utx)
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

func (a *PeerAPI) AddCancellationTx(ev *wire.Envelope) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ctx := &wire.CancellationTx{}
		err = proto.Unmarshal(ev.Payload, ctx)
		if err != nil {
			return err
		}

		ch, err := access.GetChannel(tx, ctx.ChannelId)
		if err != nil {
			return err
		}

		ch.AddCancellationTx(ev)
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

func (a *PeerAPI) AddFollowOnTx(ev *wire.Envelope) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		fol := &wire.FollowOnTx{}
		err = proto.Unmarshal(ev.Payload, fol)
		if err != nil {
			return err
		}

		ch, err := access.GetChannel(tx, fol.ChannelId)
		if err != nil {
			return err
		}

		ch.AddFollowOnTx(ev)
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
