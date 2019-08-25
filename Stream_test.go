package packet_test

import (
	"errors"
	"net"
	"strconv"
	"testing"
	"time"

	"github.com/aerogo/packet"
	"github.com/akyoto/assert"
)

// connectionWithReadError errors the Read call after `errorOnReadNumber` tries.
type connectionWithReadError struct {
	net.Conn
	countReads        int
	errorOnReadNumber int
}

func (conn *connectionWithReadError) Read(buffer []byte) (int, error) {
	conn.countReads++

	if conn.countReads == conn.errorOnReadNumber {
		return 0, errors.New("Artificial error")
	}

	return conn.Conn.Read(buffer)
}

func startServer(t *testing.T, port int) net.Listener {
	listener, err := net.Listen("tcp", ":"+strconv.Itoa(port))

	assert.NotNil(t, listener)
	assert.Nil(t, err)

	go func() {
		for {
			conn, err := listener.Accept()

			if conn == nil {
				return
			}

			assert.NotNil(t, conn)
			assert.Nil(t, err)

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
	server := startServer(t, 7000)
	defer server.Close()

	// Client
	conn, err := net.Dial("tcp", "localhost:7000")
	assert.Nil(t, err)

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
	assert.Nil(t, err)

	// Hot-swap connection
	client.SetConnection(conn)

	// Receive message
	msg = <-client.Incoming
	assert.Equal(t, "pong", string(msg.Data))

	// Close
	client.Close()
}

func TestDisconnect(t *testing.T) {
	listener, err := net.Listen("tcp", ":7001")
	assert.NotNil(t, listener)
	assert.Nil(t, err)
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()

			if conn == nil {
				return
			}

			assert.NotNil(t, conn)
			assert.Nil(t, err)

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
	conn, err := net.Dial("tcp", "localhost:7001")
	assert.Nil(t, err)
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
	assert.Equal(t, len(ping.Bytes()), 1+8+4)

	length, err := packet.Int64FromBytes(packet.Int64ToBytes(ping.Length))
	assert.Nil(t, err)
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
	server := startServer(t, 7002)
	defer server.Close()

	// Client
	conn, err := net.Dial("tcp", "localhost:7002")
	assert.Nil(t, err)
	defer conn.Close()

	client := packet.NewStream(0)
	client.SetConnection(conn)

	// Send message
	err = conn.SetWriteDeadline(time.Now())
	assert.Nil(t, err)
	client.Outgoing <- packet.New(0, []byte("ping"))
}

func TestReadError(t *testing.T) {
	// Server
	server := startServer(t, 7003)
	defer server.Close()

	// Client
	for failNumber := 1; failNumber <= 3; failNumber++ {
		conn, err := net.Dial("tcp", "localhost:7003")
		assert.NotNil(t, conn)
		assert.Nil(t, err)

		// Make the 2nd read fail
		conn = &connectionWithReadError{
			Conn:              conn,
			errorOnReadNumber: failNumber,
		}

		client := packet.NewStream(1)
		client.SetConnection(conn)

		// Send message
		client.Outgoing <- packet.New(0, []byte("ping"))

		// err = conn.Close()
		// assert.Nil(t, err)
	}

	// Send a real message without read errors
	conn, err := net.Dial("tcp", "localhost:7003")
	assert.Nil(t, err)
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
