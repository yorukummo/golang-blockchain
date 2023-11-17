package blockchain

import "github.com/dgraph-io/badger"

// BlockChainIterator используется для итерации по блокам цепочки.
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

// Iterator возвращает итератор для блоков цепочки.
func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

// Next возвращает следующий блок из цепочки и перемещает итератор.
func (iter *BlockChainIterator) Next() *Block {
	var block *Block

	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		encodedBlock, err := item.Value()
		block = Deserialize(encodedBlock)

		return err
	})
	Handle(err)

	iter.CurrentHash = block.PrevHash

	return block
}
