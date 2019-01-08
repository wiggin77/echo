//
// Server
//
package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"strconv"
)

func main() {

	var port int
	var ssl bool
	var cert string
	var key string
	flag.IntVar(&port, "port", 5980, "port to listen on")
	flag.BoolVar(&ssl, "ssl", false, "true if ssl listener needed")
	flag.StringVar(&cert, "cert", "server.crt", "filespec for x509 server cert")
	flag.StringVar(&key, "key", "server.key", "filespec for x509 cert key")

	flag.Parse()

	var server net.Listener
	var err error

	fmt.Printf("Starting echo server on port %d;  ssl=%v, cert=%s, key=%s\n", port, ssl, cert, key)

	if ssl {
		cer, err := tls.LoadX509KeyPair(cert, key)
		if err != nil {
			panic("error loading cert: " + err.Error())
		}
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		server, err = tls.Listen("tcp", ":"+strconv.Itoa(port), config)
	} else {
		server, err = net.Listen("tcp", ":"+strconv.Itoa(port))
	}

	if server == nil || err != nil {
		panic("couldn't start listening: " + err.Error())
	}

	conns := clientConns(server)
	for {
		go handleConn(<-conns)
	}
}

func clientConns(listener net.Listener) chan net.Conn {
	ch := make(chan net.Conn)
	i := 0
	go func() {
		for {
			client, err := listener.Accept()
			if client == nil {
				fmt.Printf("couldn't accept: " + err.Error())
				continue
			}
			i++
			fmt.Printf("%d: %v <-> %v\n", i, client.LocalAddr(), client.RemoteAddr())
			ch <- client
		}
	}()
	return ch
}

func handleConn(client net.Conn) {
	b := bufio.NewReader(client)
	for {
		line, err := b.ReadBytes('\n')
		if err != nil { // EOF, or worse
			break
		}
		fmt.Println("echo: ", string(line))
		client.Write(line)
	}
}
