package packet_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aerogo/packet"
)

func startServer(t *testing.T) {
	listener, err := net.Listen("tcp", ":7000")

	assert.NotNil(t, listener)
	assert.NoError(t, err)

	go func() {
		for {
			conn, err := listener.Accept()

			assert.NotNil(t, conn)
			assert.NoError(t, err)

			client := packet.NewStream(1024)
			client.SetConnection(conn)

			go func() {
				for msg := range client.Incoming {
					assert.Equal(t, "ping", string(msg.Data))
					client.Outgoing <- packet.New(0, []byte("pong"))
				}
			}()
		}
	}()
}

func TestCommunication(t *testing.T) {
	// Server
	startServer(t)

	// Client
	conn, err := net.Dial("tcp", "localhost:7000")
	assert.NoError(t, err)

	client := packet.NewStream(1024)
	client.SetConnection(conn)

	// Send message
	client.Outgoing <- packet.New(0, []byte("ping"))

	// Receive message
	msg := <-client.Incoming
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
}

func TestUtils(t *testing.T) {
	ping := packet.New(0, []byte("ping"))
	assert.Len(t, ping.Bytes(), 1+8+4)

	length, err := packet.Int64FromBytes(packet.Int64ToBytes(ping.Length))
	assert.NoError(t, err)
	assert.Equal(t, ping.Length, length)
}
