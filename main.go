package main

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

type Block struct {
	position  int
	Data      bookCheckout
	timeStamp string
	Hash      string
	prevHash  string
}

type bookCheckout struct {
	bookID       string `json:"book_id"`
	user         string `json:"	user"`
	checkoutDate string `json:"checkout_date"`
	isGenesis    bool   `json:"	is_genesis"`
}

type Book struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Author      string `json:"author"`
	PublishDate string `json:"publish_date"`
	ISBN        string `json:"isbn"`
}

type Blockchain struct {
	blocks []*Block
}

var BlockChain *Blockchain

// struct function
func (b *Block) generateHash() {
	bytes, _ := json.Marshal(b.Data)

	data := string(b.position) + b.timeStamp + string(bytes) + b.prevHash
	hash := sha256.New()
	hash.Write([]byte(data))
	b.Hash = hex.EncodeToString(hash.Sum(nil))

}

func createBlock(prevBlock *Block, checkOutItem bookCheckout) *Block {
	block := &Block{}
	block.position = prevBlock.position + 1
	block.timeStamp = time.Now().String()
	block.prevHash = prevBlock.Hash
	block.generateHash()

	return block
}

// struct function
func (bc *Blockchain) addBlock(data bookCheckout) {
	prevBlock := bc.blocks[len(bc.blocks)-1]
	block := createBlock(prevBlock, data)
	if validBlock(block, prevBlock) {
		bc.blocks = append(bc.blocks, block)
	}

}

// validate hash
func (b *Block) validateHash(hash string) bool {
	b.generateHash()
	if b.Hash != hash {
		return false
	}
	return true
}

// checking valid block
func validBlock(block, prevBlock *Block) bool {
	if prevBlock.Hash != block.prevHash {
		return false
	}

	//validate hash

	if !block.validateHash(block.Hash) {
		return false
	}

	if prevBlock.position+1 != block.position {
		return false
	}

	return true
}

func writeBlock(w http.ResponseWriter, r *http.Request) {
	var checkOutItem bookCheckout
	err := json.NewDecoder(r.Body).Decode(&checkOutItem)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not create block: %v", err)
		w.Write([]byte("could not create new block"))
		//return
	}

	BlockChain.addBlock(checkOutItem)

}

func newBook(w http.ResponseWriter, r *http.Request) {
	var book Book
	err := json.NewDecoder(r.Body).Decode(&book)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not create book: %v", err)
		w.Write([]byte("could not create new book"))
		return
	}

	h := md5.New()
	io.WriteString(h, book.ISBN+" "+book.PublishDate)
	book.ID = fmt.Sprintf("%x", h.Sum(nil))
	resp, err := json.MarshalIndent(book, "", " ")
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		log.Printf("could not marshal payload: %v", err)
		w.Write([]byte("could not save book data"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write(resp)

}

// create a block by sending bookcheckout information
func GenesisBlock() *Block {
	return createBlock(&Block{}, bookCheckout{isGenesis: true})
}

// create a new blockchain with a genesis block
func NewBlockchain() *Blockchain {
	return &Blockchain{[]*Block{GenesisBlock()}}
}

// get blockchain
func getBlockchain(w http.ResponseWriter, r *http.Request) {
	jbytes, err := json.MarshalIndent(BlockChain.blocks, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(err)
		return
	}
	io.WriteString(w, string(jbytes))
}

func main() {
	BlockChain = NewBlockchain()
	route := mux.NewRouter()
	route.HandleFunc("/", getBlockchain).Methods("GET")
	route.HandleFunc("/", writeBlock).Methods("POST")
	route.HandleFunc("/new", newBook).Methods("POST")

	go func() {
		for _, block := range BlockChain.blocks {
			fmt.Printf("Prev. hash: %x\n", block.prevHash)
			bytes, _ := json.MarshalIndent(block.Data, "", "")
			fmt.Printf("Data:%v\n", string(bytes))
			fmt.Printf("Hash: %x\n", block.Hash)
			fmt.Println()
		}
	}()

	log.Println("listening on port:3000")
	log.Fatal(http.ListenAndServe(":3000", route))

}
