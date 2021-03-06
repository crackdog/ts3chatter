//Package server contains the ts3chatter server.
package server

import (
	"fmt"
	"github.com/crackdog/ts3sqlib"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"
)

//Server contains the ts3 sq connection and other ts3 server data.
type Server struct {
	ts3conn       *ts3sqlib.SqConn
	servemux      *http.ServeMux
	data          *serverData
	address       string
	loginname     string
	disconnected  bool
	password      string
	nickname      string
	path          string
	virtualserver int
	logger        *log.Logger
	sleepseconds  int
	handlermutex  *sync.Mutex
	datamutex     *sync.Mutex
	quit          chan bool
	closed        bool
}

type serverData struct {
	clientlist  ts3sqlib.Clientlist
	n           int //Number of online clients.
	channellist ts3sqlib.Channellist
}

//New creates a new Server structure.
func New(address, login, password string, virtualserver int,
	logger *log.Logger, sleepseconds int, nick string, path string) (s *Server, err error) {

	s = new(Server)
	s.address = address
	s.loginname = login
	s.password = password
	s.nickname = nick
	s.path = path
	s.virtualserver = virtualserver
	s.logger = logger
	s.data = new(serverData)
	s.data.clientlist = make([]ts3sqlib.Client, 0)
	s.data.channellist = make([]ts3sqlib.Channel, 0)
	s.data.n = 0
	s.sleepseconds = sleepseconds
	s.handlermutex = new(sync.Mutex)
	s.datamutex = new(sync.Mutex)
	s.servemux = s.NewServeMux()

	s.disconnected = true

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

	s.disconnected = false
	s.log("disconnected = false")

	return
}

//Quit disconnects the ts3 sq connection of a Server.
func (s *Server) Quit() (err error) {
	s.closed = true
	s.quit <- true
	if s.ts3conn != nil {
		err = s.ts3conn.Quit()
	} else {
		err = fmt.Errorf("quit: s.ts3conn nil error")
	}
	return
}

func (s *Server) log(v ...interface{}) {
	if s.logger != nil {
		s.logger.Print(v...)
	}
}

func (s *Server) reconnect() (err error) {
	if !s.ts3conn.IsClosed() {
		_ = s.ts3conn.Quit()
	}

	for {

		s.ts3conn, err = ts3sqlib.Dial(s.address, s.logger)
		if err == nil {
			err = s.login()
		}

		if err != nil {
			time.Sleep(2 * time.Second)
		} else {
			break
		}
	}

	if err != nil {
		s.closed = false
		s.disconnected = false
		s.log("disconnected = false")
	}

	return
}

func (s *Server) handleError(err error) {
	switch {
	case ts3sqlib.ClosedError.Equals(err):
		s.disconnected = true
		s.log("connection error, trying to reconnect")
		err = s.reconnect()
	case ts3sqlib.PermissionError.Equals(err):
		err = s.login()
	default:
		//nop
	}

	if err != nil {
		s.log("handleError: ", err)
	}
}

func (s *Server) dataReceiver(sleeptime time.Duration) {
	var (
		data *serverData
		err  error
	)

	err = s.login()
	if err != nil {
		s.handleError(err)
	}

	for !s.closed {
		err = nil
		data = new(serverData)

		data.clientlist, err = s.ts3conn.ClientlistToClients("")
		if err != nil {
			s.handleError(err)
			continue
		}

		data.clientlist, err = s.ts3conn.GetConnectionTimeForCL(data.clientlist)
		if err != nil {
			s.handleError(err)
			continue
		}

		channelmaps, err := s.ts3conn.SendToMaps("channellist\n")
		if err != nil {
			s.handleError(err)
			continue
		}

		data.channellist = make([]ts3sqlib.Channel, len(channelmaps))
		for i := range data.channellist {
			data.channellist[i].Data = channelmaps[i]
			data.channellist[i].Name = data.channellist[i].Data["channel_name"]
			data.channellist[i].Clients = make([]ts3sqlib.Client, 0, 2)
		}

		for i := range data.channellist {
			channelIndex, tmperr := strconv.Atoi(data.channellist[i].Data["cid"])
			if tmperr != nil {
				continue
			}
			for _, c := range data.clientlist {
				if c.Cid == channelIndex {
					data.channellist[i].Clients = append(data.channellist[i].Clients, c)
				}
			}
		}

		s.datamutex.Lock()
		s.data = data
		s.datamutex.Unlock()

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
