// Package blockchain represents the core logic for blockchain operations such as managing blocks,
// transactions, and their interrelationships like merkle trees and proof of work.
package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/argonautts/golang-blockchain/wallet"
	"log"
	"math/big"
	"strings"
)

// Transaction represents a blockchain transaction with inputs and outputs.
type Transaction struct {
	ID      []byte     // Unique identifier of the transaction
	Inputs  []TxInput  // Inputs to the transaction
	Outputs []TxOutput // Outputs from the transaction
}

// Hash generates a hash of the transaction, used as its ID.
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{} // Resetting the ID field to hash the rest of the transaction

	hash = sha256.Sum256(txCopy.Serialize()) // Hashing the serialized transaction

	return hash[:]
}

// Serialize converts the transaction to a byte slice for storage or transmission.
func (tx Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

// DeserializeTransaction converts a byte slice back into a Transaction struct.
func DeserializeTransaction(data []byte) Transaction {
	var transaction Transaction

	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&transaction)
	Handle(err)
	return transaction
}

// CoinbaseTx creates a new coinbase transaction, which is the first transaction in a block.
func CoinbaseTx(to, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 24)
		_, err := rand.Read(randData)
		Handle(err)
		data = fmt.Sprintf("%x", randData) // Generating a random data if none provided
	}

	txin := TxInput{[]byte{}, -1, nil, []byte(data)} // Creating a special input for coinbase transaction
	txout := NewTXOutput(20, to)                     // Creating output for the transaction

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{*txout}}
	tx.ID = tx.Hash() // Setting the transaction ID as the hash of the transaction

	return &tx
}

// NewTransaction creates a new regular transaction from a wallet to a target address.
func NewTransaction(w *wallet.Wallet, to string, amount int, UTXO *UTXOSet) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	pubKeyHash := wallet.PublicKeyHash(w.PublicKey)
	acc, validOutputs := UTXO.FindSpendableOutputs(pubKeyHash, amount)

	if acc < amount {
		log.Panic("ERROR: Not enough funds")
	}

	// Building a list of inputs
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, nil, w.PublicKey}
			inputs = append(inputs, input)
		}
	}

	from := fmt.Sprintf("%s", w.Address())

	// Creating output for the receiver
	outputs = append(outputs, *NewTXOutput(amount, to))

	// Creating a change output if needed
	if acc > amount {
		outputs = append(outputs, *NewTXOutput(acc-amount, from))
	}

	tx := Transaction{nil, inputs, outputs}
	tx.ID = tx.Hash()                                  // Setting the transaction ID
	UTXO.Blockchain.SignTransaction(&tx, w.PrivateKey) // Signing the transaction

	return &tx
}

// IsCoinbase checks if the transaction is a coinbase transaction.
func (tx *Transaction) IsCoinbase() bool {
	// Coinbase transaction has one input with no ID and a -1 Out index.
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

// Sign signs each input of the transaction.
func (tx *Transaction) Sign(privKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return // Coinbase transactions don't require a signature.
	}

	txCopy := tx.TrimmedCopy() // Creating a trimmed copy for signing

	for inId, in := range txCopy.Inputs {
		prevTX := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTX.Outputs[in.Out].PubKeyHash

		dataToSign := fmt.Sprintf("%x\n", txCopy)

		// Signing the data
		r, s, err := ecdsa.Sign(rand.Reader, &privKey, []byte(dataToSign))
		Handle(err)
		signature := append(r.Bytes(), s.Bytes()...)

		tx.Inputs[inId].Signature = signature
		txCopy.Inputs[inId].PubKey = nil
	}
}

// Verify verifies the signatures of Transaction inputs.
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	if tx.IsCoinbase() {
		return true // Coinbase transactions don't require verification.
	}

	txCopy := tx.TrimmedCopy() // Creating a trimmed copy for verification
	curve := elliptic.P256()   // Using P256 elliptic curve

	for inId, in := range tx.Inputs {
		prevTx := prevTXs[hex.EncodeToString(in.ID)]
		txCopy.Inputs[inId].Signature = nil
		txCopy.Inputs[inId].PubKey = prevTx.Outputs[in.Out].PubKeyHash

		r := big.Int{}
		s := big.Int{}
		sigLen := len(in.Signature)
		r.SetBytes(in.Signature[:(sigLen / 2)])
		s.SetBytes(in.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(in.PubKey)
		x.SetBytes(in.PubKey[:(keyLen / 2)])
		y.SetBytes(in.PubKey[(keyLen / 2):])

		dataToVerify := fmt.Sprintf("%x\n", txCopy)

		rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
		if ecdsa.Verify(&rawPubKey, []byte(dataToVerify), &r, &s) == false {
			return false // Signature does not match.
		}
		txCopy.Inputs[inId].PubKey = nil
	}

	return true // All signatures are valid.
}

// TrimmedCopy creates a copy of the transaction to be used in signing and verification.
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	// Copying inputs and outputs without the signature and public key.
	for _, in := range tx.Inputs {
		inputs = append(inputs, TxInput{in.ID, in.Out, nil, nil})
	}

	for _, out := range tx.Outputs {
		outputs = append(outputs, TxOutput{out.Value, out.PubKeyHash})
	}

	txCopy := Transaction{tx.ID, inputs, outputs}

	return txCopy
}

// String returns a human-readable representation of the transaction.
func (tx Transaction) String() string {
	var lines []string

	lines = append(lines, fmt.Sprintf("--- Transaction %x:", tx.ID))
	for i, input := range tx.Inputs {
		lines = append(lines, fmt.Sprintf("     Input %d:", i))
		lines = append(lines, fmt.Sprintf("       TXID:     %x", input.ID))
		lines = append(lines, fmt.Sprintf("       Out:       %d", input.Out))
		lines = append(lines, fmt.Sprintf("       Signature: %x", input.Signature))
		lines = append(lines, fmt.Sprintf("       PubKey:    %x", input.PubKey))
	}

	for i, output := range tx.Outputs {
		lines = append(lines, fmt.Sprintf("     Output %d:", i))
		lines = append(lines, fmt.Sprintf("       Value:  %d", output.Value))
		lines = append(lines, fmt.Sprintf("       Script: %x", output.PubKeyHash))
	}

	return strings.Join(lines, "\n")
}
