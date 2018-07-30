package main

// BlockChain structure
type BlockChain struct {
	blocks []*Block
}

// AddBlock adds a new block to blockchain
func (blockchain *BlockChain) AddBlock(data string) {
	prevBlock := blockchain.blocks[len(blockchain.blocks)-1]
	newBlock := CreateBlock(data, prevBlock.Hash)
	blockchain.blocks = append(blockchain.blocks, newBlock)
}

// CreateGenesisBlock creates a genesis block - the first block of the chain
func CreateGenesisBlock() *Block {
	return CreateBlock("Genesis Block", []byte{})
}

// CreateBlockchain creates a new blockchain
func CreateBlockchain() *BlockChain {
	return &BlockChain{[]*Block{CreateGenesisBlock()}}
}
