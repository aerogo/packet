# packet

[![Godoc][godoc-image]][godoc-url]
[![Report][report-image]][report-url]
[![Tests][tests-image]][tests-url]
[![Coverage][coverage-image]][coverage-url]
[![Patreon][patreon-image]][patreon-url]

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
if string(msg.Data) != "pong" 
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

## Style

Please take a look at the [style guidelines](https://github.com/akyoto/quality/blob/master/STYLE.md) if you'd like to make a pull request.

## Sponsors

| [![Scott Rayapoullé](https://avatars3.githubusercontent.com/u/11772084?s=70&v=4)](https://github.com/soulcramer) | [![Eduard Urbach](https://avatars2.githubusercontent.com/u/438936?s=70&v=4)](https://twitter.com/eduardurbach) |
| --- | --- |
| [Scott Rayapoullé](https://github.com/soulcramer) | [Eduard Urbach](https://eduardurbach.com) |

Want to see [your own name here?](https://www.patreon.com/eduardurbach)

[godoc-image]: https://godoc.org/github.com/aerogo/packet?status.svg
[godoc-url]: https://godoc.org/github.com/aerogo/packet
[report-image]: https://goreportcard.com/badge/github.com/aerogo/packet
[report-url]: https://goreportcard.com/report/github.com/aerogo/packet
[tests-image]: https://cloud.drone.io/api/badges/aerogo/packet/status.svg
[tests-url]: https://cloud.drone.io/aerogo/packet
[coverage-image]: https://codecov.io/gh/aerogo/packet/graph/badge.svg
[coverage-url]: https://codecov.io/gh/aerogo/packet
[patreon-image]: https://img.shields.io/badge/patreon-donate-green.svg
[patreon-url]: https://www.patreon.com/eduardurbach
