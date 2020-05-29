package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/dgraph-io/badger"
)

// type Blockchain struct {
// 	Blocks []*Block //use uper case naming to make it public
// }
//refactor the blockchain struct keeping the badgrdatabase in mind
const (
	dbPath      = "./tmp/blocks_%s"       // this %s allows us to have multiple databases for each of our nodes
	dbFile      = "./tmp/blocks/MANIFEST" //tocheck whether database exists or not
	genesisData = "First Transaction from genesis"
)

type Blockchain struct {
	LastHash []byte
	Database *badger.DB
}

//helper function to check whether blockchain exists or not
func DBexisits(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}
	return true
}

//if one instance of the datatabse is already running then it
// will conatin the lock file and if the instance terminates without properly garbage removal
//then this lock file will remain in the folder

//retry function deletes the badger file and change the original options
func retry(dir string, originalOpts badger.Options) (*badger.DB, error) {
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil {
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOpts := originalOpts
	retryOpts.Truncate = true
	db, err := badger.Open(retryOpts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error) {
	if db, err := badger.Open(opts); err != nil {
		if strings.Contains(err.Error(), "LOCK") {
			if db, err := retry(dir, opts); err == nil {
				log.Println("database unlocked, value log truncated")
				return db, nil
			}
			log.Println("could not unlock database:", err)
		}
		return nil, err
	} else {
		return db, nil
	}
}

func InitBlockchain(address string, nodeId string) *Blockchain {
	path := fmt.Sprintf(dbPath, nodeId)
	var lastHash []byte

	if DBexisits(path) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	opts := badger.DefaultOptions(path)
	opts.Dir = path      //store keys and metadata
	opts.ValueDir = path //databse will store all the values

	db, err := openDB(path, opts)
	Handle(err)
	err = db.Update(func(txn *badger.Txn) error {
		//check if there is a blockchain already stored or not
		//if there is alreadya blockchainthen we will create a new blockchain instance in memory and we will get the
		//last hash of our blockchian in our disk database and we will push to this instance in memory
		//the reason why the last hash isimportant is that it helps derive a new block in our blockchain
		//if there is no existing blockchain we will create a genesis block we will push it in our databse then we will save the genesis block hash as the lastblock hash in our databse
		cbtx := CoinbaseTx(address, genesisData)
		genesis := Genesis(cbtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)
		lastHash = genesis.Hash
		return err
	})

	Handle(err)
	blockchain := Blockchain{lastHash, db} // create a new blockchain in the memory
	return &blockchain                     // return a referenec to this blockchain

}
func ContinueBlockchain(nodeId string) *Blockchain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBexisits(path) == false {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}
	var lastHash []byte
	opts := badger.DefaultOptions(path)
	opts.Dir = path      //store keys and metadata
	opts.ValueDir = path //databse will store all the values

	db, err := openDB(path, opts)
	Handle(err)
	err = db.Update(func(txn *badger.Txn) error {

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
	chain := Blockchain{lastHash, db}
	return &chain

}

func (chain *Blockchain) MineBlock(transactions []*Transaction) *Block {
	var lastHeight int
	var lastHash []byte
	var lastBlockData []byte
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err1 := item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		Handle(err1)
		item, err = txn.Get([]byte(lastHash))
		Handle(err)
		err1 = item.Value(func(val []byte) error {
			lastBlockData = append([]byte{}, val...)
			return nil
		})
		Handle(err1)
		lastBlock := Deserialize(lastBlockData)
		lastHeight = lastBlock.Height

		return err
	})
	Handle(err)
	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)
		chain.LastHash = newBlock.Hash
		return err
	})
	Handle(err)
	return newBlock

}

func (chain *Blockchain) AddBlock(block *Block) {
	var lastHash []byte
	var lastBlockData []byte
	err := chain.Database.Update(func(txn *badger.Txn) error {

		if _, err := txn.Get(block.Hash); err == nil {
			return nil
		}
		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		Handle(err)

		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err1 := item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		Handle(err1)

		item, err = txn.Get([]byte(lastHash))
		Handle(err)
		err1 = item.Value(func(val []byte) error {
			lastBlockData = append([]byte{}, val...)
			return nil
		})
		Handle(err1)
		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			Handle(err)
			chain.LastHash = block.Hash
		}

		return nil
	})
	Handle(err)
}

func (chain *Blockchain) GetBestHeight() int {
	var lastBlock Block
	var lastHash []byte
	var lastBlockData []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		err1 := item.Value(func(val []byte) error {
			lastHash = append([]byte{}, val...)
			return nil
		})
		Handle(err1)

		item, err = txn.Get([]byte(lastHash))
		Handle(err)
		err1 = item.Value(func(val []byte) error {
			lastBlockData = append([]byte{}, val...)
			return nil
		})
		Handle(err1)
		lastBlock = *Deserialize(lastBlockData)

		return nil
	})
	Handle(err)

	return lastBlock.Height
}

func (chain *Blockchain) GetBlock(blockHash []byte) (Block, error) {
	var block Block
	var blockData []byte

	err := chain.Database.View(func(txn *badger.Txn) error {
		if item, err := txn.Get(blockHash); err != nil {
			return errors.New("Block is not found")
		} else {
			err1 := item.Value(func(val []byte) error {
				blockData = append([]byte{}, val...)
				return nil
			})
			Handle(err1)

			block = *Deserialize(blockData)
		}
		return nil
	})
	if err != nil {
		return block, err
	}

	return block, nil
}

func (chain *Blockchain) GetBlockHashes() [][]byte {
	var blocks [][]byte

	iter := chain.Iterator()

	for {
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0 {
			break
		}
	}

	return blocks
}

func (chain *Blockchain) FindUTXO() map[string]TxOutputs {
	UTXOs := make(map[string]TxOutputs)
	spentTXOs := make(map[string][]int)
	iter := chain.Iterator()
	for {
		block := iter.Next()
		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Outputs {
				if spentTXOs[txID] != nil {
					for _, spendOut := range spentTXOs[txID] {
						if spendOut == outIdx {
							continue Outputs
						}
					}
				}
				outs := UTXOs[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXOs[txID] = outs
			}
			if tx.IsCoinbase() == false {
				for _, in := range tx.Inputs {
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}

	}
	return UTXOs
}

func (bc *Blockchain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()
	for {
		block := iter.Next()
		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction does not exist")
}
func (bc *Blockchain) SignTansaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTxs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}
	tx.Sign(privKey, prevTxs)
}
func (bc *Blockchain) VerifyTansaction(tx *Transaction) bool {

	if tx.IsCoinbase() {
		return true
	}

	prevTxs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTx, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTxs[hex.EncodeToString(prevTx.ID)] = prevTx
	}
	return tx.Verify(prevTxs)
}
