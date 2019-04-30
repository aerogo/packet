# packet

[![Reference][godoc-image]][godoc-url]
[![Report][report-image]][report-url]
[![Tests][tests-image]][tests-url]
[![Coverage][codecov-image]][codecov-url]
[![License][license-image]][license-url]

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

## Author

| [![Eduard Urbach on Twitter](https://gravatar.com/avatar/16ed4d41a5f244d1b10de1b791657989?s=70)](https://twitter.com/eduardurbach "Follow @eduardurbach on Twitter") |
|---|
| [Eduard Urbach](https://eduardurbach.com) |

[godoc-image]: https://godoc.org/github.com/aerogo/packet?status.svg
[godoc-url]: https://godoc.org/github.com/aerogo/packet
[report-image]: https://goreportcard.com/badge/github.com/aerogo/packet
[report-url]: https://goreportcard.com/report/github.com/aerogo/packet
[tests-image]: https://cloud.drone.io/api/badges/aerogo/packet/status.svg
[tests-url]: https://cloud.drone.io/aerogo/packet
[codecov-image]: https://codecov.io/gh/aerogo/packet/graph/badge.svg
[codecov-url]: https://codecov.io/gh/aerogo/packet
[license-image]: https://img.shields.io/badge/license-MIT-blue.svg
[license-url]: https://github.com/aerogo/packet/blob/master/LICENSE
