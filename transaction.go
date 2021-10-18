package main

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"log"
	"math/big"
)

const subsidy = 10 // BTC

type Transaction struct {
	ID   []byte
	Vin  []TXInput
	Vout []TXOutput
}

func NewTransaction(vin []TXInput, vout []TXOutput) *Transaction {
	tx := Transaction{nil, vin, vout}
	tx.SetID()

	return &tx
}

func (tx *Transaction) SetID() {
	buf := new(bytes.Buffer)

	encoder := gob.NewEncoder(buf)
	err := encoder.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	hash := sha256.Sum256(buf.Bytes())
	tx.ID = hash[:]
}

func NewCoinbaseTX(data, to string) *Transaction {
	txin := TXInput{[]byte{}, -1, nil, []byte(data)}
	txout := NewTXOutput(subsidy, to)

	return NewTransaction([]TXInput{txin}, []TXOutput{*txout})
}

func (tx *Transaction) IsCoinbase() bool {
	return bytes.Compare(tx.Vin[0].Txid, []byte{}) == 0 && tx.Vin[0].Vout == -1 && len(tx.Vin) == 1
}

func (tx *Transaction) Sign(privKey *ecdsa.PrivateKey, prevTXs map[string]*Transaction) {
	if tx.IsCoinbase() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inID, in := range txCopy.Vin {
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTXs[hex.EncodeToString(in.Txid)].Vout[in.Vout].PubKeyHash
		txCopy.SetID()
		txCopy.Vin[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, privKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}

		tx.Vin[inID].Signature = append(r.Bytes(), s.Bytes()...)
	}
}

func (tx *Transaction) TrimmedCopy() *Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, in := range tx.Vin {
		inputs = append(inputs, TXInput{in.Txid, in.Vout, nil, nil})
	}
	for _, out := range tx.Vout {
		outputs = append(outputs, TXOutput{out.Value, out.PubKeyHash})
	}

	return &Transaction{nil, inputs, outputs}
}

func (tx *Transaction) Verify(prevTXs map[string]*Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, in := range tx.Vin {
		txCopy.Vin[inID].Signature = nil
		txCopy.Vin[inID].PubKey = prevTXs[hex.EncodeToString(in.Txid)].Vout[in.Vout].PubKeyHash
		txCopy.SetID()
		txCopy.Vin[inID].PubKey = nil

		var r, s big.Int

		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:sigLen/2])
		s.SetBytes(in.Signature[sigLen/2:])

		var x, y big.Int

		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:keyLen/2])
		y.SetBytes(in.PubKey[keyLen/2:])

		pubKey := ecdsa.PublicKey{curve, &x, &y}

		if isVerified := ecdsa.Verify(&pubKey, txCopy.ID, &r, &s); !isVerified {
			return false
		}
	}

	return true
}