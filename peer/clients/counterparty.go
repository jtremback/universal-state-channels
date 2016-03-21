package clients

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/jtremback/usc/core/wire"
)

type CounterpartyHTTP struct{}

func (a *CounterpartyHTTP) sendEnvelope(ev *wire.Envelope, address string) error {
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

func (a *CounterpartyHTTP) AddChannel(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_channel")
}

func (a *CounterpartyHTTP) AddProposedUpdateTx(ev *wire.Envelope, address string) error {
	return a.sendEnvelope(ev, address+"/add_update_tx")
}
