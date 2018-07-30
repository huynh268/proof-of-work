package main

import (
	"bytes"
	"crypto/sha256"
	"strconv"
	"time"
)

// Block structure
type Block struct {
	Timestamp int64
	Data      []byte
	PrevHash  []byte
	Hash      []byte
}

// CreateHash creates a hash value for a block
func (block *Block) CreateHash() {
	timestamp := []byte(strconv.FormatInt(block.Timestamp, 10))
	headers := bytes.Join([][]byte{block.PrevHash, block.Data, timestamp}, []byte{})
	hash := sha256.Sum256(headers)

	block.Hash = hash[:]
}

// CreateBlock creates a new block
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{time.Now().Unix(), []byte(data), prevHash, []byte{}}
	block.CreateHash()
	return block
}
