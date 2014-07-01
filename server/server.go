//server contains the ts3chatter server.
package server

import (
	"fmt"
	"github.com/crackdog/ts3sqlib"
	"log"
	"strings"
	"sync"
	"time"
)

type Server struct {
	ts3conn       *ts3sqlib.SqConn
	clientlist    *clients
	loginname     string
	password      string
	virtualserver int
	logger        *log.Logger
	handlermutex  *sync.Mutex
	clmutex       *sync.Mutex
	quit          chan bool
	closed        bool
}

type clients struct {
	cl []map[string]string
	n  int
}

func New(address, login, password string, virtualserver int,
	logger *log.Logger, sleepseconds int) (s *Server, err error) {

	s = new(Server)
	s.loginname = login
	s.password = password
	s.virtualserver = virtualserver
	s.logger = logger
	s.clientlist = new(clients)
	s.clientlist.cl = make([]map[string]string, 0)
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
		clientlist.cl, err = s.ts3conn.ClientlistToMaps("")
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
				if strings.Contains(c["client_type"], "0") {
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
