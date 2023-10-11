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

// Статическая сложность, в наст блокчейне со временем сложность наростает
const Difficulty = 18

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}

// Инициализация данных для выполнения работы по доказательству работы.
func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,       // Хэш предыдущего блока
			pow.Block.Data,           // Данные текущего блока
			ToHex(int64(nonce)),      // Счетчик (nonce) в виде байтов
			ToHex(int64(Difficulty)), // Сложность в виде байтов
		},
		[]byte{},
	)
	return data
}

// Метод Run выполняет майнинг для поиска правильного значения nonce и соответствующего хэша.
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash) // Вывод текущего хэша в процессе майнинга
		intHash.SetBytes(hash[:])

		if intHash.Cmp(pow.Target) == -1 {
			break // Если хэш соответствует требованиям, завершаем майнинг
		} else {
			nonce++ // Увеличиваем nonce и продолжаем поиск
		}
	}
	fmt.Println()

	return nonce, hash[:]
}

// Метод Validate проверяет, соответствует ли хэш блока текущей сложности (Difficulty)
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce) // Используем текущее значение nonce блока

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

// Функция ToHex преобразует целое число в байты для представления в шестнадцатеричной форме
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
