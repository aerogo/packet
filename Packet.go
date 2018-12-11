package packet

import (
	"encoding/binary"
	"io"
)

// Send buffer size determines how many bytes we send in a single TCP write call.
// This can be anything from 1 to 65495.
// A good default value for this can be read from: /proc/sys/net/ipv4/tcp_wmem
const sendBufferSize = 16384

// Packet represents a single network message.
// It has a byte code indicating the type of the message
// and a data payload in the form of a byte slice.
type Packet struct {
	Type   byte
	Length int64
	Data   []byte
}

// New creates a new packet.
// It expects a byteCode for the type of message and
// a data parameter in the form of a byte slice.
func New(byteCode byte, data []byte) *Packet {
	return &Packet{
		Type:   byteCode,
		Length: int64(len(data)),
		Data:   data,
	}
}

// Write writes the packet to the IO device.
func (packet *Packet) Write(writer io.Writer) error {
	_, err := writer.Write([]byte{packet.Type})

	if err != nil {
		return err
	}

	err = binary.Write(writer, binary.BigEndian, packet.Length)

	if err != nil {
		return err
	}

	n := 0
	bytesWritten := 0
	writeUntil := 0

	for bytesWritten < len(packet.Data) {
		writeUntil = bytesWritten + sendBufferSize

		if writeUntil > len(packet.Data) {
			writeUntil = len(packet.Data)
		}

		n, err = writer.Write(packet.Data[bytesWritten:writeUntil])

		if err != nil {
			return err
		}

		bytesWritten += n
	}

	return err
}

// Bytes returns the raw byte slice serialization of the packet.
func (packet *Packet) Bytes() []byte {
	result := []byte{packet.Type}
	result = append(result, Int64ToBytes(packet.Length)...)
	result = append(result, packet.Data...)
	return result
}
