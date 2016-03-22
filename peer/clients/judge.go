package clients

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/jtremback/usc/core/wire"
)

type JudgeHTTP struct{}

func (a *JudgeHTTP) sendEnvelope(ev *wire.Envelope, address string) error {
	b, err := proto.Marshal(ev)

	resp, err := http.Post(address, "application/octet-stream", bytes.NewReader(b))
	if err != nil {
		return errors.New("network error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return errors.New("counterparty error")
	}

	return nil
}

func (a *JudgeHTTP) getEnvelope(address string) (*wire.Envelope, error) {
	resp, err := http.Post(address, "application/octet-stream", bytes.NewReader([]byte(address)))
	if err != nil {
		return nil, errors.New("network error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("judge error")
	}

	data, err := ioutil.ReadAll(resp.Body)

	ev := &wire.Envelope{}
	err = proto.Unmarshal(data, ev)
	if err != nil {
		return nil, errors.New("error parsing envelope")
	}

	return ev, nil
}

func (a *JudgeHTTP) getData(address string, key []byte) ([]byte, error) {
	resp, err := http.Post(address, "application/octet-stream", bytes.NewReader([]byte(key)))
	if err != nil {
		return nil, errors.New("network error")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, errors.New("judge error")
	}

	return ioutil.ReadAll(resp.Body)
}

func (a *JudgeHTTP) GetLastFullUpdateTx(address string) (*wire.Envelope, error) {
	ev, err := a.getEnvelope(address + "/get_final_update_tx")
	if err != nil {
		return nil, errors.New("can't reach judge")
	}
	return ev, nil
}

func (a *JudgeHTTP) AddChannel(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_channel")
}

func (a *JudgeHTTP) AddClosingTx(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_closing_tx")
}

func (a *JudgeHTTP) AddFullUpdateTx(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_update_tx")
}

func (a *JudgeHTTP) AddFollowOnTx(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_follow_on_tx")
}

func (a *JudgeHTTP) GetChannel(chId string, address string) ([]byte, error) {
	data, err := a.getData(address+"/check_account", []byte(chId))
	if err != nil {
		return nil, errors.New("can't reach judge")
	}

	return data, nil
}

// func (a *JudgeHTTP) CheckAccount(key []byte, address string) {
// 	str := base64.URLEncoding.EncodeToString(key)
// 	data, err := a.getData(address + "/check_account/" + str)
// 	if err != nil {
// 		return nil, errors.New("can't reach judge")
// 	}
// }
