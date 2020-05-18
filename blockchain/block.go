package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

type Block struct {
	Hash     []byte
	Data     []byte
	PrevHash []byte
	Nonce    int
}

func CreateBlock(data string, prevHash []byte) *Block {
	block := &Block{[]byte{}, []byte(data), prevHash, 0} // a call to the Block constructor
	//the data field converted into a slice of bytes
	pow := Newproof(block)
	nonce, hash := pow.Run()
	block.Hash = hash[:]
	block.Nonce = nonce
	return block
}

func Genesis() *Block {
	return CreateBlock("Genesis", []byte{})
}

//database working -- give us the byte representation of the block
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)
	Handle(err)
	return res.Bytes()
}
func Deserialize(data []byte) *Block {
	var block Block
	decoder := gob.NewDecoder(bytes.NewReader(data))
	err := decoder.Decode(&block)
	Handle(err)
	return &block
}
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
