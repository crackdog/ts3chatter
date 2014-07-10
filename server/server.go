//server contains the ts3chatter server.
package server

import (
	"fmt"
	"github.com/crackdog/ts3sqlib"
	"log"
	"sync"
	"time"
)

const (
	tryN = 3
)

type Server struct {
	ts3conn       *ts3sqlib.SqConn
	data          *serverData
	address       string
	loginname     string
	password      string
	nickname      string
	virtualserver int
	logger        *log.Logger
	sleepseconds  int
	handlermutex  *sync.Mutex
	datamutex     *sync.Mutex
	quit          chan bool
	closed        bool
}

type serverData struct {
	clientlist  []ts3sqlib.Client
	n           int //Number of online clients.
	channellist []channel
}

type channel struct {
	name    string            `json:"channel_name"`
	data    map[string]string `json:"-"`
	clients []ts3sqlib.Client `json:"clients"`
}

func New(address, login, password string, virtualserver int,
	logger *log.Logger, sleepseconds int, nick string) (s *Server, err error) {

	s = new(Server)
	s.address = address
	s.loginname = login
	s.password = password
	s.nickname = nick
	s.virtualserver = virtualserver
	s.logger = logger
	//s.clientlist = new(clients)
	//s.clientlist.cl = make([]ts3sqlib.Client, 0)
	//s.clientlist.n = 0
	s.data = new(serverData)
	s.data.clientlist = make([]ts3sqlib.Client, 0)
	s.data.channellist = make([]channel, 0)
	s.data.n = 0
	s.sleepseconds = sleepseconds
	s.handlermutex = new(sync.Mutex)
	s.datamutex = new(sync.Mutex)

	s.quit = make(chan bool)
	s.closed = false

	s.ts3conn, err = ts3sqlib.Dial(address, logger)
	if err != nil {
		s = nil
		return
	}

	go s.dataReceiver(time.Duration(s.sleepseconds) * time.Second)

	return
}

func (s *Server) login() (err error) {
	if s.ts3conn == nil {
		err = fmt.Errorf("login: nil pointer")
		return
	}

	err = s.ts3conn.Use(s.virtualserver)
	if err != nil {
		return
	}

	err = s.ts3conn.Login(s.loginname, s.password)
	if err != nil {
		return
	}

	//changing nickname...
	pairs, err := s.ts3conn.SendToMap("whoami\n")
	if err != nil {
		return
	}

	clid, ok := pairs["client_id"]
	if !ok {
		err = fmt.Errorf("error at collecting client_id")
		return
	}

	_, err = s.ts3conn.Send("clientupdate clid=" + clid + " client_nickname=" +
		s.nickname + "\n")

	return
}

func (s *Server) Quit() (err error) {
	s.closed = true
	s.quit <- true
	if s.ts3conn != nil {
		err = s.ts3conn.Quit()
	} else {
		err = fmt.Errorf("Quit: s.ts3conn nil error")
	}
	return
}

func (s *Server) log(v ...interface{}) {
	if s.logger != nil {
		s.logger.Print(v...)
	}
}

func (s *Server) handleError(err error) {
	switch {
	case ts3sqlib.ClosedError.Equals(err):
		_ = s.Quit()
		s.closed = false
		s.ts3conn, err = ts3sqlib.Dial(s.address, s.logger)
		go s.dataReceiver(time.Duration(s.sleepseconds) * time.Second)
	case ts3sqlib.PermissionError.Equals(err):
		err = s.login()
	default:
		//nop
	}

	if err != nil {
		s.log("handleError: ", err)
	}
}

//clientlistReceiver receives a Clientlist every
func (s *Server) dataReceiver(sleeptime time.Duration) {
	var (
		data *serverData
		err  error
	)

	err = s.login()
	if err != nil {
		s.log(err)
		s.Quit()
		return
	}

	for !s.closed {
		//for n := 0; !s.closed && n < tryN; i++ {
		err = nil
		data = new(serverData)

		data.clientlist, err = s.ts3conn.ClientlistToClients("")
		if err != nil {
			s.handleError(err)
			continue
		}

		channelmaps, err := s.ts3conn.SendToMaps("channellist\n")
		if err != nil {
			s.handleError(err)
			continue
		}

		data.channellist = make([]channel, len(channelmaps))
		for i, c := range data.channellist {
			c.data = channelmaps[i]
			c.name = c.data["channel_name"]
			c.clients = make([]ts3sqlib.Client, 0, 5) //maybe more or less than 5
		}

		for _, c := range data.clientlist {
			if c.Cid >= 0 && c.Cid < len(data.channellist) { //maybe c.Cid -> uint
				data.channellist[c.Cid].clients = append(data.channellist[c.Cid].clients, c)
			}
		}

		s.datamutex.Lock()
		s.data = data
		s.datamutex.Unlock()

		/*s.login()
		if err != nil {
			s.log(err)
		}

		for !s.closed {
			clientlist = new(clients)
			clientlist.cl, err = s.ts3conn.ClientlistToClients("") //Maps("")
			if err != nil {
				s.log(err)
				if ts3sqlib.PermissionError.Equals(err) {
					s.login()
					if err != nil {
						s.log(err)
					}
				}
			} else {
				clientlist.n = 0
				for _, c := range clientlist.cl {
					if c.ClientType == 0 {
						clientlist.n++
					}
				}

				s.clmutex.Lock()
				s.clientlist = clientlist
				s.clmutex.Unlock()
			}
		}*/

		time.Sleep(sleeptime)
	}
}

func (s *Server) notificationHandler() {
	var (
		answer string
		err    error
	)

	for {
		answer, err = s.ts3conn.RecvNotify()
		if err != nil {
			s.log(err)
		} else {
			//handle notification
			s.log(answer)
		}
	}
}
