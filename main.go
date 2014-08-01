package main

import (
	"flag"
	"fmt"
	"github.com/crackdog/ts3chatter/server"
	"github.com/crackdog/ts3sqlib"
	"log"
	"net"
	"net/http/fcgi"
	"os"
)

func main() {
	var (
		loggerFlag bool
		fcgiport   string
		address    string
		logger     *log.Logger
		lname      string
		lpw        string
		nick       string
	)

	flag.BoolVar(&loggerFlag, "log", false, "enable stdout logger")
	flag.StringVar(&fcgiport, "fcgiport", "9001", "change fcgi port")
	flag.StringVar(&address, "ts3addr", "localhost",
		"change the ts3 server query 'address:port'")
	flag.StringVar(&lname, "login", "ts3chatter", "set ts3 server query login name")
	flag.StringVar(&lpw, "pw", "********", "set ts3 server query password")
	flag.StringVar(&nick, "nick", "ts3chatter", "set the nickname for the server query")

	flag.Parse()

	if loggerFlag {
		logger = ts3sqlib.StdoutLogger
	}

	ts3, err := server.New(address, lname, lpw, 1, logger, 5, nick)
	if err != nil {
		log.Fatal(err)
		return
	}
	defer ts3.Quit()

	listener, err := net.Listen("tcp", "127.0.0.1:"+fcgiport)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	log.Fatal(fcgi.Serve(listener, ts3))
}
