package wallet

import (
	"log"

	"github.com/mr-tron/base58"
)

// Base58Encode принимает байтовый срез в качестве входных данных и возвращает его кодировку в формате Base58.
func Base58Encode(input []byte) []byte {
	// Кодирование входного байтового среза в строку Base58
	encode := base58.Encode(input)

	// Возвращаем результат кодирования в виде байтового среза
	return []byte(encode)
}

// Base58Decode принимает закодированный в формате Base58 байтовый срез и возвращает его декодировку.
func Base58Decode(input []byte) []byte {
	// Декодирование входного байтового среза
	decode, err := base58.Decode(string(input[:]))
	// Если при декодировании возникла ошибка, паникуем и записываем в лог
	if err != nil {
		log.Panic(err)
	}

	// Возвращаем результат декодирования
	return decode
}

// nety 0 O l I + /
