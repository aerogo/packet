package packet

import (
	"net"
)

// IOError is the data type for errors occurring in case of failure.
type IOError struct {
	Connection net.Conn
	Error      error
}
