package main

import (
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	blockchain := CreateBlockchain()

	blockchain.AddBlock("send 1 eth to A")
	blockchain.AddBlock("send 2 xlm to B")

	spew.Dump(blockchain)
	fmt.Println()

	for _, block := range blockchain.blocks {
		fmt.Printf("Data: %s\n", block.Data)
		fmt.Printf("PrevHash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println()
	}
}