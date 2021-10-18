package main

import (
"bytes"
"crypto/sha256"
"encoding/gob"
"log"
"time"
)

type Block struct {
	PrevBlockHash []byte
	Hash          []byte
	Timestamp     int64
	Transactions  []*Transaction
	Nonce         int64
}

func NewBlock(transactions []*Transaction, prevBlockHash []byte) *Block {
	block := &Block{prevBlockHash, []byte{}, time.Now().Unix(), transactions, 0}
	pow := NewProofOfWork(block)
	block.Nonce, block.Hash = pow.Run()

	return block
}

func (b *Block) HashTransaction() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}

	txHash := sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

func (b *Block) Serialize() []byte {
	var buf bytes.Buffer

	encoder := gob.NewEncoder(&buf)

	err := encoder.Encode(b)
	if err != nil {
		log.Panic(err)
	}

	return buf.Bytes()
}

func DeserializeBlock(encodedBlock []byte) *Block {
	var buf bytes.Buffer
	var block Block

	buf.Write(encodedBlock)
	decoder := gob.NewDecoder(&buf)

	err := decoder.Decode(&block)
	if err != nil {
		log.Panic(err)
	}

	return &block
}