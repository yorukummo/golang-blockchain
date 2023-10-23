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
	checksumLength = 4
	version        = byte(0x00)
)

// Wallet представляет кошелек с приватным и публичным ключами.
type Wallet struct {
	PrivateKey ecdsa.PrivateKey // Алгоритм цифровой подписи с элиптической кривой.
	PublicKey  []byte
}

// Address генерирует адрес кошелька на основе публичного ключа.
func (w Wallet) Address() []byte {
	// Получаем хеш от публичного ключа.
	pubHash := PublicKeyHash(w.PublicKey)

	// Добавляем версию к хешу публичного ключа.
	versionedHash := append([]byte{version}, pubHash...)
	// Генерируем контрольную сумму для версированного хеша.
	checksum := Checksum(versionedHash)

	// Объединяем версированный хеш и контрольную сумму.
	fullHash := append(versionedHash, checksum...)
	// Кодируем все в Base58.
	address := Base58Encode(fullHash)

	return address
}

// ValidateAddress проверяет корректность адреса кошелька.
func ValidateAddress(address string) bool {
	pubKeyHash := Base58Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-checksumLength:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-checksumLength]
	targetChecksum := Checksum(append([]byte{version}, pubKeyHash...))

	// Сравниваем контрольные суммы.
	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

// NewKeyPair генерирует новую пару приватного и публичного ключей.
func NewKeyPair() (ecdsa.PrivateKey, []byte) {
	curve := elliptic.P256() // 256 бит элиптической кривой.

	private, err := ecdsa.GenerateKey(curve, rand.Reader)

	if err != nil {
		log.Panic(err)
	}

	// Получаем публичный ключ как комбинацию координат X и Y.
	pub := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return *private, pub
}

// MakeWallet создает новый кошелек с парой ключей.
func MakeWallet() *Wallet {
	private, public := NewKeyPair()
	wallet := Wallet{private, public}

	return &wallet
}

// PublicKeyHash вычисляет хеш публичного ключа с использованием sha256 и ripemd160.
func PublicKeyHash(pubKey []byte) []byte {
	pubHash := sha256.Sum256(pubKey)

	hasher := ripemd160.New()
	_, err := hasher.Write(pubHash[:])
	if err != nil {
		log.Panic(err)
	}

	publicRipMD := hasher.Sum(nil)

	return publicRipMD
}

// Checksum вычисляет контрольную сумму для переданного байтового среза.
func Checksum(payload []byte) []byte {
	firstHash := sha256.Sum256(payload)
	secondHash := sha256.Sum256(firstHash[:])

	return secondHash[:checksumLength]
}
