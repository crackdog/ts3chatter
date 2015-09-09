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

	sm.HandleFunc("/ts3chatter/clientlist", s.ServeClientlist)

	sm.HandleFunc("/ts3chatter/channellist", s.ServeChannellist)

	sm.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})

	return sm
}

func (s *Server) ServeChannellist(w http.ResponseWriter, r *http.Request) {
	s.datamutex.Lock()
	cl := s.data.channellist
	s.datamutex.Unlock()

	s.serveJSON(w, r, cl)
}

func (s *Server) ServeClientlist(w http.ResponseWriter, r *http.Request) {
	s.datamutex.Lock()
	cl := s.data.clientlist
	s.datamutex.Unlock()

	s.serveJSON(w, r, cl)
}

func (s *Server) serveJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	//m, err := json.MarshalIndent(v, "", "  ")
	if !s.disconnected {
		m, err := json.Marshal(v)
		if err != nil {
			http.Error(w, "internal server error", 500)
		} else {
			fmt.Fprintln(w, string(m))
		}
	} else {
		http.Error(w, "server offline", 500)
	}
	return
}
