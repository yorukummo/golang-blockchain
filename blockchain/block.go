package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

// Определение структуры Block, которая представляет блок в цепочке.
type Block struct {
	Hash     []byte // Хэш блока
	Data     []byte // Данные, хранящиеся в блоке
	PrevHash []byte // Хэш предыдущего блока
	Nonce    int
}

// Функция CreateBlock создает новый блок с заданными данными и хэшом предыдущего блока.
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0}
	pow := NewProof(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

// Функция Genesis создает первый блок (генезис-блок) без данных и без предыдущего хэша.
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)

	return res.Bytes()
}

func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	Handle(err)

	return &block
}

func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
