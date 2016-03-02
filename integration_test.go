package test

import (
	"fmt"
	"testing"

	"github.com/boltdb/bolt"
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

}
