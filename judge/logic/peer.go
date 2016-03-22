package logic

import (
	"errors"

	"github.com/boltdb/bolt"
	"github.com/golang/protobuf/proto"
	core "github.com/jtremback/usc/core/judge"
	"github.com/jtremback/usc/core/wire"
	"github.com/jtremback/usc/judge/access"
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

		_, nilErr := access.GetChannel(tx, otx.ChannelId)
		if nilErr == nil {
			return errors.New("channel already exists")
		}
		_, ok := nilErr.(*access.NilError)
		if !ok {
			return err
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
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

// Gets a channel for a peer and sanitizes it to be sent to them.
func (a *PeerAPI) GetChannel(chId string) (*core.Channel, error) {
	var err error
	ch := &core.Channel{}
	err = a.DB.View(func(tx *bolt.Tx) error {
		ch, err = access.GetChannel(tx, chId)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	ch.Sanitize()

	return ch, nil
}

func (a *PeerAPI) AddFullUpdateTx(ev *wire.Envelope) error {
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

		err = ch.AddFullUpdateTx(ev, utx)
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

func (a *PeerAPI) AddClosingTx(ev *wire.Envelope) error {
	var err error
	return a.DB.Update(func(tx *bolt.Tx) error {
		ctx := &wire.ClosingTx{}
		err = proto.Unmarshal(ev.Payload, ctx)
		if err != nil {
			return err
		}

		ch, err := access.GetChannel(tx, ctx.ChannelId)
		if err != nil {
			return err
		}

		ch.AddClosingTx(ev)
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

// func (a *PeerAPI) GetLog() ([]*wire.Envelope, error) {
// 	var err error
// 	return a.DB.View(func(tx *bolt.Tx) error {
//         access
// 	})
// 	return nil, nil
// }
