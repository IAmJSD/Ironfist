package wrapper

import (
	"errors"
	"net"
	"os"
)

func packPacket(PacketID byte, Data []byte) []byte {
	// Get the key.
	Key := []byte(os.Getenv("IRONFIST_KEY"))

	// Get the key length.
	KeyLen := len(Key)

	// Create the packet.
	Packet := make([]byte, 1+KeyLen+len(Data))

	// Set the packet ID.
	Packet[0] = PacketID

	// Insert the key.
	for i, v := range Key {
		Packet[1+i] = v
	}

	// Insert the rest of the data.
	for i, v := range Data {
		Packet[1+KeyLen+i] = v
	}

	// Return the packed packet.
	return Packet
}

func createSocket(PacketID byte, Data []byte) ([]byte, error) {
	// Check Ironfist is active first.
	if !IsActive() {
		return nil, errors.New("ironfist is not active")
	}

	// Create the packet.
	Packet := packPacket(PacketID, Data)

	// Create the socket.
	s, err := net.Dial("tcp", os.Getenv("IRONFIST_HOSTNAME"))
	if err != nil {
		return nil, err
	}

	// Send the packet down the socket.
	_, err = s.Write(Packet)
	if err != nil {
		return nil, err
	}

	// Try reading the response.
	MaxPacket := make([]byte, 1000000)
	n, err := s.Read(MaxPacket)
	if err != nil {
		return nil, err
	}
	realloc := make([]byte, n)
	for i, v := range MaxPacket {
		if i == n {
			break
		}
		realloc[i] = v
	}

	// Return the response.
	return realloc, nil
}
