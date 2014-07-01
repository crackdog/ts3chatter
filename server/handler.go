package server

import (
	"fmt"
	"net/http"
	"strings"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//s.handlermutex.Lock()
	//defer s.handlermutex.Unlock()

	var i int

	if !s.closed {
		s.clmutex.Lock()
		clients := s.clientlist
		s.clmutex.Unlock()

		for _, client := range clients {
			nick, ok := client["client_nickname"]
			if ok {
				if strings.Contains(client["client_type"], "0") {
					fmt.Fprintln(w, "<p>", nick, "</p>")
				}
			} else {
				fmt.Fprintln(w, "error: ", client, ", empty map")
				if s.logger != nil {
					s.logger.Print("error: " + fmt.Sprint(client))
				}
			}
		}
		if i == 0 {
			fmt.Fprintln(w, "<h1>No one is online right now.</h1>")
		}
	} else {
		http.Error(w, "internal error", 500)
	}

	return
}
