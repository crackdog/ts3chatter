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
	var loggerFlag bool
	var logger *log.Logger

	flag.BoolVar(&loggerFlag, "log", false, "enable stdout logger")

	flag.Parse()

	if loggerFlag {
		logger = ts3sqlib.StdoutLogger
	}

	ts3, err := server.New("localhost", "testlogin", "bwu7tzVh", 1, logger, 5)
	if err != nil {
		//fmt.Fprintln(os.Stderr, err)
		log.Fatal(err)
		return
	}
	defer ts3.Quit()

	listener, err := net.Listen("tcp", "127.0.0.1:9001")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}

	log.Fatal(fcgi.Serve(listener, ts3))
}
