package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"
)

// IntToByte converts an int64 to a byte array
func IntToByte(i int64) []byte {
	b := make([]byte, binary.MaxVarintLen64)
	binary.BigEndian.PutUint64(b, uint64(i))
	return b
}

// Serialize serializes input to a byte array
func Serialize(x interface{}) []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(x)

	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}
