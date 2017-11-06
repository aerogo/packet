package packet_test

import (
	"net"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/aerogo/packet"
)

func TestCommunication(t *testing.T) {
	// Server
	listener, err := net.Listen("tcp", ":7000")
	assert.NoError(t, err)

	go func() {
		conn, err := listener.Accept()
		assert.NoError(t, err)

		client := packet.NewStream(1024)
		client.SetConnection(conn)

		go func() {
			for msg := range client.Incoming {
				assert.Equal(t, "ping", string(msg.Data))
				client.Outgoing <- packet.New(0, []byte("pong"))
			}
		}()
	}()

	// Client
	conn, err := net.Dial("tcp", "localhost:7000")
	assert.NoError(t, err)

	client := packet.NewStream(1024)
	client.SetConnection(conn)

	client.Outgoing <- packet.New(0, []byte("ping"))
	msg := <-client.Incoming
	assert.Equal(t, "pong", string(msg.Data))
}
