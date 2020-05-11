package main

import "net"

func ping(conn net.Conn) {
	_, _ = conn.Write([]byte("Pong!"))
}

var tcpHandlers = map[byte]func(conn net.Conn){
	0x01: ping,
}
