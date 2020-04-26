package wrapper

import "errors"

// Ping is used to ping the Ironfist instance which is wrapping this application.
func Ping() error {
	b, err := createSocket(0x01, []byte{})
	if err != nil {
		return err
	}
	if string(b) != "Pong!" {
		return errors.New("invalid bytes returned")
	}
	return nil
}
