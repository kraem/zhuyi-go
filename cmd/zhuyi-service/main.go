package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kraem/zhuyi-go/server"
)

func main() {

	s := server.NewServer()

	r := mux.NewRouter()
	r.Handle("/status", server.StatusHandler(s)).Methods("GET")
	r.Handle("/d3/graph", server.GraphHandler(s)).Methods("GET")
	r.Handle("/unlinked", server.UnlinkedHandler(s)).Methods("GET")
	r.Handle("/node/add", server.AddNodeHandler(s)).Methods("POST")
	r.Handle("/node/del", server.DelNodeHandler(s)).Methods("POST")
	// TODO Handle options like this for all endpoints
	r.Handle("/node/add", server.AddNodeHandlerOptions(s)).Methods("OPTIONS")

	log.Fatal(http.ListenAndServe(s.Cfg.Addr, r))

}
