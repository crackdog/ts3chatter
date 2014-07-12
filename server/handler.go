package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.Contains(r.URL.Path, "/ts3chatter/clientlist"):
		s.datamutex.Lock()
		s.ServeJSON(w, r, s.data.clientlist)
		s.datamutex.Unlock()
	case strings.Contains(r.URL.Path, "/ts3chatter/channellist"):
		s.datamutex.Lock()
		s.ServeJSON(w, r, s.data.channellist)
		s.datamutex.Unlock()
	default:
		http.Error(w, "404 not found", 404)
	}
	return
}

func (s *Server) ServeJSON(w http.ResponseWriter, r *http.Request, v interface{}) {
	m, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, "internal server error", 500)
	} else {
		fmt.Fprintln(w, string(m))
	}
	return
}
