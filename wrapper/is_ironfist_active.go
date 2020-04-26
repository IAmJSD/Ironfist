package wrapper

import "os"

// IsActive is used to check if Ironfist is active.
func IsActive() bool {
	return os.Getenv("IRONFIST_HOSTNAME") != "" && os.Getenv("IRONFIST_KEY") != ""
}
