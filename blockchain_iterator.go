package main

import (
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

// BlockChainIterator iterates over the blocks in database
type BlockchainIterator struct {
	currentBlockHash []byte
	db               *leveldb.DB
}

// NextBlock gets the next block
func (iter *BlockchainIterator) NextBlock() *Block {
	var block *Block

	encodedBlock, err := iter.db.Get(iter.currentBlockHash, nil)
	if err != nil {
		log.Panic(err)
	}
	block = DeserializeBlock(encodedBlock)
	iter.currentBlockHash = block.PrevHash

	return block
}
