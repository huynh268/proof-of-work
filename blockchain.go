package main

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"

	"github.com/syndtr/goleveldb/leveldb"
)

const file = "blocksDB"
const genesisCoinbaseData = "Genesis block!"

// Blockchain structure
type Blockchain struct {
	tip []byte
	db  *leveldb.DB
}

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

	blockchain := Blockchain{tip, db}

	return &blockchain
}

// MineBlock mines a new block
func (bc *Blockchain) MineBlock(transactions []*Transaction) *Block {
	prevHash, err := bc.db.Get([]byte("prevHash"), nil)
	if err != nil {
		log.Panic(err)
	}

	if prevHash == nil {
		fmt.Println("Blockchain does not exist.")
	}

	for _, tx := range transactions {
		if !bc.VerifyTransaction(tx) {
			log.Panic("ERROR: Invalid transaction!")
		}
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

//FindUTXO finds all unspent transaction outputs
func (bc *Blockchain) FindUTXO() map[string]TXOuts {
	UTXO := make(map[string]TXOuts)
	spentTXOs := make(map[string][]int)
	bci := bc.Iterator()

	for {
		block := bci.NextBlock()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				if spentTXOs[txID] != nil { //check if output is spent
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}

			if !tx.IsCoinbase() {
				for _, in := range tx.Vin {
					inIdx := hex.EncodeToString(in.TxID)
					spentTXOs[inIdx] = append(spentTXOs[inIdx], in.Vout)
				}
			}
		}

		if len(block.Transactions) == 0 {
			break
		}
	}

	return UTXO
}

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.NextBlock()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction is not found.")
}

func (bc *Blockchain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.TxID)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}
	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]Transaction)

	for _, vin := range tx.Vin {
		prevTX, err := bc.FindTransaction(vin.TxID)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}
