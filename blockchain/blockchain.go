package blockchain

import (
	"fmt"

	"github.com/dgraph-io/badger"
)

// type Blockchain struct {
// 	Blocks []*Block //use uper case naming to make it public
// }
//refactor the blockchain struct keeping the badgrdatabase in mind
const (
	dbPath = "./tmp/blocks"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

// to allow us to iterate through the blockchain
type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func InitBlockchain() *Blockchain {
	var lastHash []byte
	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath      //store keys and metadata
	opts.ValueDir = dbPath //databse will store all the values

	db, err := badger.Open(opts)
	Handle(err)
	err = db.Update(func(txn *badger.Txn) error {
		//check if there is a blockchain already stored or not
		//if there is alreadya blockchainthen we will create a new blockchain instance in memory and we will get the
		//last hash of our blockchian in our disk database and we will push to this instance in memory
		//the reason why the last hash isimportant is that it helps derive a new block in our blockchain
		//if there is no existing blockchain we will create a genesis block we will push it in our databse then we will save the genesis block hash as the lastblock hash in our databse

		if _, err := txn.Get([]byte("lh")); err == badger.ErrKeyNotFound {
			fmt.Println("No existing blockchain found ")
			genesis := Genesis()
			fmt.Println("Genesis proved")
			err = txn.Set(genesis.Hash, genesis.Serialize())
			Handle(err)
			err = txn.Set([]byte("lh"), genesis.Hash)
			lastHash = genesis.Hash
			return err
		} else {
			item, err := txn.Get([]byte("lh"))
			Handle(err)
			err1 := item.Value(func(val []byte) error {
				lastHash = append([]byte{}, val...)
				return nil
			})
			Handle(err1)
			return err
		}

	})

	Handle(err)
	blockchain := Blockchain{lastHash, db} // create a new blockchain in the memory
	return &blockchain                     // return a referenec to this blockchain

}

func (chain *Blockchain) AddBlock(data string) {
	var lastHash []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err1 := item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		Handle(err1)
		return err
	})
	Handle(err)
	newBlock := CreateBlock(data, lastHash)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	Handle(err)

}
func (chain *Blockchain) Iterator() *BlockchainIterator {
	iter := &BlockchainIterator{chain.LastHash, chain.Database}
	// as we are starting from the last hash of the blockchain we will be iterating backwards through the blocks

	return iter
}
func (iter *BlockchainIterator) Next() *Block {
	var block *Block
	var encodedBlock []byte
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash)
		Handle(err)
		err1 := item.Value(func(val []byte) error {
			encodedBlock = append([]byte{}, val...)
			return nil
		})
		block = Deserialize(encodedBlock)
		return err1
	})
	Handle(err)
	iter.CurrentHash = block.PrevHash
	return block
}
