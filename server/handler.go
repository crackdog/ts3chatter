package server

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	htmlHeader = "<!DOCTYPE html><html><head><title>fkarchery ts3</title></head><body bgcolor=\"000000\"><font face=\"Monospace\" color=\"00FF00\">\r\n"
	htmlFooter = "</body></html>\r\n"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//s.handlermutex.Lock()
	//defer s.handlermutex.Unlock()

	var i int

	if !s.closed {
		s.clmutex.Lock()
		clients := s.clientlist.cl
		n := s.clientlist.n
		s.clmutex.Unlock()

		fmt.Fprint(w, htmlHeader)
		if n > 0 {
			fmt.Fprintln(w, "<h2>", n, " Clients are online:</h2><ol>")

			for _, client := range clients {
				nick, ok := client["client_nickname"]
				if ok {
					if strings.Contains(client["client_type"], "0") {
						fmt.Fprintln(w, "<li>", nick, "</li>")
					}
				} else {
					fmt.Fprintln(w, "error: ", client, ", empty map")
					if s.logger != nil {
						s.logger.Print("error: " + fmt.Sprint(client))
					}
				}
				i++
			}
			fmt.Fprintln(w, "</ol>")
		} else {
			fmt.Fprintln(w, "<h1>No one is online right now.</h1>")
		}

		fmt.Fprint(w, htmlFooter)
	} else {
		http.Error(w, "internal error", 500)
	}

	return
}
