package wrapper

import "errors"

// JoinUpdateChannel is used to make Ironfist join a update channel.
func JoinUpdateChannel(ChannelName string) error {
	b, err := createSocket(0x03, []byte(ChannelName))
	if err != nil {
		return err
	}
	if b[0] == 0x02 {
		return errors.New(string(b[1:]))
	}
	return nil
}

// LeaveUpdateChannel is used to make Ironfist leave a update channel.
func LeaveUpdateChannel(ChannelName string) error {
	b, err := createSocket(0x04, []byte(ChannelName))
	if err != nil {
		return err
	}
	if b[0] == 0x02 {
		return errors.New(string(b[1:]))
	}
	return nil
}
