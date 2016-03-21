package servers

import (
	"encoding/json"
	"net/http"

	"github.com/jtremback/usc/judge/logic"
)

type CallerHTTP struct {
	Logic *logic.CallerAPI
}

func (a *CallerHTTP) MountRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/confirm_channel", a.confirmChannel)
	mux.HandleFunc("/close_channel", a.closeChannel)
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

	err = a.Logic.AcceptChannel(req.ChannelId)
	if err != nil {
		a.fail(w, err.Error(), 500)
	}
}

func (a *CallerHTTP) closeChannel(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		a.fail(w, "no body", 500)
		return
	}

	req := &struct {
		ChannelId     string
		UpdateTxIndex int
	}{}
	err := json.NewDecoder(r.Body).Decode(req)
	if err != nil {
		a.fail(w, "body parsing error", 500)
	}

	err = a.Logic.CloseChannel(req.ChannelId, req.UpdateTxIndex)
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
