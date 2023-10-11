package main

import (
	"bytes"
	"crypto/sha256"
	"fmt"
)

// Определение структуры BlockChain, которая представляет собой цепочку блоков.
type BlockChain struct {
	blocks []*Block
}

// Определение структуры Block, которая представляет блок в цепочке.
type Block struct {
	Hash     []byte // Хэш блока
	Data     []byte // Данные, хранящиеся в блоке
	PrevHash []byte // Хэш предыдущего блока
}

// Метод DeriveHash вычисляет хэш текущего блока на основе его данных и хэша предыдущего блока.
func (b *Block) DeriveHash() {
	info := bytes.Join([][]byte{b.Data, b.PrevHash}, []byte{})
	hash := sha256.Sum256(info)
	b.Hash = hash[:]
}

// Функция CreateBlock создает новый блок с заданными данными и хэшом предыдущего блока.
func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash}
	block.DeriveHash()
	return block
}

// Метод AddBlock добавляет новый блок в цепочку, основываясь на предыдущем блоке.
func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.blocks[len(chain.blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.blocks = append(chain.blocks, new)
}

// Функция Genesis создает первый блок (генезис-блок) без данных и без предыдущего хэша.
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

// Функция InitBlockChain инициализирует новую цепочку блоков, начиная с генезис-блока.
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}

func main() {
	// Инициализация новой цепочки блоков.
	chain := InitBlockChain()

	// Добавление нескольких блоков в цепочку.
	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	// Вывод информации о блоках в цикле.
	for _, block := range chain.blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println("----------------------")
	}
}
