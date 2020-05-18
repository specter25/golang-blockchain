package main

import (
	"flag"
	"fmt"
	"os"
	"runtime"
	"strconv"

	"github.com/specter25/golang-blockchain/blockchain"
)

type CommandLine struct {
	blockchain *blockchain.Blockchain
}

func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("add -block BLOCK_DATA -- add a block to the blockchain")
	fmt.Println("Print  - Prints the block in th blockchain")

}
func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit() // to shut down the database before the program terminates
	}
}
func (cli *CommandLine) addBlock(data string) {
	cli.blockchain.AddBlock(data)
	fmt.Println("Added Block!")
}
func (cli *CommandLine) printChain() {
	iter := cli.blockchain.Iterator()
	for {
		block := iter.Next()
		fmt.Printf("Previous Hash %x\n", block.PrevHash)
		fmt.Printf("Data %s\n", block.Data)
		fmt.Printf("Hash %x\n", block.Hash)
		pow := blockchain.Newproof(block)
		fmt.Printf("Pow:%s \n", strconv.FormatBool(pow.Validate()))
		fmt.Println()
	}
}
func (cli *CommandLine) run() {
	cli.validateArgs()
	addBlockCmd := flag.NewFlagSet("add", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("print", flag.ExitOnError)
	addBlockData := addBlockCmd.String("block", "", "BlockData")
	switch os.Args[1] {
	case "add":
		err := addBlockCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	case "print":
		err := printChainCmd.Parse(os.Args[2:])
		blockchain.Handle(err)
	default:
		cli.printUsage()
		runtime.Goexit()
	}
	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			runtime.Goexit()
		}
		cli.addBlock(*addBlockData)
	}
	if printChainCmd.Parsed() {
		cli.printChain()
	}

}

func main() {
	defer os.Exit(0)
	chain := blockchain.InitBlockchain()
	defer chain.Database.Close() // only executes if there is proper completion of the program
	cli := CommandLine{chain}
	cli.run()

}
