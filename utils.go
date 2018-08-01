package main

import "encoding/binary"

// IntToByte converts an int64 to a byte array
func IntToByte(i int64) []byte {
	b := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(b, uint64(i))

	return b
}
