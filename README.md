# golang-blockchain

### Commands
1. go run main.go createwallet
2. go run main.go listaddresses
3. go run main.go printchain
4. go run main.go send -from FROM -to TO -amount AMOUNT
5. go run main.go getbalance -address ADDRESS
6. go run main.go reindexutxo
7. go run main.go createblockchain -address ADDRESS

---
### Test 
1. go run main.go createwallet
2. go run main.go createwallet
3. go run main.go createblockchain -address ADDRESS(FROM 1 dot)
4. go run main.go getbalance -address ADDRESS(FROM 1 dot)
5. go run main.go reindexutxo
6. go run main.go getbalance -address ADDRESS(FROM 1 dot)

# FAQ
Q: Почему в при транзакции становиться у отправляющего блока 40 ?
A: Потому что он добыл Genesis и транзакцию 2 блока