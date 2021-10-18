package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

const targetBits = 16

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, 256-targetBits)

	return &ProofOfWork{b, target}
}

func (pow *ProofOfWork) prepareData(nonce int64) []byte {
	data := bytes.Join([][]byte{
		pow.block.PrevBlockHash,
		pow.block.HashTransaction(),
		IntToHex(pow.block.Timestamp),
		IntToHex(nonce),
		IntToHex(targetBits),
	}, []byte{})

	return data
}

func (pow *ProofOfWork) Run() (int64, []byte) {
	var nonce int64

	var hashInt big.Int
	var hash [32]byte

	for nonce < math.MaxInt64 {
		data := pow.prepareData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash)

		hashInt.SetBytes(hash[:])
		if hashInt.Cmp(pow.target) == -1 {
			break
		}
		nonce++
	}

	return nonce, hash[:]
}

func (pow *ProofOfWork) Validate(b *Block) bool {
	var hashInt big.Int
	data := pow.prepareData(b.Nonce)
	hash := sha256.Sum256(data)

	hashInt.SetBytes(hash[:])

	return hashInt.Cmp(pow.target) == -1
}