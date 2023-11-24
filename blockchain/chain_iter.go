package blockchain

import "github.com/dgraph-io/badger"

// BlockChainIterator is used to iterate over the blockchain blocks.
type BlockChainIterator struct {
	CurrentHash []byte     // The hash of the current block being examined
	Database    *badger.DB // The database where the blockchain is stored
}

// Iterator creates and returns an iterator to traverse the blockchain starting from the last block.
func (chain *BlockChain) Iterator() *BlockChainIterator {
	// Initializing the iterator with the last block hash and the database
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

// Next moves the iterator to the next block in the blockchain and returns it.
func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	// Accessing the block from the database using the current hash
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedBlock, err := item.Value() // Retrieving the encoded block data
		block = Deserialize(encodedBlock) // Deserializing the block

		return err
	})
	Handle(err)

	// Moving the iterator to the previous block
	iter.CurrentHash = block.PrevHash

	return block // Returning the deserialized block
}
