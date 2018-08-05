package main

// TXIn transaction input structure
type TXIn struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}
