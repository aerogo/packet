package packet

import (
	"fmt"
	"net"
)

// Stream ...
type Stream struct {
	Connection net.Conn
	Incoming   chan *Packet
	Outgoing   chan *Packet
}

// Read ...
func (stream *Stream) Read() {
	typeBuffer := make([]byte, 1)
	lengthBuffer := make([]byte, 8)

	for {
		_, err := stream.Connection.Read(typeBuffer)

		if err != nil {
			fmt.Println("R Packet Type fail", err)
			break
		}

		_, err = stream.Connection.Read(lengthBuffer)

		if err != nil {
			fmt.Println("R Packet Length fail", stream.Connection.RemoteAddr(), err)
			break
		}

		length, err := fromBytes(lengthBuffer)

		if err != nil {
			fmt.Println("R Packet Length decode fail", stream.Connection.RemoteAddr(), err)
			break
		}

		data := make([]byte, length)
		readLength := 0

		for readLength < len(data) {
			n, err := stream.Connection.Read(data[readLength:])
			readLength += n

			if err != nil {
				fmt.Println("R Data read fail", stream.Connection.RemoteAddr(), err)
				break
			}
		}

		if readLength < len(data) {
			fmt.Println("R Data read length fail", stream.Connection.RemoteAddr(), err)
			break
		}

		// msg, err := ioutil.ReadAll(stream.Connection)

		// if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
		// 	fmt.Println("R Timeout", stream.Connection.RemoteAddr())
		// 	break
		// }

		// if err != nil && err != io.EOF && strings.Contains(err.Error(), "Connection reset") {
		// 	fmt.Println("R Disconnected", stream.Connection.RemoteAddr())
		// 	break
		// }

		stream.Incoming <- New(typeBuffer[0], data)
	}
}

// Write ...
func (stream *Stream) Write() {
	for packet := range stream.Outgoing {
		msg := packet.Bytes()
		totalWritten := 0

		for totalWritten < len(msg) {
			writtenThisCall, err := stream.Connection.Write(msg[totalWritten:])

			if err != nil {
				fmt.Println("Error writing", err)
				return
			}

			totalWritten += writtenThisCall
		}
	}
}
