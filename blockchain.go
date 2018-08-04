package main

import (
	"fmt"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

const file = "blocksDB"
const genesisCoinbaseData = "Genesis block!"

// BlockChain structure
type Blockchain struct {
	tip []byte
	db  *leveldb.DB
}

// AddBlock adds a new block to blockchain
func (blockchain *Blockchain) AddBlock(transaction []*Transaction) {
	var prevBlockHash []byte

	dat, err := blockchain.db.Get([]byte("prevBlockHash"), nil)
	if err != nil {
		log.Panic(err)
	}
	fmt.Printf("data: %x\n", dat)
	if dat == nil {
		fmt.Println("Blockchain does not exist.")
	}

	prevBlockHash = dat

	newBlock := CreateBlock(transaction, prevBlockHash)

	err = blockchain.db.Put(newBlock.Hash, newBlock.SerializeBlock(), nil)
	if err != nil {
		log.Panic(err)
	}

	err = blockchain.db.Put([]byte("prevBlockHash"), newBlock.Hash, nil)
	if err != nil {
		log.Panic(err)
	}

	blockchain.tip = newBlock.Hash
}

// CreateGenesisBlock creates a genesis block - the first block of the chain
func CreateGenesisBlock(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

// CreateBlockchain creates a new blockchain
func CreateBlockchain(address string) *Blockchain {
	var tip []byte

	db, err := leveldb.OpenFile(file, nil)
	if err != nil {
		log.Panic(err)
	}

	data, err := db.Get([]byte("prevBlockHash"), nil)
	if err != nil {
		log.Panic(err)
	}

	if data == nil {
		fmt.Println("Blockchain does not exists. Creating a new one...")

		coinbaseTX := CreateCoinbaseTX(address, genesisCoinbaseData)
		genesisBlock := CreateGenesisBlock(coinbaseTX)

		err = db.Put(genesisBlock.Hash, genesisBlock.SerializeBlock(), nil)
		if err != nil {
			log.Panic(err)
		}

		err = db.Put([]byte("prevBlockHash"), genesisBlock.Hash, nil)
		if err != nil {
			log.Panic(err)
		}

		tip = genesisBlock.Hash
	} else {
		tip = data
	}

	blockchain := Blockchain{tip, db}

	return &blockchain
}

func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	blockchainIterator := &BlockchainIterator{blockchain.tip, blockchain.db}
	return blockchainIterator
}

func dbExists(file string) bool {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		return false
	}
	return true
}
