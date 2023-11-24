package blockchain

import (
	"bytes"
	"encoding/hex"
	"log"

	"github.com/dgraph-io/badger"
)

var (
	utxoPrefix   = []byte("utxo-") // Prefix for UTXO keys in the database
	prefixLength = len(utxoPrefix)
)

// UTXOSet represents the set of unspent transaction outputs (UTXOs) of a blockchain.
type UTXOSet struct {
	Blockchain *BlockChain // Reference to the blockchain to which the UTXO set belongs
}

// FindSpendableOutputs finds and returns unspent outputs to meet a given amount for a public key hash.
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOuts := make(map[string][]int) // Map for storing unspent outputs
	accumulated := 0                      // Total amount accumulated
	db := u.Blockchain.Database           // Database reference

	// Reading from the database
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts) // Creating a new iterator
		defer it.Close()

		// Iterating over UTXO set
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			k := item.Key()
			v, err := item.Value() // Getting the UTXO set value
			Handle(err)
			k = bytes.TrimPrefix(k, utxoPrefix) // Removing the prefix
			txID := hex.EncodeToString(k)       // Transaction ID
			outs := DeserializeOutputs(v)       // Deserializing outputs

			// Checking each output
			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOuts[txID] = append(unspentOuts[txID], outIdx) // Adding unspent output
				}
			}
		}
		return nil
	})
	Handle(err)

	return accumulated, unspentOuts // Returning the accumulated amount and unspent outputs
}

// FindUnspentTransactions finds all unspent transaction outputs for a given public key hash.
func (u UTXOSet) FindUnspentTransactions(pubKeyHash []byte) []TxOutput {
	var UTXOs []TxOutput

	db := u.Blockchain.Database

	// Reading from the database
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts) // Creating a new iterator
		defer it.Close()

		// Iterating over UTXO set
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			item := it.Item()
			v, err := item.Value() // Getting the UTXO set value
			Handle(err)
			outs := DeserializeOutputs(v) // Deserializing outputs

			// Adding unspent outputs
			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}
		return nil
	})
	Handle(err)

	return UTXOs // Returning all unspent transaction outputs
}

// CountTransactions counts the number of transactions in the UTXO set.
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.Database
	counter := 0 // Counter for transactions

	// Reading from the database
	err := db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions

		it := txn.NewIterator(opts) // Creating a new iterator
		defer it.Close()
		// Counting transactions in the UTXO set
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next() {
			counter++
		}

		return nil
	})

	Handle(err)

	return counter // Returning the count of transactions
}

// Reindex rebuilds the UTXO set from the blockchain transactions.
func (u UTXOSet) Reindex() {
	db := u.Blockchain.Database

	// Delete all UTXOs from the database before rebuilding
	u.DeleteByPrefix(utxoPrefix)

	UTXO := u.Blockchain.FindUTXO()

	// Update the database with all unspent transaction outputs
	err := db.Update(func(txn *badger.Txn) error {
		for txId, outs := range UTXO {
			key, err := hex.DecodeString(txId)
			Handle(err)
			key = append(utxoPrefix, key...)

			err = txn.Set(key, outs.Serialize())
			Handle(err)
		}

		return nil
	})
	Handle(err)
}

// Update updates the UTXO set with transactions from a new block.
func (u *UTXOSet) Update(block *Block) {
	db := u.Blockchain.Database

	// Update the UTXO set with each transaction in the block
	err := db.Update(func(txn *badger.Txn) error {
		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					updatedOuts := TxOutputs{}
					inID := append(utxoPrefix, in.ID...)
					item, err := txn.Get(inID)
					Handle(err)
					v, err := item.Value()
					Handle(err)

					outs := DeserializeOutputs(v)

					// Remove spent outputs and update the database
					for outIdx, out := range outs.Outputs {
						if outIdx != in.Out {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						if err := txn.Delete(inID); err != nil {
							log.Panic(err)
						}
					} else {
						if err := txn.Set(inID, updatedOuts.Serialize()); err != nil {
							log.Panic(err)
						}
					}
				}
			}
			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			txID := append(utxoPrefix, tx.ID...)
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	Handle(err)
}

// DeleteByPrefix deletes all keys in the database with a given prefix.
func (u *UTXOSet) DeleteByPrefix(prefix []byte) {
	deleteKeys := func(keysForDelete [][]byte) error {
		// Internal function to delete keys in a database transaction
		if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {
			for _, key := range keysForDelete {
				if err := txn.Delete(key); err != nil {
					return err
				}
			}
			return nil
		}); err != nil {
			return err
		}
		return nil
	}

	collectSize := 100000 // Number of keys to collect before deleting in batch
	u.Blockchain.Database.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0

		// Collect and delete keys with the specified prefix
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize {
				err := deleteKeys(keysForDelete)
				Handle(err)
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			err := deleteKeys(keysForDelete)
			Handle(err)
		}
		return nil
	})
}
