package main

type TXIn struct {
	Txid      []byte
	Vout      int
	ScriptSig string
}
