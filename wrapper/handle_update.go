package wrapper

// HandleUpdate is used to handle the update.
func HandleUpdate() error {
	_, err := createSocket(0x05, []byte{})
	if err != nil {
		return err
	}
	return nil
}
