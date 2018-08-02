package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

var (
	maxNonce = math.MaxInt64
)

const difficulty = 2 << 3

// PoW structure
type PoW struct {
	block  *Block
	target *big.Int
}

// CreatePoW creates a new PoW
func CreatePoW(block *Block) *PoW {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-difficulty))
	pow := &PoW{block, target}

	return pow
}

// prepareData prepares data for mining
func (pow *PoW) prepareData(nonce int) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevHash,
		pow.block.Data,
		IntToByte(pow.block.Timestamp),
		IntToByte(int64(difficulty)),
		IntToByte(int64(nonce)),
	}, []byte{})

	return data
}

// Mine runs Proof of Work for sealing a new block
func (pow *PoW) Mine() (int, []byte) {
	var hashInt big.Int
	var hash [32]byte
	nonce := 0

	fmt.Printf("Data of the block being mined: \"%s\"\n", pow.block.Data)

	for nonce < maxNonce {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)
		fmt.Printf("\r%x", hash)
		hashInt.SetBytes(hash[:])

		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Print("\n\n")

	return nonce, hash[:]
}

// Validate Validate PoW
func (pow *PoW) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid
}
