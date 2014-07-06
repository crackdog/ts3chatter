//server contains the ts3chatter server.
package server

import (
	"fmt"
	"github.com/crackdog/ts3sqlib"
	"log"
	"sync"
	"time"
)

type Server struct {
	ts3conn       *ts3sqlib.SqConn
	clientlist    *clients
	loginname     string
	password      string
	nickname      string
	virtualserver int
	logger        *log.Logger
	handlermutex  *sync.Mutex
	clmutex       *sync.Mutex
	quit          chan bool
	closed        bool
}

type clients struct {
	cl []ts3sqlib.Client
	n  int
}

func New(address, login, password string, virtualserver int,
	logger *log.Logger, sleepseconds int, nick string) (s *Server, err error) {

	s = new(Server)
	s.loginname = login
	s.password = password
	s.nickname = nick
	s.virtualserver = virtualserver
	s.logger = logger
	s.clientlist = new(clients)
	s.clientlist.cl = make([]ts3sqlib.Client, 0)
	s.clientlist.n = 0
	s.handlermutex = new(sync.Mutex)
	s.clmutex = new(sync.Mutex)

	s.quit = make(chan bool)
	s.closed = false

	s.ts3conn, err = ts3sqlib.Dial(address, logger)
	if err != nil {
		s = nil
		return
	}

	go s.clientlistReceiver(time.Duration(sleepseconds) * time.Second)

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

//clientlistReceiver receives a Clientlist every
func (s *Server) clientlistReceiver(sleeptime time.Duration) {
	var (
		clientlist *clients
		err        error
	)

	s.login()
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
