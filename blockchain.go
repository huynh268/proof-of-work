package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

const file = "blocksDB"
const genesisCoinbaseData = "Genesis block!"

// Blockchain structure
type Blockchain struct {
	tip []byte
	db  *leveldb.DB
}

/*
// AddBlock adds a new block to blockchain
func (blockchain *Blockchain) AddBlock(transaction []*Transaction) {
	data, err := blockchain.db.Get([]byte("prevHash"), nil)
	if err != nil {
		log.Panic(err)
	}
	//fmt.Printf("data: %x\n", dat)
	if data == nil {
		fmt.Println("Blockchain does not exist.")
	}

	newBlock := CreateBlock(transaction, data)

	err = blockchain.db.Put(newBlock.Hash, newBlock.SerializeBlock(), nil)
	if err != nil {
		log.Panic(err)
	}

	err = blockchain.db.Put([]byte("prevHash"), newBlock.Hash, nil)
	if err != nil {
		log.Panic(err)
	}

	blockchain.tip = newBlock.Hash
}*/

// CreateBlockchain creates a new blockchain
func CreateBlockchain(address string) *Blockchain {
	var tip []byte

	db, err := leveldb.OpenFile(file, nil)
	if err != nil {
		log.Panic(err)
	}

	// TODO: check this again!!!
	data, err := db.Get([]byte("prevHash"), nil)
	if err != nil {
		if err.Error() == "leveldb: not found" {
			fmt.Printf("Database is not created, error: %s\n\n", err)
		} else {
			log.Panic(err)
		}
	}

	if data == nil {
		fmt.Println("Blockchain does not exists. Creating a new one...")

		coinbaseTX := CreateCoinbaseTX(address, genesisCoinbaseData)
		genesisBlock := CreateGenesisBlock(coinbaseTX)

		err = db.Put(genesisBlock.Hash, genesisBlock.SerializeBlock(), nil)
		if err != nil {
			log.Panic(err)
		}

		err = db.Put([]byte("prevHash"), genesisBlock.Hash, nil)
		if err != nil {
			log.Panic(err)
		}

		tip = genesisBlock.Hash
	} else {
		tip = data
	}

	if isExists(file) {

	}

	blockchain := Blockchain{tip, db}

	return &blockchain
}

func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	prevHash, err := bc.db.Get([]byte("prevHash"), nil)
	if err != nil {
		log.Panic(err)
	}

	if prevHash == nil {
		fmt.Println("Blockchain does not exist.")
	}

	newBlock := CreateBlock(transactions, prevHash)

	err = bc.db.Put(newBlock.Hash, newBlock.SerializeBlock(), nil)
	if err != nil {
		log.Panic(err)
	}

	err = bc.db.Put([]byte("prevHash"), newBlock.Hash, nil)
	if err != nil {
		log.Panic(err)
	}

	return newBlock
}

// Iterator iterates a blockchain
func (blockchain *Blockchain) Iterator() *BlockchainIterator {
	blockchainIterator := &BlockchainIterator{blockchain.tip, blockchain.db}
	return blockchainIterator
}

// isExists checks if the file exists
func isExists(file string) bool {
	if _, err := os.Stat(file); err == nil {
		//file exists
		return true
	}
	return true
}

func (bc *Blockchain) FindUnspentTransactions(address string) []Transaction {
	var unspentTXs []Transaction
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.NextBlock()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}

				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.Vin {
					if in.CanUnlockOutputWith(address) {
						inTxID := hex.EncodeToString(in.Txid)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Vout)
					}
				}
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO(address string) []TXOut {
	var UTXOs []TXOut

	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)

		for outIdx, out := range tx.Vout {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}
