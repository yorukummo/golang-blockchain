package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

// Block represents a single block in the blockchain.
type Block struct {
	Timestamp    int64          // Timestamp of block creation
	Hash         []byte         // Hash of the block
	Transactions []*Transaction // Transactions included in the block
	PrevHash     []byte         // Hash of the previous block in the chain
	Nonce        int            // Nonce used for mining (Proof of Work)
	Height       int            // Height of the block in the blockchain
}

// HashTransactions creates a hash of all the transactions in the block using a Merkle Tree.
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize()) // Serializing each transaction
	}
	tree := NewMerkleTree(txHashes) // Creating a new Merkle Tree from the transaction hashes

	return tree.RootNode.Data // Returning the root hash of the Merkle Tree
}

// CreateBlock creates a new block with the given transactions and previous hash.
func CreateBlock(txs []*Transaction, prevHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), []byte{}, txs, prevHash, 0, height}
	pow := NewProof(block)   // Creating a new proof of work for the block
	nonce, hash := pow.Run() // Running the proof of work algorithm to mine the block

	block.Hash = hash[:] // Setting the hash of the block
	block.Nonce = nonce  // Setting the nonce of the block

	return block
}

// Genesis creates the first block in the blockchain with a coinbase transaction.
func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{}, 0) // Creating the genesis block
}

// Serialize encodes the block into a byte slice.
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res) // Creating a new encoder

	err := encoder.Encode(b) // Encoding the block
	Handle(err)              // Handling any encoding errors

	return res.Bytes() // Returning the encoded byte slice
}

// Deserialize decodes a byte slice into a Block.
func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data)) // Creating a new decoder

	err := decoder.Decode(&block) // Decoding the data into a block
	Handle(err)                   // Handling any decoding errors

	return &block // Returning the decoded block
}

// Handle is a utility function for error handling.
func Handle(err error) {
	if err != nil {
		log.Panic(err) // Logging and panicking on error
	}
}
