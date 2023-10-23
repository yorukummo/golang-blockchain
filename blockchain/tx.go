package blockchain

import (
	"bytes"

	"github.com/argonautts/golang-blockchain/wallet"
)

// TxOutput структура выходных данных транзакции.
type TxOutput struct {
	Value      int    // Значение в криптовалюте
	PubKeyHash []byte // Хэш публичного ключа получателя
}

// TxInput структура входных данных транзакции.
type TxInput struct {
	ID        []byte // Идентификатор транзакции, к которой относится этот вход
	Out       int    // Индекс выхода в транзакции, на который ссылается этот вход
	Signature []byte // Подпись владельца входа
	PubKey    []byte // Публичный ключ владельца входа
}

// NewTXOutput создает новый выход транзакции.
func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}

// UsesKey проверяет, использует ли входной объект транзакции данный публичный ключ.
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

// Lock связывает выход транзакции с адресом.
func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

// IsLockedWithKey проверяет, закрыт ли выход транзакции данным публичным ключом.
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}
