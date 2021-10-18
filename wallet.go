package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"github.com/btcsuite/btcutil/tree/master/base58"
	"golang.org/x/crypto/ripemd160"
	"log"
)

type Wallet struct {
	PrivKey *ecdsa.PrivateKey
	PubKey  []byte
}

func NewWallet() *Wallet {
	curve := elliptic.P256()
	privKey, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Panic(err)
	}

	pubKey := append(privKey.PublicKey.X.Bytes(), privKey.PublicKey.Y.Bytes()...)

	return &Wallet{privKey, pubKey}
}

func (w *Wallet) GetAddress() string {
	publicRIPEMD160 := HashPubKey(w.PubKey)
	version := byte(0x00)

	return base58.CheckEncode(publicRIPEMD160, version)
}

func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)

	RIPEMD160Hasher := ripemd160.New()
	_, err := RIPEMD160Hasher.Write(publicSHA256[:])
	if err != nil {
		log.Panic(err)
	}

	return RIPEMD160Hasher.Sum(nil)
}