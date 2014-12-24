package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

//ServeHTTP serves a given http.Request.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasSuffix(r.URL.Path, s.path+"clientlist"):
		s.datamutex.Lock()
		s.serveJSON(w, r, s.data.clientlist)
		s.datamutex.Unlock()
	case strings.HasSuffix(r.URL.Path, s.path+"channellist"):
		s.datamutex.Lock()
		s.serveJSON(w, r, s.data.channellist)
		s.datamutex.Unlock()
	default:
		http.Error(w, "404 not found", 404)
	}
	return
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
