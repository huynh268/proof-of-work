package main

import "bytes"

// TXIn transaction input structure
type TXIn struct {
	TxID      []byte
	Vout      int
	Signature []byte
	PublicKey []byte
}

func (intput *TXIn) UseKey(pubKeyHash []byte) bool {
	lockingHash := HashPublicKey(intput.PublicKey)
	return bytes.Compare(lockingHash, pubKeyHash) == 0
}
