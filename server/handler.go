package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//ServeHTTP serves a given http.Request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.servemux.ServeHTTP(w, r)
}

func (s *Server) NewServeMux() *http.ServeMux {
	sm := http.NewServeMux()

	//sm.HandleFunc("/ts3chatter/clientlist",

	//sm.HandleFunc("/ts3chatter/channellist",

	sm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "404 not found", 404)
	})

	return sm
}

func (s *Server) GetChannellist() interface{} {
	s.datamutex.Lock()
	defer s.datamutex.Unlock()
	return s.data.channellist
}

func (s *Server) GetClientlist() interface{} {
	s.datamutex.Lock()
	defer s.datamutex.Unlock()
	return s.data.clientlist
}

func (s *Server) serveJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	//m, err := json.MarshalIndent(v, "", "  ")
	m, err := json.Marshal(v)
	if err != nil {
		http.Error(w, "internal server error", 500)
	} else {
		fmt.Fprintln(w, string(m))
	}
	return
}
