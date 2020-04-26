package main

import "net"

func ping(conn net.Conn) {
	_, _ = conn.Write([]byte("Pong!"))
}

var tcpHandlers = map[byte]func(conn net.Conn){
	0x01: ping,
}

// StartTCPServer is used to start a TCP server.
func StartTCPServer() {
	// Create the socket and get the hostname.
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		panic(err)
	}
	Hostname = l.Addr().String()

	// The main socket loop.
	go func() {
		for {
			conn, err := l.Accept()
			if err != nil {
				continue
			}
			go func(c net.Conn) {
				// Defer killing the socket.
				defer func() { _ = c.Close() }()

				// Get the first byte.
				onebyte := make([]byte, 1)
				n, err := c.Read(onebyte)
				if err != nil || n == 0 {
					return
				}

				// Check if the first byte is supported.
				handler, ok := tcpHandlers[onebyte[0]]
				if !ok {
					return
				}

				// Read ahead for the key (this is the authentication).
				KeyBytes := make([]byte, len(ApplicationKey))
				n, err = c.Read(KeyBytes)
				if err != nil || n != len(ApplicationKey) {
					return
				}

				// Check if the key bytes are the same as the token.
				if string(KeyBytes) != ApplicationKey {
					return
				}

				// Call the handler.
				handler(c)
			}(conn)
		}
	}()
}
