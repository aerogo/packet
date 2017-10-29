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
		n, err := stream.Connection.Read(typeBuffer)

		if err != nil || n != 1 {
			break
		}

		n, err = stream.Connection.Read(lengthBuffer)

		if err != nil || n != 8 {
			break
		}

		length, err := fromBytes(lengthBuffer)

		if err != nil {
			break
		}

		data := make([]byte, length)
		readLength := 0

		for readLength < len(data) {
			n, err := stream.Connection.Read(data[readLength:])
			readLength += n

			if err != nil {
				break
			}
		}

		if readLength < len(data) {
			break
		}

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
