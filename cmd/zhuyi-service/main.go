package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jpillora/ipfilter"
	"github.com/kraem/zhuyi-go/server"
)

func main() {

	s := server.NewServer()

	// TODO read these as env vars
	var fw server.IpFilter
	fw.IPFilter = ipfilter.New(ipfilter.Options{
		AllowedIPs:     []string{"127.0.0.1", "10.0.0.0/24"},
		BlockByDefault: true,
	})

	r := mux.NewRouter()
	r.Handle("/status", fw.FilterMiddleware(server.StatusHandler(s))).Methods("GET")
	r.Handle("/d3/graph", fw.FilterMiddleware(server.GraphHandler(s))).Methods("GET")
	r.Handle("/unlinked", fw.FilterMiddleware(server.UnlinkedHandler(s))).Methods("GET")
	r.Handle("/zettel/add", fw.FilterMiddleware(server.AddZettelHandler(s))).Methods("POST")
	// TODO Handle options like this for all endpoints
	r.Handle("/zettel/add", fw.FilterMiddleware(server.AddZettelHandlerOptions(s))).Methods("OPTIONS")

	log.Fatal(http.ListenAndServe(s.Cfg.Addr, r))

}
