package packet

import (
	"net"
	"sync/atomic"
	"time"
)

// Stream ...
type Stream struct {
	connection atomic.Value
	Incoming   chan *Packet
	Outgoing   chan *Packet
}

// NewStream ...
func NewStream(channelBufferSize int) *Stream {
	stream := &Stream{
		Incoming: make(chan *Packet, channelBufferSize),
		Outgoing: make(chan *Packet, channelBufferSize),
	}

	go stream.Read()
	go stream.Write()

	return stream
}

// Connection ...
func (stream *Stream) Connection() net.Conn {
	return stream.connection.Load().(net.Conn)
}

// SetConnection ...
func (stream *Stream) SetConnection(connection net.Conn) {
	stream.connection.Store(connection)
}

// Read ...
func (stream *Stream) Read() error {
	typeBuffer := make([]byte, 1, 1)
	lengthBuffer := make([]byte, 8, 8)

	for {
		connection := stream.Connection()

		if connection == nil {
			time.Sleep(1 * time.Millisecond)
			continue
		}

		_, err := connection.Read(typeBuffer)

		if err != nil {
			return err
		}

		_, err = connection.Read(lengthBuffer)

		if err != nil {
			return err
		}

		length, err := Int64FromBytes(lengthBuffer)

		if err != nil {
			return err
		}

		data := make([]byte, length)
		readLength := 0

		for readLength < len(data) {
			n, err := connection.Read(data[readLength:])
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
		connection := stream.Connection()
		msg := packet.Bytes()
		totalWritten := 0

		for totalWritten < len(msg) {
			writtenThisCall, err := connection.Write(msg[totalWritten:])

			if err != nil {
				return err
			}

			totalWritten += writtenThisCall
		}
	}

	return nil
}
