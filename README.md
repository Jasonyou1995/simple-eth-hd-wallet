# Simple Hierarchical Deterministic Wallet For Ether (Go Implementation)

## Before get started

1. Please first [download go](https://golang.org/dl/) for testing.
2. Download required packages (downloaded in terminal for Unix like system): `source download-packages.sh`
3. Run tests: `go test`
4. Run some demos in the `./demo/` repository with `go run ./demo/<demo-name.go>`


## Very Simple Wallet and Hierarchical Deterministic (HD) Wallet Introduction

### Purpose of Wallet

* Used for storing the *public* and *private* key pairs of different cryptocurrencies.
* Wallet can also be the platform or interface that be used for interacting with Decentralized Applications (DAPPs).
	* This is accomplished by creating, signing and broadcasting transactions that include smart contracts (or chaincode as called in Hyperledger) to update the states of decentralized applications (DAPPs) on some state machines built on blockchain, such as the *EVM* we used in Ethereum blockchain platform.
	* The Ethereum Virtual Machine (EVM) is a 256-bit Turing complete Virtual Machine that allows anyone to run Byte Code on it, which is also part of the Ethereum Protocal.

### Wallet Types

* Non-deterministic/Random Wallet
	* Pros: Very secure, since it will generate a new address and account for each new transaction.
	* Cons: But this is also a drawback when users need to backup the wallet: there are just too many accounts users need to manage and backup. And thus in practice, we normally never use a Non-deterministic wallet.
* Hierarchical Deterministic (HD) Wallet
	* This is most used wallet type in the market, and it is called Hierarchical because of the level in its account structure:
		* 5-level path: `m/purpose'/coin-type'/account'/change/address_index`
	* Mnemonic code: these are `2^11 = 2048` phrases that is used to derive private keys, and it now supports multiple languages: [BIP-39 Wordlists](https://github.com/bitcoin/bips/blob/master/bip-0039/bip-0039-wordlists.md). You can find more information in the resources section.
	* BIP-32-39-43-44 provides an excellent guide line about how to design a HD-Wallet, here I will cover the based steps to generate Mnemonic Words and then generate new seed based on that (refer to [Mastering Ethereum](https://www.amazon.com/s/?ie=UTF8&keywords=mastering+ethereum&tag=googhydr-20&index=aps&hvadid=241643286910&hvpos=1t1&hvnetw=g&hvrand=17035223864127781598&hvpone=&hvptwo=&hvqmt=e&hvdev=c&hvdvcmdl=&hvlocint=&hvlocphy=9016722&hvtargid=kwd-278823743929&ref=pd_sl_2fo8ttzbc9_e) Chapter 5).

#### Generating Mnemonic Words

1. Create a cryptographically random sequence S of 128 to 256 bits long.
2. Create a checksum of S by taking the first `length-of-S / 32` bits of the SHA-256 hash of S.
3. Append this checksum to the end of the random sequence S.
4. Divide the sequence-and-checksum concatenation into sections of 11 bits (each block has 11-bit).
5. Map each of the 11-bit block/value to a phrase in the [mnemonic dictionary](https://github.com/bitcoin/bips/blob/master/bip-0039/bip-0039-wordlists.md) (2,048 words).
6. Create the mnemonic code from the sequence of words, and maintaining the order.

#### From Mnemonic To Seed

1. Using the output the the **step 6** from the *Generating Mnemonic Words* section as the input to the *PBKDF2* key-stretching function.
2. The second parameter to the PBKDF2 key-stretching function is a salt. The salt is composed of the string constant "mnemonic" concatenated with an optional **user-supplied passphrase**. 
	3. That's why we need to be very cautious about remembering the passphrase used to generate seed, cause it is included in each steps of cryptographical key generation. If we forgot it, then no body can help us to retrive the fund.
4. PBKDF2 stretches the mnemonic and salt parameters using 2,048 rounds of hashing with the HMAC-SHA512 algorithm, producing a 512-bit value as its final output. That 512-bit value is the seed.

#### The Brain Wallet...

Normally all words are created randomly and presented to user after generated. (normally 12-24 words based on 128-256 bits of entropy. But *Brain Wallet* let users to generate the mnemonic words by themselve, which is obviously not a good idea. So we don't use brain wallet in real products normally.


### The Art of Backup

* The ideal goal of backing up is to never lose one bit of coin by losing devise or access to private keys, and at the same time prevent attackers gain any access to the private key or fund in the account.
* In the Chapter 5 Wallet of [Mastering Ethereum](https://www.amazon.com/s/?ie=UTF8&keywords=mastering+ethereum&tag=googhydr-20&index=aps&hvadid=241643286910&hvpos=1t1&hvnetw=g&hvrand=17035223864127781598&hvpone=&hvptwo=&hvqmt=e&hvdev=c&hvdvcmdl=&hvlocint=&hvlocphy=9016722&hvtargid=kwd-278823743929&ref=pd_sl_2fo8ttzbc9_e) the authors covered a lot about good practice to backup wallet and some issues to consider when designing a wallet.


## More resources

### Related Bitcoin Improvement Proposals (BIPs)

* [BIP-32](https://github.com/bitcoin/bips/blob/master/bip-0032.mediawiki): Hierarchical Deterministic Wallets
* [BIP-39](https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki): Mnemonic code for generating deterministic keys
	* [BIP-39 Wordlists](https://github.com/bitcoin/bips/blob/master/bip-0039/bip-0039-wordlists.md)
* [BIP-43](https://github.com/bitcoin/bips/blob/master/bip-0043.mediawiki): Purpose Field for Deterministic Wallets
* [BIP-44](https://github.com/bitcoin/bips/blob/master/bip-0044.mediawiki): Multi-Account Hierarchy for Deterministic Wallets
* [BIP-70](https://github.com/bitcoin/bips/blob/master/bip-0070.mediawiki): Payment Protocol
* [BIP-75](https://github.com/bitcoin/bips/blob/master/bip-0075.mediawiki): Out of Band Address Exchange using Payment Protocol Encryption

### Github Repos

* [btcsuit/btcd](https://github.com/btcsuite/btcd/blob/master/chaincfg/params.go#L225)

### GoDoc

* [btcsuit/btcd](https://godoc.org/github.com/btcsuite/btcd/chaincfg#Params)
* [go-ethereum/accounts](https://godoc.org/github.com/ethereum/go-ethereum/accounts)
* [go-ethereum/blob/master/interface.go](https://github.com/ethereum/go-ethereum/blob/master/interfaces.go)

### Cryptography Tools And Blogs

* [Keccak-256sum](https://github.com/maandree/sha3sum)
* [Create Full Ethereum Keypair And Address](https://kobl.one/blog/create-full-ethereum-keypair-and-address/)
* [How Are Ethereum Address Generated](https://ethereum.stackexchange.com/questions/3542/how-are-ethereum-addresses-generated)


### Books
* [Mastering Ethereum](https://www.amazon.com/s/?ie=UTF8&keywords=mastering+ethereum&tag=googhydr-20&index=aps&hvadid=241643286910&hvpos=1t1&hvnetw=g&hvrand=17035223864127781598&hvpone=&hvptwo=&hvqmt=e&hvdev=c&hvdvcmdl=&hvlocint=&hvlocphy=9016722&hvtargid=kwd-278823743929&ref=pd_sl_2fo8ttzbc9_e)

