package packet

import (
	"net"
	"sync/atomic"
	"time"
)

// Stream represents a writable and readable network stream.
type Stream struct {
	connection atomic.Value
	Incoming   chan *Packet
	Outgoing   chan *Packet
	onError    func(IOError)
	close      chan bool
	closed     atomic.Value
}

// IOError is the data type for errors occuring in case of failure.
type IOError struct {
	Connection net.Conn
	Error      error
}

// NewStream creates a new stream with the given channelBufferSize.
func NewStream(channelBufferSize int) *Stream {
	stream := &Stream{
		Incoming: make(chan *Packet, channelBufferSize),
		Outgoing: make(chan *Packet, channelBufferSize),
		close:    make(chan bool),
		onError:  func(IOError) {},
	}

	stream.closed.Store(false)

	go stream.Write()

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
	stream.connection.Store(connection)
	go stream.Read(connection)
}

// OnError sets the callback that should be called when IO errors occur.
func (stream *Stream) OnError(callback func(IOError)) {
	if callback == nil {
		panic("OnError using nil callback")
	}

	stream.onError = callback
}

// Close closes the stream.
func (stream *Stream) Close() {
	stream.closed.Store(true)
	stream.close <- true
	<-stream.close
	close(stream.Incoming)
}

// Read starts a blocking routine that will read incoming messages.
// This function is meant to be called as a concurrent goroutine.
func (stream *Stream) Read(connection net.Conn) {
	typeBuffer := make([]byte, 1)
	lengthBuffer := make([]byte, 8)

	for {
		if stream.closed.Load().(bool) {
			return
		}

		_, err := connection.Read(typeBuffer)

		if err != nil {
			stream.onError(IOError{connection, err})
			return
		}

		_, err = connection.Read(lengthBuffer)

		if err != nil {
			stream.onError(IOError{connection, err})
			return
		}

		length, err := Int64FromBytes(lengthBuffer)

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

		stream.Incoming <- New(typeBuffer[0], data)
	}
}

// Write starts a blocking routine that will write outgoing messages.
// This function is meant to be called as a concurrent goroutine.
func (stream *Stream) Write() {
	for {
		select {
		case packet := <-stream.Outgoing:
			msg := packet.Bytes()

		retry:
			if stream.closed.Load().(bool) {
				break
			}

			connection := stream.Connection()
			totalWritten := 0

			for totalWritten < len(msg) {
				writtenThisCall, err := connection.Write(msg[totalWritten:])

				if err != nil {
					stream.onError(IOError{connection, err})
					time.Sleep(1 * time.Millisecond)
					goto retry
				}

				totalWritten += writtenThisCall
			}

		case <-stream.close:
			connection := stream.Connection()
			err := connection.Close()

			if err != nil {
				stream.onError(IOError{connection, err})
			}

			close(stream.close)
			return
		}
	}
}
