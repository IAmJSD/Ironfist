package main

import (
	"encoding/json"
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

func toggleUpdateChannel(join bool) func(conn net.Conn) {
	return func(conn net.Conn) {
		b := make([]byte, 250)
		n, err := conn.Read(b)
		if err != nil {
			_, _ = conn.Write(append([]byte{0x02}, []byte(err.Error())...))
			return
		}
		channel := string(b[:n])
		b, err = json.Marshal(&channel)
		if err != nil {
			_, _ = conn.Write(append([]byte{0x02}, []byte(err.Error())...))
			return
		}
		req := "Leave-Update-Channel"
		if join {
			req = "Join-Update-Channel"
		}
		_, err = processApplicationRequest(b, req)
		if err != nil {
			_, _ = conn.Write(append([]byte{0x02}, []byte(err.Error())...))
			return
		}
		_, _ = conn.Write([]byte{0x01})
	}
}

func handleUpdate(conn net.Conn) {
	defer func() { _, _ = conn.Write([]byte{0x00}) }()

	r, err := structuredhttp.POST(ConfigInitialised.Endpoint).
		Header("Ironfist-Action", "Get-Latest-Update").
		Header("Ironfist-Version", "1.0.0").
		Header("Ironfist-Install-ID", InstallID).
		Run()
	if err != nil {
		return
	}

	j, err := r.JSON()
	if err != nil || j == nil {
		return
	}

	UpDowngradeRelease(j.(map[string]interface{})["hash"].(string))
}

var tcpHandlers = map[byte]func(conn net.Conn){
	0x01: ping,
	0x02: updatePending,
	0x03: toggleUpdateChannel(true),
	0x04: toggleUpdateChannel(false),
	0x05: handleUpdate,
}
