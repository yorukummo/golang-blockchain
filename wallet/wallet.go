// Package wallet implements the functionality of creating and managing wallets,
// and provides methods for encoding and decoding wallet information.
package wallet

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"log"

	"golang.org/x/crypto/ripemd160"
)

const (
	checksumLength = 4          // Length of the checksum in bytes
	version        = byte(0x00) // Version byte to prepend to the wallet address
)

// Wallet represents a cryptocurrency wallet.
type Wallet struct {
	PrivateKey ecdsa.PrivateKey // ECDSA private key
	PublicKey  []byte           // Corresponding public key
}

// Address generates a public address for this wallet.
func (w Wallet) Address() []byte {
	pubHash := PublicKeyHash(w.PublicKey)

	versionedHash := append([]byte{version}, pubHash...) // Appending version byte to the public hash
	checksum := Checksum(versionedHash)                  // Generating checksum for the versioned hash

	fullHash := append(versionedHash, checksum...) // Combining versioned hash and checksum
	address := Base58Encode(fullHash)              // Encoding to Base58

	return address
}

// NewKeyPair generates a new ECDSA private and public key pair.
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256() // Using P256 elliptic curve for generating the key

	private, err := ecdsa.GenerateKey(curve, rand.Reader) // Generating ECDSA key
	if err != nil {
		log.Panic(err)
	}

	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...) // Appending X and Y coordinates of the public key
	return *private, pub
}

// MakeWallet creates a new Wallet with a generated key pair.
func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// PublicKeyHash generates a public key hash using SHA256 and RIPEMD160 algorithms.
func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey) // Hashing the public key using SHA256

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicRipMD := hasher.Sum(nil) // Hashing the result using RIPEMD160

	return publicRipMD
}

// Checksum generates a checksum for a given payload.
func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)       // First SHA256 hash
	secondHash := sha256.Sum256(firstHash[:]) // Second SHA256 hash

	return secondHash[:checksumLength] // Returning first few bytes specified by checksumLength
}

// ValidateAddress checks if the provided address is valid.
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))                        // Decoding the address from Base58
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]      // Extracting checksum from the address
	version := pubKeyHash[0]                                           // Extracting version byte
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]        // Extracting public key hash
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...)) // Generating target checksum for comparison

	return bytes.Compare(actualChecksum, targetChecksum) == 0 // Comparing actual checksum with target checksum
}
