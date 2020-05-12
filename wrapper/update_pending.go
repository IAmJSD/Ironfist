package wrapper

import "errors"

// UpdatePending allows the user to check if an update is pending.
func UpdatePending() (bool, error) {
	b, err := createSocket(0x02, []byte{})
	if err != nil {
		return false, err
	}
	if b[0] == 0x02 {
		return false, errors.New(string(b[1:]))
	}
	return b[0] == 0x01, nil
}
