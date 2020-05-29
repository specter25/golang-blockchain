package blockchain

import "github.com/dgraph-io/badger"

// to allow us to iterate through the blockchain
type BlockchainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
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
