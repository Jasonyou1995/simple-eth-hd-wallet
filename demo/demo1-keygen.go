package main

import (
    "fmt"
    "log"

    "github.com/Jasonyou1995/simple-eth-hd-wallet"
)

func main() {

    mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"

    eth_path := "m/44'/60'/0'/0/0"
    
    wallet, err := hdwallet.NewFromMnemonic(mnemonic)
    if err != nil { log.Fatal(err) }

    path := hdwallet.StrictParseDerivationPath(eth_path)
    account, err := wallet.Derive(path, false)
    if err != nil { log.Fatal(err) }

    fmt.Printf("Account address: %s\n", account.Address.Hex())

    privateKey, err := wallet.PrivateKeyHex(account)
    if err != nil { log.Fatal(err) }

	fmt.Printf("Private key in hex: %s\n", privateKey)

	publicKey, _ := wallet.PublicKeyHex(account)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Public key in hex: %s\n", publicKey)

}
