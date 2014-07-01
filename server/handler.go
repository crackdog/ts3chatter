package server

import (
	"fmt"
	"net/http"
	"strings"
)

const (
	htmlHeader = "<html><head><title>fkarchery ts3</title></head><body>\r\n"
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
			fmt.Fprint(w, "<h6>", n, " Clients are online:<h6>\r\n")

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
				i++
			}
		} else {
			fmt.Fprintln(w, "<h1>No one is online right now.</h1>")
		}

		fmt.Fprint(w, htmlFooter)
	} else {
		http.Error(w, "internal error", 500)
	}

	return
}
