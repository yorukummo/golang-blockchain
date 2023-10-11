package blockchain

// Определение структуры BlockChain, которая представляет собой цепочку блоков.
type BlockChain struct {
	Blocks []*Block
}

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

// Метод AddBlock добавляет новый блок в цепочку, основываясь на предыдущем блоке.
func (chain *BlockChain) AddBlock(data string) {
	prevBlock := chain.Blocks[len(chain.Blocks)-1]
	new := CreateBlock(data, prevBlock.Hash)
	chain.Blocks = append(chain.Blocks, new)
}

// Функция Genesis создает первый блок (генезис-блок) без данных и без предыдущего хэша.
func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

// Функция InitBlockChain инициализирует новую цепочку блоков, начиная с генезис-блока.
func InitBlockChain() *BlockChain {
	return &BlockChain{[]*Block{Genesis()}}
}
