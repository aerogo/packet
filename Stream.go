package packet

import (
	"net"
)

// Stream ...
type Stream struct {
	Connection net.Conn
	Incoming   chan *Packet
	Outgoing   chan *Packet
}

// Read ...
func (stream *Stream) Read() error {
	typeBuffer := make([]byte, 1)
	lengthBuffer := make([]byte, 8)

	for {
		_, err := stream.Connection.Read(typeBuffer)

		if err != nil {
			return err
		}

		_, err = stream.Connection.Read(lengthBuffer)

		if err != nil {
			return err
		}

		length, err := fromBytes(lengthBuffer)

		if err != nil {
			return err
		}

		data := make([]byte, length)
		readLength := 0

		for readLength < len(data) {
			n, err := stream.Connection.Read(data[readLength:])
			readLength += n

			if err != nil {
				return err
			}
		}

		stream.Incoming <- New(typeBuffer[0], data)
	}
}

// Write ...
func (stream *Stream) Write() error {
	for packet := range stream.Outgoing {
		msg := packet.Bytes()
		totalWritten := 0

		for totalWritten < len(msg) {
			writtenThisCall, err := stream.Connection.Write(msg[totalWritten:])

			if err != nil {
				return err
			}

			totalWritten += writtenThisCall
		}
	}

	return nil
}
