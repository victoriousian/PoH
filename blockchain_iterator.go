package main

import (
	"bytes"
	"github.com/boltdb/bolt"
	"log"
)

type blockchainIterator struct {
	db   *bolt.DB
	hash []byte
}

func NewBlockchainIterator(bc *Blockchain) *blockchainIterator {
	return &blockchainIterator{bc.db, bc.l}
}

func (i *blockchainIterator) Next() *Block {
	var block *Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		encodedBlock := b.Get(i.hash)
		block = DeserializeBlock(encodedBlock)

		i.hash = block.PrevBlockHash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return block
}

func (i *blockchainIterator) HasNext() bool {
	return bytes.Compare(i.hash, []byte{}) != 0
}