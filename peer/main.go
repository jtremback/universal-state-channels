package main

import (
	"fmt"
	"net/http"

	"github.com/boltdb/bolt"
	"github.com/jtremback/usc-peer/clients"
	"github.com/jtremback/usc-peer/logic"
	"github.com/jtremback/usc-peer/servers"
)

func main() {
	db, err := bolt.Open("main.db", 0600, nil)
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	counterpartyCl := &clients.Counterparty{}
	judgeCl := &clients.Judge{}

	callerLog := &logic.Caller{
		DB:             db,
		CounterpartyCl: counterpartyCl,
		JudgeCl:        judgeCl,
	}

	callerMux := http.NewServeMux()
	callerSrv := &servers.Caller{
		Logic: callerLog,
	}

	callerSrv.MountRoutes(callerMux)
	http.ListenAndServe(":3000", callerMux)
}
