// Package network implements a network module for the code.
package network

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"github.com/argonautts/golang-blockchain/blockchain"
	"github.com/vrecan/death"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"runtime"
	"syscall"
)

// Constant definitions for network protocol parameters.
const (
	protocol      = "tcp" // Network protocol used
	version       = 1     // Version of the protocol
	commandLength = 12    // Length of the command in the protocol
)

// Global variables used in the network package.
var (
	nodeAddress     string                                    // Address of the current node
	mineAddress     string                                    // Mining address
	KnownNodes      = []string{"localhost:3000"}              // Initial known nodes
	blocksInTransit = [][]byte{}                              // Blocks being transmitted
	memoryPool      = make(map[string]blockchain.Transaction) // Pool of transactions
)

type Addr struct {
	AddrList []string
}

type Block struct {
	AddrFrom string
	Block    []byte
}

type GetBlocks struct {
	AddrFrom string
}

type GetData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type Inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type Tx struct {
	AddrFrom    string
	Transaction []byte
}

type Version struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

// Utility functions for the network communication

// CmdToBytes converts command string to bytes.
func CmdToBytes(cmd string) []byte {
	var bytes [commandLength]byte

	for i, c := range cmd {
		bytes[i] = byte(c)
	}

	return bytes[:]
}

// BytesToCmd converts bytes back to command string.
func BytesToCmd(bytes []byte) string {
	var cmd []byte

	for _, b := range bytes {
		if b != 0x0 {
			cmd = append(cmd, b)
		}
	}

	return fmt.Sprintf("%s", cmd)
}

// ExtractCmd extracts the command from the received request.
func ExtractCmd(request []byte) []byte {
	return request[:commandLength]
}

// RequestBlocks Requests blocks from all known nodes.
func RequestBlocks() {
	for _, node := range KnownNodes {
		SendGetBlocks(node)
	}
}

// SendAddr sends the 'addr' command to a specified address.
func SendAddr(address string) {
	nodes := Addr{KnownNodes}
	nodes.AddrList = append(nodes.AddrList, nodeAddress)
	payload := GobEncode(nodes)
	request := append(CmdToBytes("addr"), payload...)

	SendData(address, request)
}

// SendBlock sends a block to a specified address.
func SendBlock(addr string, b *blockchain.Block) {
	data := Block{nodeAddress, b.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("block"), payload...)

	SendData(addr, request)
}

// SendData general method to send data to a specified address.
func SendData(addr string, data []byte) {
	conn, err := net.Dial(protocol, addr)

	if err != nil {
		fmt.Printf("%s unavailable\n", addr)
		var updatedNodes []string

		for _, node := range KnownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}

		KnownNodes = updatedNodes

		return
	}

	defer conn.Close()

	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

// SendInv sends 'inv' command with inventory details.
func SendInv(address, kind string, items [][]byte) {
	inventory := Inv{nodeAddress, kind, items}
	payload := GobEncode(inventory)
	request := append(CmdToBytes("inv"), payload...)

	SendData(address, request)
}

// SendGetBlocks sends 'getblocks' command to a specified address.
func SendGetBlocks(address string) {
	payload := GobEncode(GetBlocks{nodeAddress})
	request := append(CmdToBytes("getblocks"), payload...)

	SendData(address, request)
}

// SendGetData sends 'getdata' command to a specified address.
func SendGetData(address, kind string, id []byte) {
	payload := GobEncode(GetData{nodeAddress, kind, id})
	request := append(CmdToBytes("getdata"), payload...)

	SendData(address, request)
}

// SendTx sends a transaction to a specified address.
func SendTx(addr string, tnx *blockchain.Transaction) {
	data := Tx{nodeAddress, tnx.Serialize()}
	payload := GobEncode(data)
	request := append(CmdToBytes("tx"), payload...)

	SendData(addr, request)
}

// SendVersion sends 'version' command to a specified address.
func SendVersion(addr string, chain *blockchain.BlockChain) {
	bestHeight := chain.GetBestHeight()
	payload := GobEncode(Version{version, bestHeight, nodeAddress})

	request := append(CmdToBytes("version"), payload...)

	SendData(addr, request)
}

// Handlers for different types of received messages

// HandleAddr handles 'addr' command.
func HandleAddr(request []byte) {
	var buff bytes.Buffer
	var payload Addr

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	KnownNodes = append(KnownNodes, payload.AddrList...)
	fmt.Printf("there are %d known nodes", len(KnownNodes))
	RequestBlocks()
}

// HandleBlock handles 'block' command.
func HandleBlock(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := blockchain.Deserialize(blockData)

	fmt.Println("Recevied a new block!")
	chain.AddBlock(block)

	fmt.Printf("Added Block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		SendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := blockchain.UTXOSet{chain}
		UTXOSet.Reindex()
	}
}

