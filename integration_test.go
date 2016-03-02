package test

import (
	"fmt"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/jtremback/usc/core/wire"
	judgeLogic "github.com/jtremback/usc/judge/logic"
	judgeServers "github.com/jtremback/usc/judge/servers"
	peerClients "github.com/jtremback/usc/peer/clients"
	peerLogic "github.com/jtremback/usc/peer/logic"
	peerServers "github.com/jtremback/usc/peer/servers"
)

type Peer struct {
	CallerSrv       *peerServers.CallerHTTP
	CounterpartySrv *peerServers.CounterpartyHTTP
}

type Judge struct {
	CallerSrv *judgeServers.CallerHTTP
	PeerSrv   *judgeServers.PeerHTTP
}

type CounterpartyClientHarness struct {
}

type JudgeClientHarness struct {
}

func (a *JudgeClientHarness) AddChannel(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_channel")
}

func (a *JudgeClientHarness) AddCancellationTx(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_cancellation_tx")
}

func (a *JudgeClientHarness) AddUpdateTx(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_update_tx")
}

func (a *JudgeClientHarness) AddFollowOnTx(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_follow_on_tx")
}

func createPeer(db *bolt.DB) *Peer {
	counterpartyCl := &peerClients.CounterpartyHTTP{}
	judgeCl := &peerClients.JudgeHTTP{}

	callerAPI := &peerLogic.CallerAPI{
		DB:             db,
		CounterpartyCl: counterpartyCl,
		JudgeCl:        judgeCl,
	}

	callerSrv := &peerServers.CallerHTTP{
		Logic: callerAPI,
	}

	counterpartyAPI := &peerLogic.CounterpartyAPI{
		DB: db,
	}

	counterpartySrv := &peerServers.CounterpartyHTTP{
		Logic: counterpartyAPI,
	}

	return &Peer{
		CallerSrv:       callerSrv,
		CounterpartySrv: counterpartySrv,
	}
}

func createJudge(db *bolt.DB) *Judge {
	callerAPI := &judgeLogic.CallerAPI{
		DB: db,
	}

	callerSrv := &judgeServers.CallerHTTP{
		Logic: callerAPI,
	}

	peerAPI := &judgeLogic.PeerAPI{
		DB: db,
	}

	peerSrv := &judgeServers.PeerHTTP{
		Logic: peerAPI,
	}

	return &Judge{
		CallerSrv: callerSrv,
		PeerSrv:   peerSrv,
	}
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

	p1 := createPeer(p1DB)
	p2 := createPeer(p2DB)
	j := createPeer(jDB)

	fmt.Println(p1, p2, j)
}
