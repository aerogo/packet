package packet_test

import (
	"errors"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/aerogo/packet"
)

// connectionWithReadError errors the Read call after `errorOnReadNumber` tries.
type connectionWithReadError struct {
	countReads        int
	errorOnReadNumber int
	inner             net.Conn
}

func (conn *connectionWithReadError) Read(buffer []byte) (int, error) {
	conn.countReads++

	if conn.countReads == conn.errorOnReadNumber {
		return 0, errors.New("Artificial error")
	}

	return conn.inner.Read(buffer)
}

func (conn *connectionWithReadError) Write(buffer []byte) (n int, err error) {
	return conn.inner.Write(buffer)
}

func (conn *connectionWithReadError) Close() error {
	return conn.inner.Close()
}

func (conn *connectionWithReadError) LocalAddr() net.Addr {
	return conn.inner.LocalAddr()
}

func (conn *connectionWithReadError) RemoteAddr() net.Addr {
	return conn.inner.RemoteAddr()
}

func (conn *connectionWithReadError) SetDeadline(t time.Time) error {
	return conn.inner.SetDeadline(t)
}

func (conn *connectionWithReadError) SetReadDeadline(t time.Time) error {
	return conn.inner.SetReadDeadline(t)
}

func (conn *connectionWithReadError) SetWriteDeadline(t time.Time) error {
	return conn.inner.SetWriteDeadline(t)
}

func startServer(t *testing.T) net.Listener {
	listener, err := net.Listen("tcp", ":7000")

	assert.NotNil(t, listener)
	assert.NoError(t, err)

	go func() {
		for {
			conn, err := listener.Accept()

			if conn == nil {
				return
			}

			assert.NotNil(t, conn)
			assert.NoError(t, err)

			client := packet.NewStream(1024)

			client.OnError(func(err packet.IOError) {
				conn.Close()
			})

			client.SetConnection(conn)

			go func() {
				for msg := range client.Incoming {
					assert.Equal(t, "ping", string(msg.Data))
					client.Outgoing <- packet.New(0, []byte("pong"))
				}
			}()
		}
	}()

	return listener
}

func TestCommunication(t *testing.T) {
	// Server
	server := startServer(t)

	// Client
	conn, err := net.Dial("tcp", "localhost:7000")
	assert.NoError(t, err)

	client := packet.NewStream(1024)
	client.SetConnection(conn)

	// Send message
	client.Outgoing <- packet.New(0, []byte("ping"))

	// Receive message
	msg := <-client.Incoming

	// Check message contents
	assert.Equal(t, "pong", string(msg.Data))

	// Close connection
	conn.Close()

	// Send packet (will be buffered until reconnect finishes)
	client.Outgoing <- packet.New(0, []byte("ping"))

	// Reconnect
	conn, err = net.Dial("tcp", "localhost:7000")
	assert.NoError(t, err)

	// Hot-swap connection
	client.SetConnection(conn)

	// Receive message
	msg = <-client.Incoming
	assert.Equal(t, "pong", string(msg.Data))

	// Close
	client.Connection().Close()
	server.Close()
}

func TestDisconnect(t *testing.T) {
	listener, err := net.Listen("tcp", ":7000")
	assert.NotNil(t, listener)
	assert.NoError(t, err)
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()

			if conn == nil {
				return
			}

			assert.NotNil(t, conn)
			assert.NoError(t, err)

			client := packet.NewStream(1024)

			client.OnError(func(err packet.IOError) {
				conn.Close()
			})

			client.SetConnection(conn)

			go func() {
				for msg := range client.Incoming {
					assert.Equal(t, "ping", string(msg.Data))
					client.Outgoing <- packet.New(0, []byte("pong"))
				}
			}()
		}
	}()

	// Client
	conn, err := net.Dial("tcp", "localhost:7000")
	assert.NoError(t, err)
	defer conn.Close()

	client := packet.NewStream(1024)
	client.SetConnection(conn)

	// Send message
	client.Outgoing <- packet.New(0, []byte("ping"))

	// Receive message
	msg := <-client.Incoming

	// Check message contents
	assert.Equal(t, "pong", string(msg.Data))
}

func TestUtils(t *testing.T) {
	ping := packet.New(0, []byte("ping"))
	assert.Len(t, ping.Bytes(), 1+8+4)

	length, err := packet.Int64FromBytes(packet.Int64ToBytes(ping.Length))
	assert.NoError(t, err)
	assert.Equal(t, ping.Length, length)
}

func TestNilConnection(t *testing.T) {
	defer func() {
		err := recover()
		assert.NotNil(t, err)
		assert.Contains(t, err.(error).Error(), "nil")
	}()

	stream := packet.NewStream(0)
	stream.SetConnection(nil)
}

func TestNilOnError(t *testing.T) {
	defer func() {
		err := recover()
		assert.NotNil(t, err)
		assert.Contains(t, err.(error).Error(), "nil")
	}()

	stream := packet.NewStream(0)
	stream.OnError(nil)
}

func TestWriteTimeout(t *testing.T) {
	// Server
	server := startServer(t)
	defer server.Close()

	// Client
	conn, err := net.Dial("tcp", "localhost:7000")
	assert.NoError(t, err)
	defer conn.Close()

	client := packet.NewStream(0)
	client.SetConnection(conn)

	// Send message
	err = conn.SetWriteDeadline(time.Now())
	assert.NoError(t, err)
	client.Outgoing <- packet.New(0, []byte("ping"))
}

func TestReadError(t *testing.T) {
	// Server
	server := startServer(t)
	defer server.Close()

	// Client
	for failNumber := 1; failNumber <= 3; failNumber++ {
		conn, err := net.Dial("tcp", "localhost:7000")
		assert.NoError(t, err)

		// Make the 2nd read fail
		conn = &connectionWithReadError{
			inner:             conn,
			errorOnReadNumber: failNumber,
		}

		defer conn.Close()

		client := packet.NewStream(0)
		client.SetConnection(conn)

		// Send message
		client.Outgoing <- packet.New(0, []byte("ping"))
	}

	// Send a real message without read errors
	conn, err := net.Dial("tcp", "localhost:7000")
	assert.NoError(t, err)
	defer conn.Close()
	client := packet.NewStream(0)
	client.SetConnection(conn)

	// Send message
	client.Outgoing <- packet.New(0, []byte("ping"))

	// Receive message
	msg := <-client.Incoming

	// Check message contents
	assert.Equal(t, "pong", string(msg.Data))
}
