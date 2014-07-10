package server

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const (
	htmlHeader = "<!DOCTYPE html><html><head><title>fkarchery ts3</title></head><body bgcolor=\"000000\"><font face=\"Monospace\" color=\"00FF00\">\r\n"
	htmlFooter = "</body></html>\r\n"
)

func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//s.handlermutex.Lock()
	//defer s.handlermutex.Unlock()

	if r.URL.Path == "/ts3chatter/channellist" {
		/*answer, err := s.ts3conn.Send("channellist\n")
		if err != nil {
			fmt.Fprintln(w, "Error: ", err)
			return
		}

		channelmaps, err := ts3sqlib.MsgToMaps(answer)
		if err != nil {
			fmt.Fprintln(w, "Error: ", err)
			return
		}

		m, err := json.MarshalIndent(channelmaps, "", "  ")
		if err != nil {
			fmt.Fprintln(w, "Error: ", err)
			return
		}

		fmt.Fprintln(w, string(m))*/
		s.ServerChannellistJSON(w, r)
	} else if r.URL.Path == "/ts3chatter/clientlist" {
		s.ServerClientlistJSON(w, r)
	} else {
		if !s.closed {
			s.datamutex.Lock()
			clients := s.data.clientlist
			n := s.data.n
			s.datamutex.Unlock()

			fmt.Fprint(w, htmlHeader)

			//fmt.Fprintln(w, r.URL.Path, "\n")

			if n > 0 {
				if n == 1 {
					fmt.Fprintln(w, "<h2>1 Client is online:</h2><ol>")
				} else {
					fmt.Fprintln(w, "<h2>", n, " Clients are online:</h2><ol>")
				}

				/*for _, client := range clients {
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
				}*/
				for _, client := range clients {
					if client.ClientType == 0 {
						fmt.Fprintln(w, "<li>", client.ClientNickname, "</li>")
					}
				}

				fmt.Fprintln(w, "</ol>")
			} else {
				fmt.Fprintln(w, "<h1>No one is online right now.</h1>")
			}

			fmt.Fprint(w, htmlFooter)
		} else {
			http.Error(w, "internal error", 500)
		}
	}

	return
}

func (s *Server) ServerClientlistJSON(w http.ResponseWriter, r *http.Request) {
	s.datamutex.Lock()
	m, err := json.MarshalIndent(s.data.clientlist, "", "  ")
	s.datamutex.Unlock()
	if err != nil {
		fmt.Fprintln(w, "Error: ", err)
	} else {
		fmt.Fprintln(w, string(m))
	}
}

func (s *Server) ServerChannellistJSON(w http.ResponseWriter, r *http.Request) {
	s.datamutex.Lock()
	m, err := json.MarshalIndent(s.data.channellist, "", "  ")
	s.datamutex.Unlock()
	if err != nil {
		fmt.Fprintln(w, "Error: ", err)
	} else {
		fmt.Fprintln(w, string(m))
	}
}
