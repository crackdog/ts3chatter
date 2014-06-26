package server

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//s.handlermutex.Lock()
	//defer s.handlermutex.Unlock()

	if !s.closed {
		s.clmutex.Lock()
		clients := s.clientlist
		s.clmutex.Unlock()

		for i := range clients {
			nick, ok := clients[i]["client_nickname"]
			if ok {
				if strings.Contains(clients[i]["client_type"], "0") {
					fmt.Fprintln(w, "<p>", nick, "</p>")
				}
			} else {
				fmt.Fprintln(w, "error: ", clients[i], ", empty map")
				if s.logger != nil {
					s.logger.Print("error: " + fmt.Sprint(clients[i]))
				}
			}
		}
	} else {
		http.Error(w, "internal error", 500)
	}

	return
}
