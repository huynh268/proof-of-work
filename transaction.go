package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

const firstReward = 10

// Transaction structure
type Transaction struct {
	ID   []byte
	Vin  []TXIn
	Vout []TXOut
}

// CreateCoinbaseTX creates the 1st transaction
func CreateCoinbaseTX(toAddress, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to %s", toAddress)
	}

	txin := TXIn{[]byte{}, -1, data}
	txout := TXOut{firstReward, toAddress}
	tx := Transaction{nil, []TXIn{txin}, []TXOut{txout}}
	tx.ID = tx.Hash()

	return &tx
}

// Hash creates hash
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(Serialize(&txCopy))

	return hash[:]
}

func (tx Transaction) serialize() []byte {
	var encoded bytes.Buffer

	newEncode := gob.NewEncoder(&encoded)
	err := newEncode.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

func (in *TXIn) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

func (out *TXOut) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].Txid) == 0 && tx.Vin[0].Vout == -1
}

func NewUTXOTransaction(from, to string, amount int, bc *Blockchain) *Transaction {
	var inputs []TXIn
	var outputs []TXOut

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: Not enough funds.")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXIn{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	outputs = append(outputs, TXOut{amount, to})
	if acc > amount {
		outputs = append(outputs, TXOut{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.Hash()

	return &tx

}
