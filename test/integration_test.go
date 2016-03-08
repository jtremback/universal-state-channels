package test

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/jtremback/usc/core/wire"
	judgeAccess "github.com/jtremback/usc/judge/access"
	judgeLogic "github.com/jtremback/usc/judge/logic"
	peerAccess "github.com/jtremback/usc/peer/access"
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
		t.Fatal(err)
	}
	peerAccess.MakeBuckets(p1DB)
	defer p1DB.Close()
	defer os.Remove("/tmp/p1.db")

	p2DB, err := bolt.Open("/tmp/p2.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	peerAccess.MakeBuckets(p2DB)
	defer p2DB.Close()
	defer os.Remove("/tmp/p2.db")

	jDB, err := bolt.Open("/tmp/j.db", 0600, nil)
	if err != nil {
		t.Fatal(err)
	}
	judgeAccess.MakeBuckets(jDB)
	defer jDB.Close()
	defer os.Remove("/tmp/j.db")

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

	jd1, err := j.CallerAPI.NewJudge("jd1")
	if err != nil {
		t.Fatal(err)
	}

	p1.CallerAPI.AddJudge(jd1.Name, jd1.Pubkey, "https://judge.com/")
	if err != nil {
		t.Fatal(err)
	}

	p2.CallerAPI.AddJudge(jd1.Name, jd1.Pubkey, "https://judge.com/")
	if err != nil {
		t.Fatal(err)
	}

	acct1, err := p1.CallerAPI.NewAccount("acct1", jd1.Pubkey)
	if err != nil {
		t.Fatal(err)
	}

	acct2, err := p2.CallerAPI.NewAccount("acct2", jd1.Pubkey)
	if err != nil {
		t.Fatal(err)
	}

	err = p1.CallerAPI.AddCounterparty("cpt1", jd1.Pubkey, acct2.Pubkey, "2.com")
	if err != nil {
		t.Fatal(err)
	}

	err = p2.CallerAPI.AddCounterparty("cpt2", jd1.Pubkey, acct1.Pubkey, "1.com")
	if err != nil {
		t.Fatal(err)
	}

	ch, err := p1.CallerAPI.ProposeChannel([]byte{20}, acct1.Pubkey, acct2.Pubkey, 23)
	if err != nil {
		t.Fatal(err)
	}

	err = p1.CallerAPI.ConfirmChannel(ch.ChannelId)
	if err != nil {
		t.Fatal(err)
	}

	chs, err := p1.CallerAPI.ViewChannels()

	b, err := json.Marshal(chs)
	if err != nil {
		t.Fatal(err)
	}

	b, err = json.MarshalIndent(chs, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(string(b))
}
