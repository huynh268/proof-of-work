package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
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

	txin := TXIn{[]byte{}, -1, nil, []byte(data)}
	txout := CreateTXOutput(firstReward, toAddress)
	tx := Transaction{nil, []TXIn{txin}, []TXOut{*txout}}
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

// func (in *TXIn) CanUnlockOutputWith(unlockingData string) bool {
// 	return in.ScriptSig == unlockingData
// }
//
// func (out *TXOut) CanBeUnlockedWith(unlockingData string) bool {
// 	return out.ScriptPubKey == unlockingData
// }

func (tx Transaction) IsCoinbase() bool {
	return len(tx.Vin) == 1 && len(tx.Vin[0].TxID) == 0 && tx.Vin[0].Vout == -1
}

func NewUTXOTransaction(wallet *Wallet, to string, amount int, UTXOSet *UTXOSet) *Transaction {
	var inputs []TXIn
	var outputs []TXOut

	pubKeyHash := HashPublicKey(wallet.PublicKey)
	acc, validOutputs := UTXOSet.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("Error: Not enough funds.")
	}

	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := TXIn{txID, out, nil, wallet.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := fmt.Sprintf("%s", wallet.GetAddress())
	outputs = append(outputs, *CreateTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *CreateTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()
	UTXOSet.Blockchain.SignTransaction(&tx, wallet.PrivateKey)

	return &tx

}

func (tx *Transaction) Sign(privateKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inputID, vin := range txCopy.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.TxID)]
		txCopy.Vin[inputID].Signature = nil
		txCopy.Vin[inputID].PublicKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inputID].PublicKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privateKey, txCopy.ID)

		signature := append(r.Bytes(), s.Bytes()...)

		tx.Vin[inputID].Signature = signature
	}
}

func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inputID, vin := range tx.Vin {
		prevTX := prevTXs[hex.EncodeToString(vin.TxID)]
		txCopy.Vin[inputID].Signature = nil
		txCopy.Vin[inputID].PublicKey = prevTX.Vout[vin.Vout].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.Vin[inputID].PublicKey = nil

		r := big.Int{}
		s := big.Int{}

		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PublicKey)
		x.SetBytes(vin.PublicKey[:(keyLen / 2)])
		y.SetBytes(vin.PublicKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{curve, &x, &y}
		if !ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) {
			return false
		}
	}

	return true
}

func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXIn
	var outputs []TXOut

	for _, vin := range tx.Vin {
		inputs = append(inputs, TXIn{vin.TxID, vin.Vout, nil, nil})
	}

	for _, vout := range tx.Vout {
		outputs = append(outputs, TXOut{vout.Value, vout.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}
