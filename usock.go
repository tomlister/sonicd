package main

import (
	"os"
	"bytes"
	"net"
	"io"
	"sync"
)

func usock_handler(client net.Conn, c chan string) {
	buf := new(bytes.Buffer)
	io.Copy(buf, client)
	command := buf.String()
	if command == "show\n" {
		c <- "show"
	} else if command == "hide\n" {
		c <- "hide"
	} else if command == "up\n" {
		c <- "up"
	} else if command == "down\n" {
		c <- "down"
	}
}

func usock_server(command_queue *[]string, command_queue_mutex *sync.Mutex) {
	sock_address := "/tmp/sonic.sock"
	err := os.RemoveAll(sock_address)
	e(err)

	listen, err := net.Listen("unix", sock_address)
	e(err)

	defer listen.Close()

	for {
		conn, err := listen.Accept()
		e(err)
		command_chan := make(chan string)
		go usock_handler(conn, command_chan)
		command := <- command_chan
		command_queue_mutex.Lock()
		*command_queue = append(*command_queue, command)
		command_queue_mutex.Unlock()
	}
}