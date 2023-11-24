// Package blockchain represents the core logic for blockchain operations such as managing blocks,
// transactions, and their interrelationships like merkle trees and proof of work.
package blockchain

import (
	"bytes"
	"encoding/gob"
	"github.com/argonautts/golang-blockchain/wallet"
)

// TxOutput represents a transaction output.
type TxOutput struct {
	Value      int    // The value of coins in the output
	PubKeyHash []byte // The hash of the public key that can unlock this output
}

// TxOutputs holds multiple transaction outputs.
type TxOutputs struct {
	Outputs []TxOutput // Slice of outputs
}

// TxInput represents a transaction input.
type TxInput struct {
	ID        []byte // The ID of the transaction the output is in
	Out       int    // The index of the output in the transaction
	Signature []byte // The signature that unlocks the output
	PubKey    []byte // The public key corresponding to the address
}

// UsesKey checks whether the input uses a specific public key hash.
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey) // Extract the public key hash from the input's public key

	return bytes.Compare(lockingHash, pubKeyHash) == 0 // Compare with the provided public key hash
}

// Lock locks the output to a specific address.
func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)     // Decoding the address
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4] // Removing the version and checksum
	out.PubKeyHash = pubKeyHash                    // Setting the public key hash on the output
}

// IsLockedWithKey checks if the output is locked with a specific public key hash.
func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0 // Compare with the provided public key hash
}

// NewTXOutput creates a new transaction output locked to the given address.
func NewTXOutput(value int, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address)) // Locking the output to the address

	return txo
}

// Serialize serializes TxOutputs for storage.
func (outs TxOutputs) Serialize() []byte {
	var buffer bytes.Buffer
	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(outs)
	Handle(err)
	return buffer.Bytes()
}

// DeserializeOutputs deserializes TxOutputs from a byte slice.
func DeserializeOutputs(data []byte) TxOutputs {
	var outputs TxOutputs
	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&outputs)
	Handle(err)
	return outputs
}
