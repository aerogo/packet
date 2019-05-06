# {name}

{go:header}

Send network packets over a TCP or UDP connection.

## Packet

Packet is the main class representing a single network message. It has a byte code indicating the type of the message and a `[]byte` type payload.

## Stream

A stream has a send and receive channel with a hot-swappable connection for reconnects.
The user has the responsibility to register a callback to consume errors via `OnError`.

## Example

```go
// Connect to a server
conn, _ := net.Dial("tcp", "localhost:7000")

// Create a stream
stream := packet.NewStream(1024)
stream.SetConnection(conn)

// Send a message
stream.Outgoing <- packet.New(0, []byte("ping"))

// Receive message
msg := <-stream.Incoming

// Check message contents
if string(msg.Data) != "pong" {
	panic("Something went wrong!")
}
```

## Hot-swap example

```go
// Close server connection to simulate server death
server.Close()

// Send packet while server is down (will be buffered until it reconnects)
client.Outgoing <- packet.New(0, []byte("ping"))

// Reconnect
newServer, _ := net.Dial("tcp", "localhost:7000")

// Hot-swap connection
client.SetConnection(newServer)

// The previously buffered messages in the Outgoing channel will be sent now.
```

{go:footer}
