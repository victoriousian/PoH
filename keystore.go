package main

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"io/ioutil"
	"log"
	"os"
)

const walletFile = "wallet.dat"

type KeyStore struct {
	Wallets map[string]*Wallet
}

func createKeyStore() error {
	file, err := os.OpenFile(walletFile, os.O_CREATE, 0644)
	if err != nil {
		return err
	}

	return file.Close()
}

func NewKeyStore() *KeyStore {
	keyStore := KeyStore{make(map[string]*Wallet)}

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		err := createKeyStore()
		if err != nil {
			log.Panic(err)
		}
	} else {
		fileContent, err := ioutil.ReadFile(walletFile)
		if err != nil {
			log.Panic(err)
		}

		gob.Register(elliptic.P256())

		decoder := gob.NewDecoder(bytes.NewReader(fileContent))
		err = decoder.Decode(&keyStore)
		if err != nil {
			log.Panic(err)
		}
	}

	return &keyStore
}

func (ks *KeyStore) Save() {
	buf := new(bytes.Buffer)

	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(ks)
	if err != nil {
		log.Panic(err)
	}

	err = ioutil.WriteFile(walletFile, buf.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

func (ks *KeyStore) CreateWallet() *Wallet {
	wallet := NewWallet()

	ks.Wallets[wallet.GetAddress()] = wallet
	ks.Save()

	return wallet
}