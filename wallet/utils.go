package wallet

import (
	"log"

	"github.com/mr-tron/base58"
)

// Base58Encode encodes a byte slice into a Base58 encoded byte slice.
// Base58 encoding is used for encoding addresses in Bitcoin and other cryptocurrencies.
// It is similar to Base64 but omits potentially ambiguous characters like 0 (zero), O (capital o), l (lowercase L), I (capital i).
func Base58Encode(input []byte) []byte {
	// Encoding the input using Base58
	encode := base58.Encode(input)

	// Returning the encoded data as a byte slice
	return []byte(encode)
}

// Base58Decode decodes a Base58 encoded byte slice back to its original byte slice.
// This function is used to decode data encoded in Base58 format, commonly used in various cryptocurrency wallets.
func Base58Decode(input []byte) []byte {
	// Decoding the input from Base58 format
	decode, err := base58.Decode(string(input[:]))
	if err != nil {
		// Logging and halting on error
		log.Panic(err)
	}

	// Returning the decoded data
	return decode
}

// The characters '0', 'O', 'l', 'I', '+' and '/' are omitted from the Base58 encoding alphabet
// to avoid ambiguity and improve readability.
