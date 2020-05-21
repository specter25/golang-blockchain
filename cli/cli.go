package cli

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"

	"github.com/specter25/golang-blockchain/blockchain"
)

type CommandLine struct {
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("getBalance -address ADDRESS - get the balance for the address")
	fmt.Println("createBlockchain - adress ADDRESS create a blockchain and this address mines it's genesis block")
	fmt.Println("printchain - Prints the blocks in the chain")
	fmt.Println("send -from FROM -to TO -amount AMOUNT -Send amount")

}
func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit() // to shut down the database before the program terminates
	}
}

func (cli *CommandLine) printChain() {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	iter := chain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("Previous Hash %x\n", block.PrevHash)
		fmt.Printf("Hash %x\n", block.Hash)
		pow := blockchain.Newproof(block)
		fmt.Printf("Pow:%s \n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
func (cli *CommandLine) createBlockchain(address string) {
	chain := blockchain.InitBlockchain(address)
	chain.Database.Close()
	fmt.Println("finished")
}
func (cli *CommandLine) getBalance(address string) {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close() // only executes if there is proper completion of the program
	balance := 0
	UTXOs := chain.FindUTXO(address)
	for _, out := range UTXOs {
		balance += out.Value
	}
	fmt.Printf("Balance of %s %d \n", address, balance)

}
func (cli *CommandLine) send(from, to string, amount int) {
	chain := blockchain.ContinueBlockchain("")
	defer chain.Database.Close()
	tx := blockchain.NewTransaction(from, to, amount, chain)
	chain.AddBlock([]*blockchain.Transaction{tx})
	fmt.Printf("Success")

}
func (cli *CommandLine) Run() {
	cli.validateArgs()
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "address of the wallet")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "address of the miner")
	sendTo := sendCmd.String("to", "", "address of the wallet")
	sendFrom := sendCmd.String("from", "", "address of the wallet")
	sendAmount := sendCmd.Int("amount", 0, "Amount to send")

	switch os.Args[1] {
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockchain(*createBlockchainAddress)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount)
	}

}
