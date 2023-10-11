package main

import (
	"fmt"
	"strconv"

	"github.com/argonautts/golang-blockchain/blockchain"
)

func main() {
	// Инициализация новой цепочки блоков.
	chain := blockchain.InitBlockChain()

	// Добавление нескольких блоков в цепочку.
	chain.AddBlock("First Block after Genesis")
	chain.AddBlock("Second Block after Genesis")
	chain.AddBlock("Third Block after Genesis")

	// Вывод информации о блоках в цикле.
	for _, block := range chain.Blocks {
		fmt.Printf("Previous Hash: %x\n", block.PrevHash)
		fmt.Printf("Data in Block: %s\n", block.Data)
		fmt.Printf("Hash: %x\n", block.Hash)
		fmt.Println("----------------------")

		pow := blockchain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
