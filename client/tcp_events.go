package main

import (
	"net"

	"github.com/jakemakesstuff/structuredhttp"
)

func processApplicationRequest(body []byte, action string) (*structuredhttp.Response, error) {
	r, err := structuredhttp.POST(ConfigInitialised.Endpoint).
		Header("Ironfist-Action", action).
		Header("Ironfist-Version", "1.0.0").
		Header("Ironfist-Install-ID", InstallID).
		Bytes(body).
		Run()

	if err != nil {
		return nil, err
	}
	err = r.RaiseForStatus()
	if err != nil {
		return nil, err
	}
	return r, nil
}

func ping(conn net.Conn) {
	_, _ = conn.Write([]byte("Pong!"))
}

func updatePending(conn net.Conn) {
	r, err := processApplicationRequest([]byte{}, "Update-Pending")
	if err != nil {
		_, _ = conn.Write(append([]byte{0x02}, []byte(err.Error())...))
		return
	}
	j, err := r.JSON()
	if err != nil {
		return
	}
	if j.(bool) {
		_, _ = conn.Write([]byte{0x01})
	} else {
		_, _ = conn.Write([]byte{0x00})
	}
}

var tcpHandlers = map[byte]func(conn net.Conn){
	0x01: ping,
	0x02: updatePending,
}
