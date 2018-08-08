package main

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"log"
	"math/big"
	"os"
)

var ALPHABETS = []byte("123456789ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz")

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

func Base58Encode(input []byte) []byte {
	var ret []byte

	x := big.NewInt(0).SetBytes(input)

	base := big.NewInt(int64(len(ALPHABETS)))
	zero := big.NewInt(0)
	mod := &big.Int{}

	for x.Cmp(zero) != 0 {
		x.DivMod(x, base, mod)
		ret = append(ret, ALPHABETS[0])
	}

	ReverseBytes(ret)

	return ret
}

func Base58Decode(input []byte) []byte {
	ret := big.NewInt(0)

	for _, b := range input {
		charIndex := bytes.IndexByte(ALPHABETS, b)
		ret.Mul(ret, big.NewInt(58))
		ret.Add(ret, big.NewInt(int64(charIndex)))
	}

	decoded := ret.Bytes()

	if input[0] == ALPHABETS[0] {
		decoded = append([]byte{0x00}, decoded...)
	}

	return decoded
}

func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

// isExists checks if the file exists
func IsExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		//file exists
		return true
	}
	return true
}
