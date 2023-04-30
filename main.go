package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

var BlockChain *Blockchain

type Block struct {
	PrevHash  string
	Hash      string
	Data      BlockData
	Timestamp string
	Pos       int
}

type Blockchain struct {
	blocks []*Block
}

type BlockData struct {
	Data      string `json:"data"`
	User      string `json:"address"`
	IsGenesis bool   `json:"is_genesis"`
}

func (bc *Blockchain) AddBlock(data BlockData) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	block := createBlock(prevBlock, data)
	if validateBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
	}
}

func (b *Block) generateHash() {
	bytes, _ := json.Marshal(b.Data)
	data := strconv.Itoa(b.Pos) + b.Timestamp + string(bytes) + b.PrevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))
}

func (b *Block) validateHash(hash string) bool {
	b.generateHash()
	return b.Hash == hash
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
}

func validateBlock(block, prevBlock *Block) bool {
	if block.PrevHash != prevBlock.Hash {
		return false
	}
	if prevBlock.Pos+1 != block.Pos {
		return false
	}
	if !block.validateHash(block.Hash) {
		return false
	}
	return true
}

func createBlock(prevBlock *Block, bc BlockData) *Block {
	block := &Block{}
	block.Pos = prevBlock.Pos + 1
	block.Data = bc
	block.Timestamp = time.Now().String()
	block.PrevHash = prevBlock.Hash
	block.generateHash()
	return block
}

func GenesisBlock() *Block {
	return createBlock(&Block{}, BlockData{IsGenesis: true})
}

func getBlockchain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.blocks, "", " ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	w.Write(jbytes)
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkoutData BlockData
	if err := json.NewDecoder(r.Body).Decode(&checkoutData); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not write Block: %v", err)
		json.NewEncoder(w).Encode(err)
		return
	}

	BlockChain.AddBlock(checkoutData)
	w.WriteHeader(http.StatusOK)
}

func main() {
	BlockChain = NewBlockchain()
	r := mux.NewRouter()

	r.HandleFunc("/", getBlockchain).Methods("GET")
	r.HandleFunc("/", writeBlock).Methods("POST")

	fmt.Println("Server listning on 3000")
	log.Fatal(http.ListenAndServe(":3000", r))
}
