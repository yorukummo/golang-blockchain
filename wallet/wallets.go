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

// walletFile defines the pattern for the filename where wallets are stored.
const walletFile = "./tmp/wallets_%s.data"

// Wallets represents a collection of wallets.
type Wallets struct {
	Wallets map[string]*Wallet // Mapping from address to Wallet
}

// CreateWallets initializes and loads wallets from a file, or creates a new set of wallets.
func CreateWallets(nodeId string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*Wallet)

	// Loading wallets from file
	err := wallets.LoadFile(nodeId)

	return &wallets, err
}

// AddWallet creates and adds a new wallet to the collection.
func (ws *Wallets) AddWallet() string {
	wallet := MakeWallet() // Creating a new wallet
	address := fmt.Sprintf("%s", wallet.Address())

	ws.Wallets[address] = wallet // Adding the new wallet to the map

	return address
}

// GetAllAddresses returns all wallet addresses in the collection.
func (ws *Wallets) GetAllAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address) // Appending each address to the list
	}

	return addresses
}

// GetWallet retrieves a wallet by its address.
func (ws Wallets) GetWallet(address string) Wallet {
	return *ws.Wallets[address]
}

// LoadFile loads wallets from a file.
func (ws *Wallets) LoadFile(nodeId string) error {
	walletFile := fmt.Sprintf(walletFile, nodeId)
	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err // File does not exist
	}

	var wallets Wallets

	fileContent, err := ioutil.ReadFile(walletFile)
	if err != nil {
		return err // Error reading file
	}

	gob.Register(elliptic.P256()) // Registering the elliptic curve
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		return err // Error decoding file content
	}

	ws.Wallets = wallets.Wallets

	return nil
}

// SaveFile saves the collection of wallets to a file.
func (ws *Wallets) SaveFile(nodeId string) {
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeId)

	gob.Register(elliptic.P256()) // Registering the elliptic curve

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err) // Handling encoding error
	}

	err = ioutil.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err) // Handling file write error
	}
}
