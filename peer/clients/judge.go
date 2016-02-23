package clients

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/jtremback/usc/core/wire"
)

type JudgeHTTP struct{}

func (a *JudgeHTTP) SendEnvelope(ev *wire.Envelope, address string) error {
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

func (a *JudgeHTTP) AddChannel(ev *wire.Envelope, address string) error {
	return a.SendEnvelope(ev, address+"/add_channel")
}

func (a *JudgeHTTP) AddCancellationTx(ev *wire.Envelope, address string) error {
	return a.SendEnvelope(ev, address+"/add_cancellation_tx")
}

func (a *JudgeHTTP) AddUpdateTx(ev *wire.Envelope, address string) error {
	return a.SendEnvelope(ev, address+"/add_update_tx")
}

func (a *JudgeHTTP) AddFollowOnTx(ev *wire.Envelope, address string) error {
	return a.SendEnvelope(ev, address+"/add_follow_on_tx")
}
