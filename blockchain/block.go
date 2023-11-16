package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

// Block представляет собой структуру блока в блокчейне.
type Block struct {
	Hash        []byte         // Хэш текущего блока.
	Transaction []*Transaction // Транзакции, включенные в блок.
	PrevHash    []byte         // Хэш предыдущего блока.
	Nonce       int            // "Nonce" используется в доказательстве работы (proof-of-work).
}

// HashTransactions создает хеш всех транзакций в блоке.
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte

	for _, tx := range b.Transaction {
		txHashes = append(txHashes, tx.Serialize())
	}
	tree := NewMerkleTree(txHashes) // Используем дерево Меркля

	return tree.RootNode.Data
}

// CreateBlock создает и возвращает новый блок, содержащий заданные транзакции и ссылку на предыдущий блок.
func CreateBlock(txs []*Transaction, prevHash []byte) *Block {
	block := &Block{[]byte{}, txs, prevHash, 0}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// Genesis создает и возвращает генезис-блок, который содержит стартовую транзакцию.
func Genesis(coinbase *Transaction) *Block {
	return CreateBlock([]*Transaction{coinbase}, []byte{})
}

// Serialize преобразует блок в байтовый массив для сохранения или передачи.
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)

	return res.Bytes()
}

// Deserialize преобразует байтовый массив обратно в структуру блока.
func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	Handle(err)

	return &block
}

// Handle является обобщенным обработчиком ошибок, который завершает программу в случае ошибки.
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