// HandleInv handles 'inv' (inventory) command by processing the inventory of blocks or transactions received from another node.
func HandleInv(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Inv

	// Decoding the payload from the request
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// Logging received inventory details
	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	// Handling block type inventory
	if payload.Type == "block" {
		blocksInTransit = payload.Items

		blockHash := payload.Items[0]
		SendGetData(payload.AddrFrom, "block", blockHash)

		// Filtering out the requested block from blocksInTransit
		newInTransit := [][]byte{}
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	// Handling transaction type inventory
	if payload.Type == "tx" {
		txID := payload.Items[0]

		// Requesting the transaction data if it's not in the memory pool
		if memoryPool[hex.EncodeToString(txID)].ID == nil {
			SendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

// HandleGetBlocks handles 'getblocks' command by sending an inventory of all block hashes to the requester.
func HandleGetBlocks(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload GetBlocks

	// Decoding the payload from the request
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// Getting all block hashes from the blockchain and sending them
	blocks := chain.GetBlockHashes()
	SendInv(payload.AddrFrom, "block", blocks)
}

// HandleGetData handles 'getdata' command by sending requested block or transaction data to the requester.
func HandleGetData(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload GetData

	// Decoding the payload from the request
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// Sending the requested block
	if payload.Type == "block" {
		block, err := chain.GetBlock([]byte(payload.ID))
		if err != nil {
			return
		}

		SendBlock(payload.AddrFrom, &block)
	}

	// Sending the requested transaction
	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := memoryPool[txID]

		SendTx(payload.AddrFrom, &tx)
	}
}

// HandleTx handles 'tx' (transaction) command by adding the transaction to the memory pool and propagating it.
func HandleTx(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Tx

	// Decoding the payload from the request
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// Adding the transaction to the memory pool
	txData := payload.Transaction
	tx := blockchain.DeserializeTransaction(txData)
	memoryPool[hex.EncodeToString(tx.ID)] = tx

	fmt.Printf("%s, %d", nodeAddress, len(memoryPool))

	if nodeAddress == KnownNodes[0] {
		for _, node := range KnownNodes {
			if node != nodeAddress && node != payload.AddrFrom {
				SendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(memoryPool) >= 2 && len(mineAddress) > 0 {
			MineTx(chain)
		}
	}
}

// MineTx mines a new block with transactions from the memory pool
func MineTx(chain *blockchain.BlockChain) {
	var txs []*blockchain.Transaction

	// Verifying and collecting valid transactions
	for id := range memoryPool {
		fmt.Printf("tx: %s\n", memoryPool[id].ID)
		tx := memoryPool[id]
		if chain.VerifyTransaction(&tx) {
			txs = append(txs, &tx)
		}
	}

	if len(txs) == 0 {
		fmt.Println("All transactions are invalid")
		return
	}

	cbTx := blockchain.CoinbaseTx(mineAddress, "")
	txs = append(txs, cbTx)

	newBlock := chain.MineBlock(txs)
	UTXOSet := blockchain.UTXOSet{chain}
	UTXOSet.Reindex()

	fmt.Println("A new block has been mined")

	for _, tx := range txs {
		txID := hex.EncodeToString(tx.ID)
		delete(memoryPool, txID)
	}

	for _, node := range KnownNodes {
		if node != nodeAddress {
			SendInv(node, "block", [][]byte{newBlock.Hash})
		}
	}

	if len(memoryPool) > 0 {
		MineTx(chain)
	}
}

// HandleVersion handles 'version' command, comparing the blockchain's height and syncing if necessary.
func HandleVersion(request []byte, chain *blockchain.BlockChain) {
	var buff bytes.Buffer
	var payload Version

	// Decoding the payload from the request
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// Comparing blockchain heights and taking appropriate action
	bestHeight := chain.GetBestHeight()
	otherHeight := payload.BestHeight

	// Requesting blocks from a node with a higher blockchain
	if bestHeight < otherHeight {
		SendGetBlocks(payload.AddrFrom)
	} else if bestHeight > otherHeight {
		// Sending version to a node with a lower blockchain
		SendVersion(payload.AddrFrom, chain)
	}

	// Adding new node to known nodes list if not already known
	if !NodeIsKnown(payload.AddrFrom) {
		KnownNodes = append(KnownNodes, payload.AddrFrom)
	}
}

// HandleConnection manages incoming connections and routes commands to appropriate handlers.
func HandleConnection(conn net.Conn, chain *blockchain.BlockChain) {
	req, err := ioutil.ReadAll(conn)
	defer conn.Close()

	if err != nil {
		log.Panic(err)
	}

	// Routing the command to the corresponding handler
	command := BytesToCmd(req[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "addr":
		HandleAddr(req)
	case "block":
		HandleBlock(req, chain)
	case "inv":
		HandleInv(req, chain)
	case "getblocks":
		HandleGetBlocks(req, chain)
	case "getdata":
		HandleGetData(req, chain)
	case "tx":
		HandleTx(req, chain)
	case "version":
		HandleVersion(req, chain)
	default:
		fmt.Println("Unknown command")
	}

}

// StartServer initializes a server for the blockchain node.
func StartServer(nodeID, minerAddress string) {
	nodeAddress = fmt.Sprintf("localhost:%s", nodeID)
	mineAddress = minerAddress
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Panic(err)
	}
	defer ln.Close()

	// Continues an existing blockchain or creates a new one
	chain := blockchain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	go CloseDB(chain)

	// Syncs with the main node if not the main node itself
	if nodeAddress != KnownNodes[0] {
		SendVersion(KnownNodes[0], chain)
	}

	// Listening for incoming connections
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go HandleConnection(conn, chain)

	}
}

// GobEncode encodes data into bytes using GOB encoding.
func GobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// NodeIsKnown checks if a node is already known.
func NodeIsKnown(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

// CloseDB gracefully closes the database connection on termination signals.
func CloseDB(chain *blockchain.BlockChain) {
	d := death.NewDeath(syscall.SIGINT, syscall.SIGTERM, os.Interrupt)

	// Function to close the database when the application terminates
	d.WaitForDeathWithFunc(func() {
		defer os.Exit(1)
		defer runtime.Goexit()
		chain.Database.Close()
	})
}
