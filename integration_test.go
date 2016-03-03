package test

import (
	"fmt"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/jtremback/usc/core/wire"
	judgeLogic "github.com/jtremback/usc/judge/logic"
	peerLogic "github.com/jtremback/usc/peer/logic"
)

type Peer struct {
	CallerAPI       *peerLogic.CallerAPI
	CounterpartyAPI *peerLogic.CounterpartyAPI
}

type Judge struct {
	CallerAPI *judgeLogic.CallerAPI
	PeerAPI   *judgeLogic.PeerAPI
}

type CounterpartyClient struct {
	Peer *Peer
}

func (a *CounterpartyClient) AddChannel(ev *wire.Envelope, address string) error {
	return nil
}

func (a *CounterpartyClient) AddUpdateTx(ev *wire.Envelope, address string) error {
	return nil
}

type JudgeClient struct {
	Judge *Judge
}

func (a *JudgeClient) GetFinalUpdateTx(address string) (*wire.Envelope, error) {
	fmt.Println("shibby")
	return nil, nil
}

func (a *JudgeClient) AddChannel(ev *wire.Envelope, address string) error {
	fmt.Println("shibby")
	return nil
}

func (a *JudgeClient) AddCancellationTx(ev *wire.Envelope, address string) error {
	fmt.Println("shibby")
	return nil
}

func (a *JudgeClient) AddUpdateTx(ev *wire.Envelope, address string) error {
	fmt.Println("shibby")
	return nil
}

func (a *JudgeClient) AddFollowOnTx(ev *wire.Envelope, address string) error {
	fmt.Println("shibby")
	return nil
}

func TestIntegration(t *testing.T) {
	p1DB, err := bolt.Open("/tmp/p1.db", 0600, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer p1DB.Close()

	p2DB, err := bolt.Open("/tmp/p2.db", 0600, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer p2DB.Close()

	jDB, err := bolt.Open("/tmp/j.db", 0600, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer jDB.Close()

	p1 := &Peer{
		CallerAPI: &peerLogic.CallerAPI{
			DB: p1DB,
		},
		CounterpartyAPI: &peerLogic.CounterpartyAPI{
			DB: p1DB,
		},
	}
	p2 := &Peer{
		CallerAPI: &peerLogic.CallerAPI{
			DB: p2DB,
		},
		CounterpartyAPI: &peerLogic.CounterpartyAPI{
			DB: p2DB,
		},
	}
	j := &Judge{
		CallerAPI: &judgeLogic.CallerAPI{
			DB: jDB,
		},
		PeerAPI: &judgeLogic.PeerAPI{
			DB: jDB,
		},
	}

	p1.CallerAPI.JudgeClient = &JudgeClient{
		Judge: j,
	}
	p1.CallerAPI.CounterpartyClient = &CounterpartyClient{
		Peer: p2,
	}

	p2.CallerAPI.JudgeClient = &JudgeClient{
		Judge: j,
	}
	p2.CallerAPI.CounterpartyClient = &CounterpartyClient{
		Peer: p1,
	}

	p1.CallerAPI.ShitHammer()

	fmt.Println(p1, p2, j)
}
