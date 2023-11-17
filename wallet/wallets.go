package wallet

import (
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"io/ioutil"
	"log"
	"os"
)

// Путь к файлу, где будут сохраняться данные кошельков.
const walletFile = "./tmp/wallets_%s.data"

// Wallets содержит мапу всех кошельков.
type Wallets struct {
	Wallets map[string]*Wallet
}

// CreateWallets создает и возвращает структуру Wallets после попытки загрузки из файла.
func CreateWallets(nodeId string) (*Wallets, error) {
	wallet := Wallets{}
	wallet.Wallets = make(map[string]*Wallet)

	err := wallet.LoadFile(nodeId)

	return &wallet, err
}

// AddWallet добавляет новый кошелек в структуру Wallets и возвращает его адрес.
func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet()
	address := fmt.Sprintf("%s", wallet.Address())

	ws.Wallets[address] = wallet

	return address
}

// GetAllAddresses возвращает все адреса кошельков, хранящихся в структуре Wallets.
func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

// GetWallet возвращает кошелек по указанному адресу.
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFile загружает кошельки из файла в структуру Wallets.
func (ws *Wallets) LoadFile(nodeId string) error {
	walletFile := fmt.Sprintf(walletFile, nodeId)

	// Проверяем существование файла.
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	var wallets Wallets

	// Читаем содержимое файла.
	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err
	}

	// Регистрируем тип для корректной десериализации.
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		return err
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveFile сохраняет текущую структуру Wallets в файл.
func (ws *Wallets) SaveFile(nodeId string) {
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeId)

	// Регистрируем тип для корректной сериализации.
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	// Записываем содержимое в файл с правами 0644.
	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}
