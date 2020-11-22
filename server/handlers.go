package server

import (
	"encoding/json"
	"net"
	"net/http"

	"github.com/jpillora/ipfilter"
	"github.com/kraem/zhuyi-go/pkg/log"
	"github.com/kraem/zhuyi-go/pkg/payloads"
)

type StatusResponse struct {
	Payload struct {
		Status string `json:"status"`
	} `json:"payload"`
}

type IpFilter struct {
	*ipfilter.IPFilter
}

func (f *IpFilter) FilterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, _ := net.SplitHostPort(r.RemoteAddr)
		if !f.Allowed(ip) {
			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func AddZettelHandlerOptions(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
	})
}

func AddZettelHandler(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}

		w.Header().Set("Content-Type", "application/json")
		var resp payloads.AppendResponse

		var payloadIncoming payloads.AppendRequest
		err := json.NewDecoder(r.Body).Decode(&payloadIncoming)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errString := err.Error()
			resp.Error = &errString
			json.NewEncoder(w).Encode(resp)
			//fmt.Printf("ERROR - Unable to unmarshal append-request. Please report issue @ GH\nerr: %s\n", err)
			log.LogError(err)
			return
		}

		zettelFileName, err := s.CfgZettel.CreateZettel(payloadIncoming.Payload.Title, payloadIncoming.Payload.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errString := err.Error()
			resp.Error = &errString
			json.NewEncoder(w).Encode(resp)
			log.LogError(err)
			return
		}

		resp.Payload.FileName = &zettelFileName

		json.NewEncoder(w).Encode(resp)
	})
}

func UnlinkedHandler(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}

		var resp payloads.UnlinkedResponse

		zs, err := s.CfgZettel.UnlinkedNodes()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errString := err.Error()
			resp.Error = &errString
			json.NewEncoder(w).Encode(resp)
			log.LogError(err)
			return
		}

		resp.Payload.Zettels = zs

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func GraphHandler(s *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}

		var resp payloads.GraphResponse

		g, err := s.CfgZettel.CreateD3jsGraph()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			errString := err.Error()
			resp.Error = &errString
			json.NewEncoder(w).Encode(resp)
			log.LogError(err)
			return
		}

		resp.Payload.Graph = g

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func StatusHandler(a *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		var resp StatusResponse

		resp.Payload.Status = "API is up and running"

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	})
}

func NotImplemented(a *Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		setupResponse(&w, r)
		if (*r).Method == "OPTIONS" {
			return
		}
		w.Write([]byte("Not Implemented"))
	})
}

func setupResponse(w *http.ResponseWriter, req *http.Request) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}
