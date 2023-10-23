package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

// Plan:
// ----------------------------------------------
// Извлечь данные из блока

// Создайте счетчик (nonce), который начинается с 0

// Создать хэш данных + счетчик

// Проверить хэш, чтобы увидеть, соответствует ли он набору требований

// Требования:
// Первые несколько байтов должны содержать 0s
// ----------------------------------------------

// Статическая сложность. В реальных блокчейнах сложность адаптируется со временем.
const Difficulty = 18

// ProofOfWork представляет собой структуру доказательства работы (Proof-of-Work) для блока.
type ProofOfWork struct {
	Block  *Block   // Ссылка на блок для которого ищется доказательство
	Target *big.Int // Цель (значение, ниже которого должен быть хэш)
}

// NewProof инициализирует новое доказательство работы для блока.
func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty)) // Установка цели по заданной сложности

	pow := &ProofOfWork{b, target}

	return pow
}

// InitData инициализирует данные для майнинга блока.
func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,           // Хэш предыдущего блока
			pow.Block.HashTransactions(), // Данные транзакций текущего блока
			ToHex(int64(nonce)),          // Текущее значение счетчика (nonce)
			ToHex(int64(Difficulty)),     // Текущая сложность в форме байтов
		},
		[]byte{},
	)
	return data
}

// Run выполняет алгоритм доказательства работы, ищет nonce и соответствующий блоку хэш.
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash) // Отображение текущего хэша в процессе майнинга
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break // Если найденный хэш соответствует условиям, завершаем майнинг
		} else {
			nonce++ // Увеличиваем nonce и продолжаем искать
		}
	}
	fmt.Println()

	return nonce, hash[:]
}

// Validate проверяет, соответствует ли хэш блока условиям сложности.
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce) // Получаем данные блока с текущим nonce

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1 // Проверка соответствия условиям сложности
}

// ToHex преобразует целое число в его шестнадцатеричное представление в виде байтов.
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
