package packet

import (
	"encoding/binary"
	"fmt"
	"net"
	"sync/atomic"
)

// Stream represents a writable and readable network stream.
type Stream struct {
	connection  atomic.Value
	Incoming    chan *Packet
	Outgoing    chan *Packet
	closeWriter chan struct{}
	onError     func(IOError)
	verbose     bool
}

// NewStream creates a new stream with the given channel buffer size.
func NewStream(channelBufferSize int) *Stream {
	stream := &Stream{
		Incoming:    make(chan *Packet, channelBufferSize),
		Outgoing:    make(chan *Packet, channelBufferSize),
		closeWriter: make(chan struct{}),
		onError:     func(IOError) {},
	}

	return stream
}

// Connection returns the internal TCP/UDP connection object.
func (stream *Stream) Connection() net.Conn {
	return stream.connection.Load().(net.Conn)
}

// SetConnection sets the connection that the stream uses and
// it can be called multiple times on a single stream,
// effectively allowing you to hot-swap connections in failure cases.
func (stream *Stream) SetConnection(connection net.Conn) {
	if connection == nil {
		panic("connection is nil")
	}

	stream.connection.Store(connection)

	go stream.read(connection)
	go stream.write(connection)
}

// OnError sets the callback that should be called when IO errors occur.
func (stream *Stream) OnError(callback func(IOError)) {
	if callback == nil {
		panic("OnError using nil callback")
	}

	stream.onError = callback
}

// read starts a blocking routine that will read incoming messages.
// This function is meant to be called as a concurrent goroutine.
func (stream *Stream) read(connection net.Conn) {
	defer func() {
		stream.closeWriter <- struct{}{}
	}()

	if stream.verbose {
		fmt.Println("start read", connection.LocalAddr(), "->", connection.RemoteAddr())
		defer fmt.Println("end read", connection.LocalAddr(), "->", connection.RemoteAddr())
	}

	var length int64
	typeBuffer := make([]byte, 1)

	for {
		_, err := connection.Read(typeBuffer)

		if err != nil {
			stream.onError(IOError{connection, err})
			return
		}

		err = binary.Read(connection, binary.BigEndian, &length)

		if err != nil {
			stream.onError(IOError{connection, err})
			return
		}

		data := make([]byte, length)
		readLength := 0
		n := 0

		for readLength < len(data) {
			n, err = connection.Read(data[readLength:])
			readLength += n

			if err != nil {
				stream.onError(IOError{connection, err})
				return
			}
		}

		if err != nil {
			stream.onError(IOError{connection, err})
			return
		}

		stream.Incoming <- New(typeBuffer[0], data)
	}
}

// write starts a blocking routine that will write outgoing messages.
// This function is meant to be called as a concurrent goroutine.
func (stream *Stream) write(connection net.Conn) {
	if stream.verbose {
		fmt.Println("start write", connection.LocalAddr(), "->", connection.RemoteAddr())
		defer fmt.Println("end write", connection.LocalAddr(), "->", connection.RemoteAddr())
	}

	for {
		select {
		case <-stream.closeWriter:
			return

		case packet := <-stream.Outgoing:
			err := packet.Write(connection)

			if err != nil {
				stream.onError(IOError{connection, err})
				return
			}
		}
	}
}
