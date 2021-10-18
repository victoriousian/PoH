package main

import (
	"flag"
	"fmt"
	"os"
)

type CLI struct{}

func (c *CLI) list() {
	bc := NewBlockchain()
	defer bc.db.Close()

	bc.List()
}

func (c *CLI) createBlockchain(address string) {
	bc := CreateBlockchain(address)
	bc.db.Close()
}

func (c *CLI) send(value uint64, from, to string) {
	bc := NewBlockchain()
	defer bc.db.Close()

	tx := bc.Send(value, from, to)
	bc.AddBlock([]*Transaction{tx})
}

func (c *CLI) getBalance(address string) uint64 {
	bc := NewBlockchain()
	defer bc.db.Close()

	return bc.GetBalance(address)
}

func (c *CLI) newWallet() string {
	return NewKeyStore().CreateWallet().GetAddress()
}

func (c *CLI) Run() {
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	newWalletCmd := flag.NewFlagSet("newwallet", flag.ExitOnError)

	newAddress := newCmd.String("address", "", "")

	sendValue := sendCmd.Uint64("value", 0, "")
	sendFrom := sendCmd.String("from", "", "")
	sendTo := sendCmd.String("to", "", "")

	getBalanceAddress := getBalanceCmd.String("address", "", "")

	switch os.Args[1] {
	case "new":
		newCmd.Parse(os.Args[2:])
	case "send":
		sendCmd.Parse(os.Args[2:])
	case "getbalance":
		getBalanceCmd.Parse(os.Args[2:])
	case "newwallet":
		newWalletCmd.Parse(os.Args[2:])
	case "list":
		listCmd.Parse(os.Args[2:])
	default:
		os.Exit(1)
	}

	if newCmd.Parsed() {
		if *newAddress == "" {
			newCmd.Usage()
			os.Exit(1)
		}
		c.createBlockchain(*newAddress)
	}
	if sendCmd.Parsed() {
		if *sendValue == 0 || *sendFrom == "" || *sendTo == "" {
			sendCmd.Usage()
			os.Exit(1)
		}
		c.send(*sendValue, *sendFrom, *sendTo)
	}
	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			os.Exit(1)
		}
		fmt.Printf("Balance of '%s': %d\n", *getBalanceAddress, c.getBalance(*getBalanceAddress))
	}
	if newWalletCmd.Parsed() {
		fmt.Printf("Address: %s", c.newWallet())
	}

	if listCmd.Parsed() {
		c.list()
	}
}