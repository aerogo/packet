package packet

import (
	"bytes"
	"encoding/binary"
)

// Int64FromBytes converts a Big Endian representation to an int64.
func Int64FromBytes(b []byte) (int64, error) {
	buffer := bytes.NewReader(b)
	var result int64
	err := binary.Read(buffer, binary.BigEndian, &result)
	return result, err
}

// Int64ToBytes converts an int64 to its Big Endian representation.
func Int64ToBytes(i int64) []byte {
	buffer := bytes.Buffer{}
	_ = binary.Write(&buffer, binary.BigEndian, i)
	return buffer.Bytes()
}
