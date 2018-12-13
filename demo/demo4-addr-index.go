package main

import (
	"github.com/Jasonyou1995/simple-eth-hd-wallet"
	"fmt"
)

func main() {
    mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	wallet, err := hdwallet.NewFromMnemonic(mnemonic)
	if err != nil {
		panic(err)
	}

	path, err := hdwallet.ParseDerivationPath("m/44'/60'/0'/0/0")
	if err != nil {
		panic(err)
	}

	account, err := wallet.Derive(path, false)
	if err != nil {
		panic(err)
	}

	if account.Address.Hex() != "0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947" {
		panic("wrong address")
	} else {
		path1, _ := wallet.Path(account)
		fmt.Println("Successfully generated address1 from path1:")
		fmt.Printf("\tPath:\t\t %s \n\tAddress:\t %s\n", path1, account.Address.Hex())
	}

	path2, err := hdwallet.ParseDerivationPath("m/44'/60'/0'/0/9")
	if err != nil {
		panic(err)
	}

	account2, err := wallet.Derive(path2, false)
	if err != nil {
		panic(err)
	}

	if account2.Address.Hex() != "0x2d69B45301b9B3E01c4797C7a48BBc7e7F9b355b" {
		panic("wrong address")
	} else {
		path2, _ := wallet.Path(account2)
		fmt.Println("Successfully generated address2 from path2:")
		fmt.Printf("\tPath2:\t\t %s \n\tAddress2:\t %s\n", path2, account2.Address.Hex())
	}
}
