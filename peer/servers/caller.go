package servers

import (
	"encoding/json"
	"net/http"

	"github.com/jtremback/usc/peer/logic"
)

type CallerHTTP struct {
	Logic *logic.CallerAPI
}

func (a *CallerHTTP) MountRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/propose_channel", a.proposeChannel)
	mux.HandleFunc("/confirm_channel", a.confirmChannel)
	mux.HandleFunc("/send_update_tx", a.sendUpdateTx)
	mux.HandleFunc("/confirm_update_tx", a.confirmUpdateTx)
}

func (a *CallerHTTP) proposeChannel(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		a.fail(w, "no body", 500)
		return
	}

	req := &struct {
		ChannelId          string
		State              []byte
		AccountPubkey      []byte
		CounterpartyPubkey []byte
		HoldPeriod         uint32
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		a.fail(w, "body parsing error", 500)
	}

	_, err = a.Logic.ProposeChannel(req.ChannelId, req.State, req.AccountPubkey, req.CounterpartyPubkey, req.HoldPeriod)
	if err != nil {
		a.fail(w, err.Error(), 500)
	}
}

func (a *CallerHTTP) getChannel(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		a.fail(w, "no body", 500)
		return
	}

	req := &struct {
		ChannelId string
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		a.fail(w, "body parsing error", 500)
	}

	err = a.Logic.CheckChannel(req.ChannelId)
	if err != nil {
		a.fail(w, err.Error(), 500)
	}
}

func (a *CallerHTTP) confirmChannel(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		a.fail(w, "no body", 500)
		return
	}

	req := &struct {
		ChannelId string
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		a.fail(w, "body parsing error", 500)
	}

	err = a.Logic.ConfirmChannel(req.ChannelId)
	if err != nil {
		a.fail(w, err.Error(), 500)
	}
}

func (a *CallerHTTP) sendUpdateTx(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		a.fail(w, "no body", 500)
		return
	}

	req := &struct {
		State     []byte
		ChannelId string
		Fast      bool
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		a.fail(w, "body parsing error", 500)
	}

	err = a.Logic.NewUpdateTx(req.State, req.ChannelId, req.Fast)
	if err != nil {
		a.fail(w, err.Error(), 500)
	}
}

func (a *CallerHTTP) confirmUpdateTx(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		a.fail(w, "no body", 500)
		return
	}

	req := &struct {
		ChannelId string
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		a.fail(w, "body parsing error", 500)
	}

	err = a.Logic.ConfirmUpdateTx(req.ChannelId)
	if err != nil {
		a.fail(w, err.Error(), 500)
	}
}

func (a *CallerHTTP) fail(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json")

	data := struct {
		Error string
	}{Error: msg}

	resp, _ := json.Marshal(data)
	w.WriteHeader(status)
	w.Write(resp)
}

func (a *CallerHTTP) send(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	resp, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		a.fail(w, "oops something evil has happened", 500)
		return
	}
	w.Write(resp)
}
