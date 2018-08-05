package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

// Block structure
type Block struct {
	Timestamp    int64
	Transactions []*Transaction
	PrevHash     []byte
	Hash         []byte
	Nonce        int
}

// CreateBlock creates a new block
func CreateBlock(transactions []*Transaction, prevHash []byte) *Block {
	block := &Block{time.Now().Unix(), transactions, prevHash, []byte{}, 0}
	pow := CreatePoW(block)
	nonce, hash := pow.Mine()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// CreateGenesisBlock creates a genesis block - the first block of the chain
func CreateGenesisBlock(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

// SerializeBlock serializes a block to byte array
func (block *Block) SerializeBlock() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)
	err := encoder.Encode(block)

	if err != nil {
		log.Panic(err)
	}

	return result.Bytes()
}

// DeserializeBlock deserializes a block
func DeserializeBlock(encodedBlock []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(encodedBlock))
	err := decoder.Decode(&block)

	if err != nil {
		log.Panic(err)
	}

	return &block
}

//HashTransaction creats a hash of transaction
func (block *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range block.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))
	return txHash[:]
}
