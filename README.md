# Golang blockchain
Assignment on distributed software systems at the university.

## Installation
1. Make sure [Golang 1.16+](https://go.dev/dl/) is installed.
2. Make sure calling the go command works in the terminal.

## Build
To start the project, clone the project and navigate to the `golang-blockchain` folder in the terminal.

```
cd $GOPATH/src/golang-blockchain
```
Install all dependencies in the `go.mod` file using `go get`.
To invoke the helper on the command line, simply run the main project file `main.go`.
```go
go run main.go
```

# Documentation

The main executable file of the project is `main.go`, where the command line is called to interact with the user.

The `go.mod` and `go.sum` files are used to store project requirements. For example, the query to the required database and its version.

The file `cli/cli.go` represents the command line logic, it provides instructions for the program to work to the user.

The `tmp` folder and the blocks folder inside it are needed for the badger database.

The network folder and inside it the `network.go` file realizes network communication in the application.

There are 7 files inside the blockchain folder:

1. `block.go`

    Provides the structure, creation of the first Genesis block and further blocks.


2. `blockchain.go`
   
    Provides the structure of the blockchain and iterates over it. Used to implement blockchain functions.


3. `chain_iter.go`
   
    Provides a structure for iterating over blockchains.


4. `merkle.go`
   
    Provides a structure for the Merkle tree and its nodes.


5. `proof.go`

    Provides structure for proof of block signing and its further mining.


6. `transaction.go`

   Provides the structure of transactions and their signatures.


7. `tx.go`

   Provides structure for input and output data for transactions.


8. `utxo.go`

    Provides logic for reindexing and updating transactions.

The wallet folder is used to store 3 files:
1. `utils.go`
   
    Stores functions for base58 encoding and decoding.


2. `wallet.go`
   
    Provides the wallet structure and all its logic.


3. `wallets.go`

    Provides a map of all wallets and their storage in a file.

### Commands

Creating a wallet for further work with blockchain
```go
go run main.go createwallet
```

Blockchain creation, including Genesis
```go
go run main.go createblockchain -address ADDRESS
```
Transactions for token exchange
```go
go run main.go send -from FROM -to TO -amount AMOUNT -mine
```
Display blockchain information
```go
go run main.go printchain
```
Wallets address output
```go
go run main.go listaddresses
```
Check balance in the wallet
```go
go run main.go getbalance -address ADDRESS
```
Re-indexing tokens
```go
go run main.go reindexutxo
```
Starting NODE and the miner
```go
go run main.go startnode -miner ADDRESS
```

---

### Example use for macOS
Below is a test example of using commands to work with the program in the terminal.
I advise to open several tabs in the terminal to work with `NODE_ID` and further work with them.

0. export set NODE_ID=3000,     export set NODE_ID=4000,     export set NODE_ID=5000
1. go run main.go createwallet, go run main.go createwallet, go run main.go createwallet
2. go run main.go createblockchain -address ADDRESS
3. cp -R blocks_3000/ blocks_4000/
4. cp -R blocks_3000/ blocks_5000/
5. cp -R blocks_3000/ blocks_gen/
6. go run main.go send -from ADDRESS(NODE_ID=3000) -to ADDRESS(NODE_ID=5000) -amount 10 -mine
7. go run main.go startnode(NODE_ID=3000)     go run main.go startnode(NODE_ID=4000) go run main.go startnode(NODE_ID=5000) -miner ADDRESS(NODE_ID=5000)
8. Exit NODE_ID=4000
9. go run main.go send -from ADDRESS(NODE_ID=5000) -to ADDRESS(NODE_ID=4000) -amount 1
10. go run main.go send -from ADDRESS(NODE_ID=4000) -to ADDRESS(NODE_ID=3000) -amount 1