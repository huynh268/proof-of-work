package main

import (
	"bytes"
	"encoding/gob"
	"log"
)

// TXOut transaction output structure
type TXOut struct {
	Value      int
	PubKeyHash []byte
}

func (output *TXOut) Lock(address []byte) {
	pubKeyHash := Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	output.PubKeyHash = pubKeyHash
}

func (output *TXOut) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(output.PubKeyHash, pubKeyHash) == 0
}

func CreateTXOutput(value int, address string) *TXOut {
	txo := &TXOut{value, nil}
	txo.Lock([]byte(address))
	return txo
}

type TXOuts struct {
	Outputs []TXOut
}

func (outs TXOuts) SerializeTXOs() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(outs)

	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

func DeserializeTXOs(encodedTXO []byte) TXOuts {
	var result TXOuts

	decoder := gob.NewDecoder(bytes.NewReader(encodedTXO))
	err := decoder.Decode(result)

	if err != nil {
		log.Panic(err)
	}

	return result
}
