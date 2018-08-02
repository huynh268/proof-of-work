package main

func main() {
	blockchain := CreateBlockchain()
	defer blockchain.db.Close()

	cli := CLI{blockchain}
	cli.Run()

	// blockchain.AddBlock("send 1 eth to A")
	// blockchain.AddBlock("send 2 xlm to B")
	//
	// spew.Dump(blockchain)
	// fmt.Println()
	//
	// for _, block := range blockchain.blocks {
	// 	fmt.Printf("Data: %s\n", block.Data)
	// 	fmt.Printf("PrevHash: %x\n", block.PrevHash)
	// 	fmt.Printf("Hash: %x\n", block.Hash)
	//
	// 	pow := CreatePoW(block)
	// 	valid := "valid"
	// 	if !pow.Validate() {
	// 		valid = "invalid"
	// 	}
	// 	fmt.Printf("PoW: %s\n", valid)
	//
	// 	fmt.Println()
	// }
}
