package main

import (
	"github.com/btcsuite/btcutil/tree/master/base58"
	"log"
)

type TXOutput struct {
	Value      uint64
	PubKeyHash []byte
}

func NewTXOutput(value uint64, address string) *TXOutput {
	txo := &TXOutput{value, nil}
	txo.Lock(address)

	return txo
}

func (out *TXOutput) Lock(address string) {
	pubKeyHash, _, err := base58.CheckDecode(address)
	if err != nil {
		log.Panic(err)
	}
	out.PubKeyHash = pubKeyHash
}