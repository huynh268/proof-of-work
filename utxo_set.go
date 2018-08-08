package main

import (
	"encoding/hex"
	"log"
	"os"

	"github.com/syndtr/goleveldb/leveldb"
)

const utxoDB = "chainstateDB"

type UTXOSet struct {
	Blockchain *Blockchain
	db         *leveldb.DB
}

func (u UTXOSet) Reindex() {
	if IsExists(utxoDB) {
		os.Remove(utxoDB)
	}

	db, err := leveldb.OpenFile(utxoDB, nil)
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	for txID, outs := range UTXO {
		key, err := hex.DecodeString(txID)
		if err != nil {
			log.Panic(err)
		}

		err = db.Put(key, outs.SerializeTXOs(), nil)
		if err != nil {
			log.Panic(err)
		}
	}
}

func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accummulated := 0
	db := u.db
	iter := db.NewIterator(nil, nil)

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		txID := hex.EncodeToString(key)
		outs := DeserializeTXOs(value)

		for outIdx, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) && accummulated < amount {
				accummulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
			}
		}
	}

	return accummulated, unspentOutputs
}

func (u UTXOSet) FindUTXO(pubKeyHash []byte) []TXOut {
	var UTXOs []TXOut
	db := u.db
	iter := db.NewIterator(nil, nil)

	for iter.Next() {
		key := iter.Key()
		value := iter.Value()

		outs := DeserializeTXOs(value)

		for _, out := range outs.Outputs {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (u UTXOSet) Update(block *Block) {
	db := u.db

	for _, tx := range block.Transactions {
		if !tx.IsCoinbase() {
			for _, vin := range tx.Vin {
				updatedOuts := TXOuts{}

				outsBytes, err := db.Get(vin.TxID, nil)
				if err != nil {
					log.Panic(err)
				}

				outs := DeserializeTXOs(outsBytes)

				for outIdx, out := range outs.Outputs {
					if outIdx != vin.Vout {
						updatedOuts.Outputs = append(updatedOuts.Outputs, out)
					}
				}

				if len(updatedOuts.Outputs) == 0 {
					err = db.Delete(vin.TxID, nil)
					if err != nil {
						log.Panic(err)
					}
				} else {
					err = db.Put(vin.TxID, updatedOuts.SerializeTXOs(), nil)
					if err != nil {
						log.Panic(err)
					}
				}
			}
		}

		newOutputs := TXOuts{}
		for _, out := range tx.Vout {
			newOutputs.Outputs = append(newOutputs.Outputs, out)
		}

		err := db.Put(tx.ID, newOutputs.SerializeTXOs(), nil)
		if err != nil {
			log.Panic(err)
		}
	}
}
