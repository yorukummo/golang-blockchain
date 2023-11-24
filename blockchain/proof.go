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

const Difficulty = 12 // Difficulty defines the complexity of the mining process.

// ProofOfWork represents the proof of work algorithm associated with a block.
type ProofOfWork struct {
	Block  *Block   // The block to which this proof of work applies.
	Target *big.Int // The target hash for this proof of work.
}

// NewProof creates a new proof of work for a given block.
func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	// Left-shifting 1 by 256-Difficulty bits to set the target.
	target.Lsh(target, uint(256-Difficulty))

	pow := &ProofOfWork{b, target}

	return pow
}

// InitData prepares the data for hashing to find a new nonce.
func (pow *ProofOfWork) InitData(nonce int) []byte {
	// Joining block data with nonce and difficulty to prepare for hashing.
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)

	return data
}

// Run performs the proof-of-work computation.
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	// Iteratively incrementing nonce to find the hash that meets the target.
	for nonce < math.MaxInt64 {
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		fmt.Printf("\r%x", hash) // Printing the hash for each attempt (optional).
		intHash.SetBytes(hash[:])

		// Comparing the hash against the target.
		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}
	}
	fmt.Println()

	return nonce, hash[:]
}

// Validate checks whether the block's proof of work is valid.
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	// Preparing and hashing the data with the block's nonce.
	data := pow.InitData(pow.Block.Nonce)
	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	// The hash must be less than the target.
	return intHash.Cmp(pow.Target) == -1
}

// ToHex converts a numerical value to a byte slice in big endian format.
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}
