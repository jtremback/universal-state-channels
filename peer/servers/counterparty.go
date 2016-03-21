package servers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/jtremback/usc/core/wire"
	"github.com/jtremback/usc/peer/logic"
)

type CounterpartyHTTP struct {
	Logic *logic.CounterpartyAPI
}

func (a *CounterpartyHTTP) MountRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/add_channel", a.addChannel)
}

func (a *CounterpartyHTTP) addChannel(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		a.fail(w, "no body", 500)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		a.fail(w, "server error", 500)
	}

	ev := &wire.Envelope{}
	proto.Unmarshal(b, ev)

	err = a.Logic.AddChannel(ev)
	if err != nil {
		a.fail(w, "server error", 500)
	}
	a.send(w, "ok")
}

func (a *CounterpartyHTTP) addProposedUpdateTx(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		a.fail(w, "no body", 500)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		a.fail(w, "server error", 500)
	}

	ev := &wire.Envelope{}
	proto.Unmarshal(b, ev)

	err = a.Logic.AddProposedUpdateTx(ev)
	if err != nil {
		a.fail(w, "server error", 500)
	}
	a.send(w, "ok")
}

func (a *CounterpartyHTTP) fail(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")

	data := struct {
		Error string
	}{Error: msg}

	resp, _ := json.Marshal(data)
	w.WriteHeader(status)
	w.Write(resp)
}

func (a *CounterpartyHTTP) send(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.fail(w, "oops something evil has happened", 500)
		return
	}
	w.Write(resp)
}
