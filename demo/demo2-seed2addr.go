package main

import (
	"fmt"
	"log"

    "github.com/Jasonyou1995/simple-eth-hd-wallet"
)

func main() {
	rand_seed, _ := hdwallet.NewSeed()
	wallet, err := hdwallet.NewFromSeed(rand_seed)
	if err != nil {		log.Fatal(err)	}
	
	// show the seed
	fmt.Println("Random Seed:")
	fmt.Print("[ ")
	for _, v := range rand_seed {
		fmt.Print(v, " ")
	}
	fmt.Println("]")
	fmt.Println()
	
	p1 := "m/44'/60'/0'/0/0"
	path := hdwallet.StrictParseDerivationPath(p1)
	account, err := wallet.Derive(path, false)
	if err != nil {		log.Fatal(err)	}
	
	fmt.Printf("Path: %s\nAddress: %s\n\n", p1, account.Address.Hex())
	
	p2 := "m/44'/60'/0'/0/9"
	path = hdwallet.StrictParseDerivationPath(p2)
	account, err = wallet.Derive(path, false)
	if err != nil {		log.Fatal(err)	}
	
	fmt.Printf("Path: %s\nAddress: %s\n\n", p2, account.Address.Hex())
}
