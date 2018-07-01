package packet

// Packet represents a single network message.
// It has a byte code indicating the type of the message
// and a data payload in the form of a byte slice.
type Packet struct {
	Type   byte
	Length int64
	Data   []byte
}

// New creates a new packet.
// It expects a |byteCode| for the type of message and
// a |data| parameter in the form of a byte slice.
func New(byteCode byte, data []byte) *Packet {
	return &Packet{
		Type:   byteCode,
		Length: int64(len(data)),
		Data:   data,
	}
}

// Bytes returns the raw byte slice serialization of the packet.
func (packet *Packet) Bytes() []byte {
	result := []byte{packet.Type}
	result = append(result, Int64ToBytes(packet.Length)...)
	result = append(result, packet.Data...)
	return result
}
