package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

// CLI structure
type CLI struct {
	blockchain *Blockchain
}

// Run runs command lines
func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	//addBlockCmd := flag.NewFlagSet("addBlock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)
	//addBlockData := addBlockCmd.String("data", "", "Block data")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to check balance")
	switch os.Args[1] {
	case "createBlockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "getBalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printChain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			os.Exit(1)
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		cli.getBalance(*getBalanceAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}
}

func (cli *CLI) createBlockchain(address string) {
	fmt.Println("creating....")
	bc := CreateBlockchain(address)
	defer bc.db.Close()

	fmt.Println("Done!")
}

/*
func (cli *CLI) addBlock(tx []*Transaction) {
	cli.blockchain.AddBlock(tx)
	fmt.Println("Success!")
}*/

func (cli *CLI) printChain() {
	bc := CreateBlockchain("printChain")
	defer bc.db.Close()

	blockchainIterator := bc.Iterator()

	for {
		block := blockchainIterator.NextBlock()

		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Printf("Previous hash: %x\n", block.PrevHash)
		pow := CreatePoW(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println("Transaction: ")
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()

		if len(block.PrevHash) == 0 {
			break
		}
	}

	fmt.Println("Success!")
}

func (cli *CLI) getBalance(address string) {
	bc := CreateBlockchain(address)
	defer bc.db.Close()

	balance := 0
	UTXOs := bc.FindUTXO(address)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" createblockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}
