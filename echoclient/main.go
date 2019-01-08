//
// Client
//
package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	"net/textproto"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func main() {

	var host string
	var port int
	var ssl bool

	flag.StringVar(&host, "host", "0.0.0.0", "domain or IP of echo server")
	flag.IntVar(&port, "port", 5980, "port to listen on")
	flag.BoolVar(&ssl, "ssl", false, "true if ssl listener needed")

	flag.Parse()

	var conn net.Conn
	var err error

	fmt.Printf("Connecting to %s on port %d;  ssl=%v\n", host, port, ssl)

	if ssl {
		config := &tls.Config{InsecureSkipVerify: true}
		conn, err = tls.Dial("tcp", host+":"+strconv.Itoa(port), config)
	} else {
		conn, err = net.Dial("tcp", host+":"+strconv.Itoa(port))
	}

	if conn == nil || err != nil {
		fmt.Println("couldn't start client: ", err.Error())
		return
	}

	done := make(chan struct{}, 1)

	wg := &sync.WaitGroup{}
	wg.Add(2)

	go func() {
		send(conn, done)
		wg.Done()
	}()

	go func() {
		recv(conn, done)
		wg.Done()
	}()

	wg.Wait()
}

func send(conn net.Conn, done chan struct{}) {
	fmt.Print("\nType stuff. Empty line to exit.\n")
	for {
		reader := bufio.NewReader(os.Stdin)
		text, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Input error: ", err)
			break
		}
		if strings.TrimSpace(text) == "" {
			fmt.Println("Empty line, quitting.")
			break
		}
		_, err = conn.Write([]byte(text))
		if err != nil {
			fmt.Println("Write error: ", err)
			break
		}
	}
	done <- struct{}{}
}

func recv(conn net.Conn, done chan struct{}) {

	reader := bufio.NewReader(conn)
	tp := textproto.NewReader(reader)

	for {
		select {
		case <-done:
			return
		default:
		}
		conn.SetReadDeadline(time.Now().Add(time.Millisecond * 1000))
		line, err := tp.ReadLine()
		if err == nil {
			fmt.Println("Recv: ", line)
		}
	}
}
