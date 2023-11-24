// Package blockchain represents the core logic for blockchain operations such as managing blocks,
// transactions, and their interrelationships like merkle trees and proof of work.
package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

const (
	dbPath      = "./tmp/blocks_%s"                // Path for storing blockchain data
	genesisData = "First transaction from Genesis" // Initial data for the genesis block
)

// BlockChain represents a blockchain with a pointer to the last block in the chain and the database.
type BlockChain struct {
	LastHash []byte     // Hash of the last block in the chain
	Database *badger.DB // Database to store the blockchain data
}

// DBexists checks if a blockchain database exists at a given path.
func DBexists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false // Database does not exist
	}

	return true // Database exists
}

// ContinueBlockChain returns an existing blockchain from the database.
func ContinueBlockChain(nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBexists(path) == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit() // Exiting if no blockchain found
	}

	var lastHash []byte

	// Setting up badger database options
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	// Opening the database
	db, err := openDB(path, opts)
	Handle(err)

	// Retrieving the last hash from the database
	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()

		return err
	})
	Handle(err)

	chain := BlockChain{lastHash, db}

	return &chain // Returning the existing blockchain
}

// InitBlockChain initializes a new blockchain with a genesis block.
func InitBlockChain(address, nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBexists(path) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit() // Exiting if blockchain already exists
	}
	var lastHash []byte

	// Setting up badger database options
	opts := badger.DefaultOptions
	opts.Dir = path
	opts.ValueDir = path

	// Opening the database
	db, err := openDB(path, opts)
	Handle(err)

	// Creating and storing the genesis block in the database
	err = db.Update(func(txn *badger.Txn) error {
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err
	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain // Returning the new blockchain
}

// AddBlock adds a new block to the blockchain.
func (chain *BlockChain) AddBlock(block *Block) {
	err := chain.Database.Update(func(txn *badger.Txn) error {
		// Checking if the block already exists in the database
		if _, err := txn.Get(block.Hash); err == nil {
			return nil // Block already exists, no need to add
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData) // Storing the serialized block in the database
		Handle(err)

		// Updating the last hash in the database if this block is the latest
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			Handle(err)
			chain.LastHash = block.Hash // Updating the last hash in the blockchain
		}

		return nil
	})
	Handle(err)
}

// GetBestHeight returns the height of the latest block in the blockchain.
func (chain *BlockChain) GetBestHeight() int {
	var lastBlock Block

	// Reading the last block from the database
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ := item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock = *Deserialize(lastBlockData)

		return nil
	})
	Handle(err)

	return lastBlock.Height
}

// GetBlock retrieves a block by its hash from the blockchain.
func (chain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block

	// Reading the block data from the database
	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")
		} else {
			blockData, _ := item.Value()

			block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil // Returning the found block
}

// GetBlockHashes returns the hashes of all the blocks in the blockchain.
func (chain *BlockChain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator() // Getting an iterator to go through the blocks

	// Iterating through all blocks in the blockchain
	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash) // Adding the hash of each block to the slice

		// Break if the genesis block is reached
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

// MineBlock mines a new block with the given transactions.
func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastHeight int

	// Verifying each transaction before adding it to the block
	for _, tx := range transactions {
		if chain.VerifyTransaction(tx) != true {
			log.Panic("Invalid transaction")
		}
	}

	// Retrieving the last block's hash and height
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.Value()

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.Value()

		lastBlock := Deserialize(lastBlockData)

		lastHeight = lastBlock.Height

		return err
	})
	Handle(err)

	// Creating and adding the new block to the chain
	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)

	return newBlock
}

// FindUTXO finds and returns all unspent transaction outputs (UTXOs).
func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator() // Getting an iterator to go through the blocks

	// Iterating through all blocks in the blockchain
	for {
		block := iter.Next()

		// Iterating through each transaction in the block
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.Outputs {
				// Checking if the output was spent
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs
			}
			// Marking inputs as spent
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}

		// Break if the genesis block is reached
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return UTXO
}

// FindTransaction finds a transaction by its ID.
func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator() // Getting an iterator to go through the blocks

	// Iterating through all blocks in the blockchain
	for {
		block := iter.Next()

		// Searching for the transaction in each block
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		// Break if the genesis block is reached
		if len(block.PrevHash) == 0 {
			break
		}
	}

	return Transaction{}, errors.New("Transaction does not exist")
}

// SignTransaction signs a transaction using a given private key.
func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	// Retrieving all previous transactions referred in the inputs
	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs) // Signing the transaction
}

// VerifyTransaction verifies a transaction's inputs.
func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	// Coinbase transactions do not require verification
	if tx.IsCoinbase() {
		return true
	}
	prevTXs := make(map[string]Transaction)

	// Retrieving all previous transactions referred in the inputs
	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs) // Verifying the transaction
}

// retry attempts to open a database if it is locked by removing the lock file.
func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	// Attempting to remove the lock file
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removal "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

// openDB attempts to open a badger database and retries if it is locked.
func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		// Retry opening the database if it is locked
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database is unblocked, the log of values is truncated")
				return db, nil
			}
			log.Println("database could not be unblocked:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}
