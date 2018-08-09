package main

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
)

type addr struct {
	AddressList []string
}

type block struct {
	AddressFrom string
	Block       []byte
}

type getBlocks struct {
	AddressFrom string
}

type getData struct {
	AddressFrom string
	Type        string
	ID          []byte
}

type inv struct {
	AddressFrom string
	Type        string
	Items       [][]byte
}

type tx struct {
	AddressFrom string
	Transaction []byte
}

type version struct {
	Version    int
	BestHeight int
	AddFrom    string
}

const protocol = "tcp"
const nodeVersion = 1
const cmdLength = 12

var nodeAddress string
var miningAddress string
var knownNodes = []string{"localhost:3000"}
var blocksInTransit [][]byte
var mempool = make(map[string]Transaction)

func StartSever(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	miningAddress = minerAddress
	listen, err := net.Listen(protocol, nodeAddress)
	defer listen.Close()

	bc := CreateBlockchain(nodeID)

	if nodeAddress != knownNodes[0] {
		sendVersion(knownNodes[0], bc)
	}

	for {
		conn, err := listen.Accept()
		go handleConn(conn, bc)
	}
}

func handleConn(conn net.Conn, bc *Blockchain) {
	request, err := ioutil.ReadAll(conn)
	command := bytesToCommand(request[:cmdLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "address":
		handleAddress(request)
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request, bc)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTX(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	conn.Close()
}

func handleAddress(request []byte) {
	var buffer bytes.Buffer
	var payload addr

	buffer.Write(request[cmdLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	knownNodes = append(knownNodes, payload.AddressList...)
	fmt.Printf("There are %d known nodes now!\n", len(knownNodes))
	requestBlocks()
}

func handleBlock(request []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var payload block

	buffer.Write(request[cmdLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := DeserializeBlock(blockData)

	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddressFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := UTXOSet{bc, nil}
		UTXOSet.Reindex()
	}
}

func handleInv(request []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var payload inv

	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Received inventory with %d %s\n", len(payload.Items), payload.Type)

	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		sendGetData(payload.AddressFrom, "block", blockHash)

		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	if payload.Type == "tx" {
		txID := payload.Items[0]

		if mempool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddressFrom, "tx", txID)
		}
	}
}

func handleGetBlocks(request []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var payload getBlocks

	buffer.Write(request[cmdLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blocks := bc.GetBlockHashes()

	sendInv(payload.AddressFrom, "block", blocks)
}

func handleGetData(request []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var payload getData

	buffer.Write(request[cmdLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock([]byte(payload.ID))
		if err != nil {
			log.Panic(err)
		}

		sendBlock(payload.AddressFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := mempool[txID]

		sendTx(payload.AddressFrom, &tx)
	}

}

func handleTX(request []byte, bc *Blockchain) {
	var buffer bytes.Buffer
	var payload tx

	buffer.Write(request[cmdLength:])
	decoder := gob.NewDecoder(&buffer)
	err := decoder.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := DeserializeTransaction(txData)
	mempool[hex.EncodeToString(tx.ID)] = tx

	if nodeAddress == knownNodes[0] {
		for _, node := range knownNodes {
			if node != nodeAddress && node != payload.AddressFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(mempool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*Transaction

			for id := range mempool {
				tx := mempool[id]
				if bc.VerifyTransaction(&tx) {
					txs := append(txs, &tx)
				}
			}

			if len(txs) == 0 {
				fmt.Printf("All transaction are invalid! Waiting for new ones...")
				return
			}

			cbTx := CreateCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := UTXOSet{bc, nil}
			UTXOSet.Reindex()

			fmt.Printf("New block is minded!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(mempool, txID)
			}

			for _, node := range knownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(mempool) > 0 {
				goto MineTransactions
			}

		}

	}
}

func sendAddress(address string) {
	nodes := addr{knownNodes}
	nodes.AddressList = append(nodes.AddressList, nodeAddress)
	payload := gobEncode(nodes)
	request := append(commandToBytes("address"), payload...)

	sendData(address, request)
}

func sendBlock(address string, b *Block) {
	data := block{nodeAddress, b.SerializeBlock()}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(address, request)
}

func sendData(address string, data []byte) {
	conn, err := net.Dial(protocol, address)
	if err != nil {
		fmt.Printf("%s is not available\n", address)
		var updatedNodes []string

		for _, node := range knownNodes {
			if node != address {
				updatedNodes = append(updatedNodes, node)
			}
		}

		knownNodes = updatedNodes

		return
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

func sendInv(address, kind string, items [][]byte) {
	inventory := inv{address, kind, items}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

func sendGetBlocks(address string) {
	payload := gobEncode(getBlocks{nodeAddress})
	request := append(commandToBytes("getblocks"), payload...)

	sendData(address, request)
}

func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getData{nodeAddress, kind, id})
	request := append(commandToBytes("getData"), payload...)
}

func sendTx(address string, tnx *Transaction) {
	data := tx{nodeAddress, tnx.serialize()}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(address, request)
}

func sendVersion(address string, bc *Blockchain) {
	bestHeight := bc.GetBestHeight()
	payload := gobEncode(version{nodeVersion, bestHeight, nodeAddress})

	request := append(commandToBytes("version"), payload...)

	sendData(address, request)
}

func gobEncode(data interface{}) []byte {
	var buffer bytes.Buffer

	encoder := gob.NewEncoder(&buffer)
	err := encoder.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buffer.Bytes()
}

func commandToBytes(command string) []byte {
	var bytes [cmdLength]byte
	copy(bytes[:], command)
	return bytes[:]
}

func bytesToCommand(bytesCmd []byte) string {
	var buffer bytes.Buffer

	for i := 0; i < len(bytesCmd); i++ {
		if bytesCmd[i] != 0x0 {
			buffer.WriteByte(bytesCmd[i])
		}
	}

	return buffer.String()
}

func requestBlocks() {
	for _, node := range knownNodes {
		sendGetBlocks(node)
	}
}
