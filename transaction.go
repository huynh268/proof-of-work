package main

import "fmt"

const subsidy = 10

type Transaction struct {
	ID   []byte
	Vin  []TXIn
	Vout []TXOut
}

func CreateCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to %s", to)
	}

	txin := TXIn{[]byte{}, -1, data}
	txout := TXOut{subsidy, to}
	tx := Transaction{nil, []TXIn{txin}, []TXOut{txout}}
	//tx.ID = tx.Hash()

	return &tx
}
