package main

import (
"bytes"
"crypto/ecdsa"
"encoding/hex"
"fmt"
"github.com/btcsuite/btcutil/tree/master/base58"
"log"

"github.com/boltdb/bolt"
)

const (
	BlocksBucket = "blocks"
	dbFile       = "chain.db"
)

type Blockchain struct {
	db *bolt.DB
	l  []byte
}

func CreateBlockchain(address string) *Blockchain {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var l []byte

	err = db.Update(func(tx *bolt.Tx) error {
		b, err := tx.CreateBucket([]byte(BlocksBucket))
		if err != nil {
			log.Panic(err)
		}

		genesis := NewBlock([]*Transaction{NewCoinbaseTX("", address)}, []byte{})

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})

	return &Blockchain{db, l}
}

func NewBlockchain() *Blockchain {
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	var l []byte

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		l = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return &Blockchain{db, l}
}

func (bc *Blockchain) AddBlock(transactions []*Transaction) {
	for _, tx := range transactions {
		if isVerified := bc.VerifyTransaction(tx); !isVerified {
			log.Panic("ERROR: Invalid transaction")
		}
	}

	block := NewBlock(transactions, bc.l)

	err := bc.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(BlocksBucket))

		err := b.Put(block.Hash, block.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), block.Hash)
		if err != nil {
			log.Panic(err)
		}

		bc.l = block.Hash

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

func (bc *Blockchain) FindTransaction(txid []byte) *Transaction {
	bci := NewBlockchainIterator(bc)
	for bci.HasNext() {
		block := bci.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, txid) == 0 {
				return tx
			}
		}
	}

	return nil
}

func (bc *Blockchain) SignTransaction(privKey *ecdsa.PrivateKey, tx *Transaction) {
	prevTXs := make(map[string]*Transaction)

	for _, in := range tx.Vin {
		prevTXs[hex.EncodeToString(in.Txid)] = bc.FindTransaction(in.Txid)
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *Blockchain) VerifyTransaction(tx *Transaction) bool {
	prevTXs := make(map[string]*Transaction)

	for _, in := range tx.Vin {
		prevTXs[hex.EncodeToString(in.Txid)] = bc.FindTransaction(in.Txid)
	}

	return tx.Verify(prevTXs)
}

func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []*Transaction {
	bci := NewBlockchainIterator(bc)

	spentTXOs := make(map[string][]int)
	var unspentTXs []*Transaction

	for bci.HasNext() {
		for _, tx := range bci.Next().Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Vout {
				// TXOutput 에서 이미 소비된 트랜잭션에 대해서는 처리하지 않는다.
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				// 그 이외의 트랜잭션은 아직 소비되지 않은 트랜잭션이다.
				if bytes.Compare(out.PubKeyHash, pubKeyHash) == 0 {
					unspentTXs = append(unspentTXs, tx)
				}
			}

			// 입력이 없는 코인베이스 트랜잭션은 제외.
			if !tx.IsCoinbase() {
				// TXInput 을 조사하여 이미 소비된 출력 집합을 얻는다.
				for _, in := range tx.Vin {
					// 입력의 공개키에 address 가 했음은 address 가 지불을 위해
					// 해당 트랜잭션 출력을 사용했다는 뜻이다.
					if in.UsesKey(pubKeyHash) {
						hash := hex.EncodeToString(in.Txid)
						spentTXOs[hash] = append(spentTXOs[hash], in.Vout)
					}
				}
			}
		}
	}

	return unspentTXs
}

func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []TXOutput {
	var UTXOs []TXOutput
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTXs {
		for _, out := range tx.Vout {
			if bytes.Compare(out.PubKeyHash, pubKeyHash) == 0 {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) List() {
	bci := NewBlockchainIterator(bc)

	for bci.HasNext() {
		block := bci.Next()

		fmt.Printf("PrevBlockHash: %x\n", block.PrevBlockHash)
		fmt.Printf("Hash: %x\n", block.Hash)

		pow := NewProofOfWork(block)
		fmt.Println("pow:", pow.Validate(block))

		fmt.Println()
	}
}

func (bc *Blockchain) GetBalance(address string) uint64 {
	var balance uint64

	pubKeyHash, _, err := base58.CheckDecode(address)
	if err != nil {
		log.Panic(err)
	}
	for _, out := range bc.FindUTXO(pubKeyHash) {
		balance += out.Value
	}

	return balance
}

func (bc *Blockchain) Send(value uint64, from, to string) *Transaction {
	var txin []TXInput
	var txout []TXOutput

	keyStore := NewKeyStore()

	wallet := keyStore.Wallets[from]
	UTXs := bc.FindUnspentTransactions(HashPubKey(wallet.PubKey))

	var acc uint64

Work:
	for _, tx := range UTXs {
		for outIdx, out := range tx.Vout {
			if bytes.Compare(out.PubKeyHash, HashPubKey(wallet.PubKey)) == 0 && acc < value {
				acc += out.Value
				txin = append(txin, TXInput{tx.ID, outIdx, nil, wallet.PubKey})
			}
			if acc >= value {
				break Work
			}
		}
	}

	if value > acc {
		log.Panic("ERROR: NOT enough funds")
	}

	txout = append(txout, *NewTXOutput(value, to))
	if acc > value {
		txout = append(txout, *NewTXOutput(acc-value, from))
	}

	tx := NewTransaction(txin, txout)
	bc.SignTransaction(wallet.PrivKey, tx)

	return tx
}