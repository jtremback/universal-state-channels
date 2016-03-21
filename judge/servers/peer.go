package servers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/golang/protobuf/proto"
	"github.com/jtremback/usc/core/wire"
	"github.com/jtremback/usc/judge/logic"
)

type PeerHTTP struct {
	Logic *logic.PeerAPI
}

func (a *PeerHTTP) MountRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/add_channel", a.addChannel)
	mux.HandleFunc("/add_update_tx", a.addChannel)
	mux.HandleFunc("/add_follow_on_data", a.addChannel)
}

func (a *PeerHTTP) addChannel(w http.ResponseWriter, r *http.Request) {
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

func (a *PeerHTTP) getChannel(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		a.fail(w, "no body", 500)
		return
	}

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		a.fail(w, "server error", 500)
	}

	ch, err := a.Logic.GetChannel(string(b))
	if err != nil {
		a.fail(w, "server error", 500)
	}

	data, err := json.Marshal(ch)
	if err != nil {
		a.fail(w, "server error", 500)
	}

	a.send(w, data)
}

func (a *PeerHTTP) addProposedUpdateTx(w http.ResponseWriter, r *http.Request) {
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

func (a *PeerHTTP) addFollowOnData(w http.ResponseWriter, r *http.Request) {
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

	err = a.Logic.AddFollowOnTx(ev)
	if err != nil {
		a.fail(w, "server error", 500)
	}
	a.send(w, "ok")
}

func (a *PeerHTTP) fail(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")

	data := struct {
		Error string
	}{Error: msg}

	resp, _ := json.Marshal(data)
	w.WriteHeader(status)
	w.Write(resp)
}

func (a *PeerHTTP) send(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.fail(w, "oops something evil has happened", 500)
		return
	}
	w.Write(resp)
}
