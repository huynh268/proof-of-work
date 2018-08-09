package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/syndtr/goleveldb/leveldb"
)

// CLI structure
type CLI struct {
	blockchain *Blockchain
	utxo       *UTXOSet
}

// Run runs command lines
func (cli *CLI) Run() {
	cli.validateArgs()

	createBlockchainCmd := flag.NewFlagSet("createBlockchain", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getBalance", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	//addBlockCmd := flag.NewFlagSet("addBlock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printChain", flag.ExitOnError)

	//addBlockData := addBlockCmd.String("data", "", "Block data")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis reward to")
	getBalanceAddress := getBalanceCmd.String("address", "", "The address to check balance")
	sendFrom := sendCmd.String("from", "", "Send from")
	sendTo := sendCmd.String("to", "", "Send to")
	sendAmount := sendCmd.Int("amount", 0, "Sent amount")

	switch os.Args[1] {
	case "createBlockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
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

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			os.Exit(1)
		}
		cli.send(*sendFrom, *sendTo, *sendAmount)
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

	db, err := leveldb.OpenFile("utxoDB", nil)
	if err != nil {
		log.Panic(err)
	}
	UTXOSet := UTXOSet{bc, db}
	defer db.Close()

	UTXOSet.Reindex()

	fmt.Println("Done!")
}

func (cli *CLI) send(from, to string, amount int, nodeID string, mineBlock bool) {
	fmt.Printf("%d coins sent to %s from %s", amount, to, from)

	bc := CreateBlockchain(from)
	db, err := leveldb.OpenFile("utxoDB", nil)
	if err != nil {
		log.Panic(err)
	}
	UTXOSet := UTXOSet{bc, db}
	defer bc.db.Close()
	defer db.Close()

	wallets, err := CreateWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}

	wallet := wallets.GetWallet(from)

	tx := NewUTXOTransaction(&wallet, to, amount, &UTXOSet)
	bc.MineBlock([]*Transaction{tx})
	fmt.Println("Success!")
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
	db, err := leveldb.OpenFile("utxoDB", nil)
	if err != nil {
		log.Panic(err)
	}
	UTXOSet := UTXOSet{bc, db}
	defer bc.db.Close()
	defer db.Close()

	balance := 0
	pubKeyHash := Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)

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
