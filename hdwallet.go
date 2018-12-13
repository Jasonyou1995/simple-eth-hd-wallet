/*
 *	This is a simple hierarchical deterministc wallet for ether crytocurrency,
 *	and it is based on the BIP-32 (Bitcoin Improvement Proposal), BIP-39, etc.
 *
 *	Last modified date: 	December 12th 2018
 */



package hdwallet;

import (
	// built in packages
	"fmt"
	"errors"
	"sync"
	"math/big"

	// cryptography packages
	"crypto/ecdsa"
	"crypto/rand"

	// Bitcoin suite packages
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcutil/hdkeychain"	// for ExtendedKey type (BIP-32)

	// Ethereum packages
	ethereum "github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts"	// high level ETH account management
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"

	// BIP-39: Mnemonic-code-for-generating-deterministic-keys
	"github.com/tyler-smith/go-bip39"
)

/*

	The address structure: 'm/purpose'/coin-type'/account'/change/address_index'

	@" purpose' " is set to 44' (or 0x8000002C, since the quote ' in here means starting
	from 2^31, which represents the hardened child address) is following the BIP-43
	recommendation. Hardened derivation is used at this level.

	Note: The BIP-44 spec defines that the `purpose` be 44' (or 0x8000002C) for crypto
	currencies, and SLIP-44 assigns the `coin_type` 60' (or 0x8000003C) to Ethereum.

	@" coin-type' " is set to be 60' (or 0x8000003C) is the identifier for ether.
	Hardened sderivation is used at this level too.

	@" account' " level splits the key space into independent user identities, so the
	wallet never mixes the coins across different accounts. Hardened derivation is used
	at this level.

	@" change " is used to distinguish between external account (denote as 1), which is
	visable outside of the wallet to receive funds, or internal account (denote as 0), 
	which is only visible inside the wallet for transaction change. Public derivation
	is used at this level.

	@"address_index" is 0 based index used for child index in BIP-32 derivation. Public
	derivation is used at this level.

	@Note: the 'm/' prefix denote the full derivation path, and relative derivation path
	(which will get appended to the default root path) must not have prefixes in front
	of the first element. Whitespace is ignored.

*/

/*
	DefaultRootDerivationPath is the root path to which custom derivation endpoints are appended. As such, the first account will be at m/44'/60'/0'/0, the second at m/44'/60'/0'/1, etc.
*/
var DefaultRootDerivationPath = accounts.DefaultRootDerivationPath;

/*
	DefaultBaseDerivationPath is the base path from which custom derivation endpoints
	are incremented. As such, the first account will be at m/44'/60'/0'/0/0, the second
	at m/44'/60'/0'/0/1, etc.
*/
var DefaultBaseDerivationPath = accounts.DefaultBaseDerivationPath;

/*
	---------------- Define the Wallet structure ----------------

	FILEDS:

	@mnemonic: a string of mnemonic phrase separated by spaces.
	
	@masterKey: ExtendedKey houses all the information needed to support a hierarchical 
	deterministic extended key. See the package overview documentation for more 
	details on how to use extended keys. We can use the IsPrivate function in 
	hdkeychain package to determine whether an extended key is a private or
	public key.
	A private extended key can be used to derive both hardened and non-hardened
	(normal) child private and public extented key. Which can be used to prevent
	the case that exploiting one child key leads to the exploitation of all of
	its siblings.

	@seed: the byte array seed.

	@url: type URL struct {
    	Scheme string // Protocol scheme to identify a capable account backend
    	Path   string // Path for the backend to identify a unique entity
	} 
	// Note: it doesn't do any URL encoding/decoding of special characters.

	@paths: map from common go-ethereum address type to derivation path type.
		type DerivationPath []uint32
	5 levels of hierarchies: m / purpose' / coin_type' / account' / change / address_index
	Note: see the "address structure" note above.
		type Address [AddressLength]byte
	Note: Address represents the 20 byte address of an Ethereum account.

	@accounts: type Account struct {
	    Address common.Address `json:"address"` // Ethereum account address derived from the key
	    URL     URL            `json:"url"`     // Optional resource locator within a backend
	}

	@stateLock: reader/writer mutual exclusion lock that can be held by an arbitrary number
	of readers or a single writer. The zero value for a RWMutex is an unlocked mutex.
	Note: A RWMutex must not be copied after first use.

*/
type Wallet struct {
	mnemonic 	string;
	masterKey 	*hdkeychain.ExtendedKey;
	seed		[]byte;
	url			accounts.URL;
	paths		map[common.Address]accounts.DerivationPath;
	accounts	[]accounts.Account;
	stateLock	sync.RWMutex;
}

/*

	INPUT:
	@seed: the seed used for generating a new wallet.

	OUTPUT:
	Returns a pointer to the new wallet, or error code.

	PACKAGES:
	hdkeychain.NewMaster:	creates the root of the hieracrchical tree based on
							the cryptographically random seed. (We can also
							use the GenerateSeed function to generate random seed)

	chaincfg.MainNetParams: defines chain configuration parameters for the
							standard mainnet of Bitcoin (without -testnet).

							Params defines a Bitcoin network by its parameters. These parameters 
							may be used by Bitcoin applications to differentiate networks as well 
							as addresses and keys for one network from those intended for use on 
							another network.
	- source: https://github.com/btcsuite/btcd/blob/master/chaincfg/params.go#L225
	- reference: https://godoc.org/github.com/btcsuite/btcd/chaincfg#Params)


*/
func newWallet(seed []byte) (*Wallet, error) {
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {	return nil, err }

	return &Wallet {
		masterKey: 		masterKey,	// 
		seed:			seed,
		accounts:		[]accounts.Account{},
		paths:			map[common.Address]accounts.DerivationPath{},
	}, nil
}

/*

	INPUT:
	@mnemonic: the mnemonic phrases string seperated by whitespaces.

	OUTPUT:
	Returns a new wallet from a BIP-39 mnemonic.

*/
func NewFromMnemonic(mnemonic string) (*Wallet, error) {
	if (mnemonic == "") { 
		return nil, errors.New("Mnemonic string is empty, require one.")
	}
	if (!bip39.IsMnemonicValid(mnemonic)) {
		return nil, errors.New("Mnemonic string is invalid.")
	}
	
	// aquiring a new seed from the given mnemonic phrases
	seed, err := NewSeedFromMnemonic(mnemonic)
	if (err != nil) { return nil, err }

	// obtain a new wallet from the seed
	wallet, err := newWallet(seed)
	if (err != nil) { return nil, err }

	// set the mnemonic phrases of the new wallet
	wallet.mnemonic = mnemonic

	return wallet, nil
}

/*

	INPUT:
	@seed: an array of seeds used for generating a new wallet

	OUTPUT:
	Returns a new wallet based on the given seed

*/
func NewFromSeed(seed []byte) (*Wallet, error) {
	if (len(seed) == 0) {
		return nil, errors.New("Seed is empty, require one.")
	}
	return newWallet(seed);
}

// ---------------- Implementations for the account.Wallet interface ----------------
// Reference: 	https://godoc.org/github.com/ethereum/go-ethereum/accounts
// Receiver:	Wallet type pointer
// ----------------------------------------------------------------------------------

/*
	URL retrieves the canonical path under which this wallet is reachable.
	Used mainly on hardware devices.
*/
func (w *Wallet) URL() accounts.URL {
	return w.url;
}

/*
	Returning a custom status message to help the user in the current state
	of the wallet.
*/
func (w *Wallet) Status() (string, error) {
	return "ok", nil
}

/*
	Open the wallet instance and build access. Not used for unlocking or
	decrypting account keys.

	@passphrase:
	In the hard ware wallet implementation, the pass phrase in here is optional,
	and can be empty.

	Note: need to be manually closed to free the allocated resources, esp.
	for hardware wallets.
*/
func (w *Wallet) Open(passphrase string) (error) {
	return nil
}

/*
	Close an opened wallet and release any used resources by the instance.
*/
func (w *Wallet) Close() error {
	return nil
}

/*
	Retrieves the list of signing accounts the wallet currently holds. For
	the hierarchical deterministic wallets, this list will only display the
	accounts be explicitly pinned during account derivation.
*/
func (w *Wallet) Accounts() []accounts.Account {
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()		// executed untile this function returned

	// copy and return the slice of the current accounts
	cpy := make([]accounts.Account, len(w.accounts))
	copy(cpy, w.accounts)
	return cpy
}

/*
	Returns whether the given account is contained in this particular wallet.
*/
func (w *Wallet) Contains(account accounts.Account) bool {
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()

	// second returned argument denotes the existenceness of the address
	_, exists := w.paths[account.Address]
	return exists
}

/*
	Helper function to unpins the account from list of pinned accounts
*/
func (w *Wallet) Unpin(account accounts.Account) error {
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()

	for i, acct := range w.accounts {
		if acct.Address.String() == account.Address.String() {
			// remove the account first
			// using variadic parameter for the second argument
			w.accounts = append(w.accounts[:i], w.accounts[i + 1:]...)
			// delete address from the map w.paths
			delete(w.paths, account.Address)
			return nil
		}
	}

	return errors.New("Account not found.")
}

/*
	@path: 	the path to be used to derive a hierarchical deterministic account.
	@pin:	determine if this new account will be pinned to the wallet's tracking
			account list.
*/
func (w *Wallet) Derive(path accounts.DerivationPath, pin bool) (accounts.Account, error) {
	w.stateLock.RLock()		// keeping the device exist
	address, err := w.deriveAddress(path)
	w.stateLock.RUnlock()

	if err != nil { return accounts.Account{}, err }

	account := accounts.Account {
		Address: address,
		URL: accounts.URL {
			Scheme: 	"",
			Path: 		path.String(),		// the comma in here is necessary for multi-line
		},
	}

	if !pin {
		// return the account now if no pin required
		return account, nil
	} else {
		// pinning and avoid concurrency bugs caused by state change
		// Lock locks rw for writing
		w.stateLock.Lock()
		defer w.stateLock.Unlock()

		_, ok := w.paths[address];
		if !ok {
			// not pinned yet
			w.accounts = append(w.accounts, account)
			w.paths[address] = path
		}
		return account, nil
	}
}

/*
	@path: 	the path to be used to derive a hierarchical deterministic account.
	@chain:	 ChainStateReader wraps access to the state trie of the canonical blockchain.
	Reference to the interface:
		https://github.com/ethereum/go-ethereum/blob/master/interfaces.go

	Sets a base account derivation path from which the wallet tries to discover
	non zero accounts and automatically add them to the list of tracking accounts.
*/
func (w *Wallet) SelfDerive(base accounts.DerivationPath, chain ethereum.ChainStateReader) {
	// TODO
}

/*
	Requests the wallet to sign the given hash with the account.
*/
func (w *Wallet) SignHash(account accounts.Account, hash []byte) ([]byte, error) {
	path, ok := w.paths[account.Address]
	if !ok { return nil, accounts.ErrUnknownAccount }

	privateKey, err := w.derivePrivateKey(path)
	if err != nil { return nil, err }

	return crypto.Sign(hash, privateKey)
}

/*
	Using the account and chainID to sign the given transaction tx.
*/
func (w *Wallet) SignTx(account accounts.Account, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	w.stateLock.RLock()
	defer w.stateLock.RUnlock()

	path, ok := w.paths[account.Address]
	if !ok { return nil, accounts.ErrUnknownAccount }

	privateKey, err := w.derivePrivateKey(path)
	if err != nil { return nil, err }

	// Sign the transaction
	signedTx, err := types.SignTx(tx, types.HomesteadSigner{}, privateKey)
	if err != nil { return nil, err }

	msg, err := signedTx.AsMessage(types.HomesteadSigner{})
	if err != nil { return nil, err }

	sender := msg.From()
	if sender != account.Address {
		return nil, fmt.Errorf("Wrong sender: want %s, got %s", account.Address.Hex(), sender.Hex())
	}

	return signedTx, nil
}

/*
	Request the wallet to sign the given hash with the account.
	Using the passphrase as an extra layer of authetication information.
*/
func (w *Wallet) SignHashWithPassphrase(account accounts.Account, passphrase string, hash []byte) ([]byte, error) {
	// TODO:	passphrase will be included in the hash derivation function (KDF) and be hashed
	//			for 262,144 times to prevent brute force attack.

	return w.SignHash(account, hash)
}

/*
	Requests the wallet to sign the given transaction with the given passphrase
	as an extra layer of authetication information.
*/
func (w *Wallet) SignTxWithPassphrase(account accounts.Account, passphrase string, tx *types.Transaction, chainID *big.Int) (*types.Transaction, error) {
	// TODO

	return w.SignTx(account, tx, chainID)
}

// -----------------------------------------------------------------------
// ----------------------- More Helper Functions -------------------------
// -----------------------------------------------------------------------

/*

	How are Ether addresses, public keys, and private keys generated?

	Tools:	Keccak-256sum (https://github.com/maandree/sha3sum), ECDSA

	Three stages:


	Reference:
	https://kobl.one/blog/create-full-ethereum-keypair-and-address/
	https://ethereum.stackexchange.com/questions/3542/how-are-ethereum-addresses-generated

*/

/*
	@account: using this account to derivate the private key.
	Obtain the private key through the ECDSA (Elliptic Curve Digital Signature Algorithm)
*/
func (w *Wallet) PrivateKey(account accounts.Account) (*ecdsa.PrivateKey, error) {
	path, err := accounts.ParseDerivationPath(account.URL.Path)
	if err != nil { return nil, err }
	return w.derivePrivateKey(path)
}

/*
	@account: using this account to derivate the private key.
	Returns the ECDSA private key in bytes format.
*/
func (w *Wallet) PrivateKeyBytes(account accounts.Account) ([]byte, error) {
	privateKey, err := w.PrivateKey(account)
	if err != nil { return nil, err }

	// exports a private key into a binary dump
	return crypto.FromECDSA(privateKey), nil
}

/*
	@account: using this account to derivate the private key.
	Returns the ECDSA private key in hexadecimal format.
*/
func (w *Wallet) PrivateKeyHex(account accounts.Account) (string, error) {
	var privateKeyHex string   	// for result storage
	privateKeyBytes, err := w.PrivateKeyBytes(account)
	if err != nil { return "", err }


	privateKeyHex = hexutil.Encode(privateKeyBytes)

	// encodes b as a hex string with 0x prefix, so we need to removed the prefix.
	privateKeyHex = privateKeyHex[2:]
	if privateKeyHex[2] == '0' && privateKeyHex[3] == '0' {
		// remove the first null byte if it starts with '00'
		privateKeyHex = privateKeyHex[2:]
	}
	return privateKeyHex, nil
}

/*
	Returns the public key of the account based on the ECDSA.
*/
func (w *Wallet) PublicKey(account accounts.Account) (*ecdsa.PublicKey, error) {
	path, err := ParseDerivationPath(account.URL.Path)
	if err != nil { return nil, err }
	return w.derivePublicKey(path)
}

/*
	Returns the public key of the account based on the ECDSA in bytes format.
*/
func (w *Wallet) PublicKeyBytes(account accounts.Account) ([]byte, error) {
	publicKey, err := w.PublicKey(account)
	if err != nil { return nil, err }
	return crypto.FromECDSAPub(publicKey), nil
}

/*
	Returns the public key of the account based on the ECDSA in hexadecimal format.
*/
func (w *Wallet) PublicKeyHex(account accounts.Account) (string, error) {
	publicKeyBytes, err := w.PublicKeyBytes(account)
	if err != nil { return "", err }

	// Note: 	Public key must be uncompressed (no whitespace or colon) and 64 bytes long.
	// 			Every EC public key begins with the 0x04 prefix before giving the location
	// 			of the two points on the curve. We need to remove this leading 0x04 byte
	// 			in order to hash it correctly.
	return hexutil.Encode(publicKeyBytes)[4:], nil
}

/*
	Return the address of the account, which is derived from the public key.
*/
func (w *Wallet) Address(account accounts.Account) (common.Address, error) {
	publicKey, err := w.PublicKey(account)
	if err != nil { return common.Address{}, err }
	return crypto.PubkeyToAddress(*publicKey), nil
}

/*
	Return the address in byte format.
*/
func (w *Wallet) AddressBytes(account accounts.Account) ([]byte, error) {
	address, err := w.Address(account)
	if err != nil { return nil, err }
	return address.Bytes(), nil
}

/*
	Return the address in hexadecimal format.
*/
func (w *Wallet) AddressHex(account accounts.Account) (string, error) {
	address, err := w.Address(account)
	if err != nil { return "", err }
	return address.Hex(), nil
}

/*
	Return the derivation path of the given account
*/
func (w *Wallet) Path(account accounts.Account) (string, error) {
	return account.URL.Path, nil
}

/*
	Convert the derivation path string to []uint32
*/
func ParseDerivationPath(path string) (accounts.DerivationPath, error) {
	return accounts.ParseDerivationPath(path)
}

/*
	Same as the ParseDerivationPath(path string), but will be panic to any error
*/
func StrictParseDerivationPath(path string) accounts.DerivationPath {
	parsed, err := accounts.ParseDerivationPath(path)
	if err != nil { panic(err) }
	return parsed
}

/*

	Returns a randomly generated mnemonic phrase based on BIP-39.
	The following table describes the relation between the initial
	entropy length (ENT), which range from 128 to 256, the checksum
	length (CS) and the length of the generated mnemonic sentence (MS)
	in words.

	CS = ENT / 32
	MS = (ENT + CS) / 11

	|  ENT  | CS | ENT+CS |  MS  |
	+-------+----+--------+------+
	|  128  |  4 |   132  |  12  |
	|  160  |  5 |   165  |  15  |
	|  192  |  6 |   198  |  18  |
	|  224  |  7 |   231  |  21  |
	|  256  |  8 |   264  |  24  |

	References:
	BIP-39:
		https://github.com/bitcoin/bips/blob/master/bip-0039.mediawiki
	Wordlist:
		https://github.com/bitcoin/bips/blob/master/bip-0039/bip-0039-wordlists.md
		Note: 	indexed from 1 to 2^11 = 2048 Mnemonic phrases options for eight
				different langauges

*/
func NewMnemonic(bits int) (string, error) {
	entropy, err := bip39.NewEntropy(bits)
	if err != nil { return "", err }
	return bip39.NewMnemonic(entropy)
}

/*
	Return a randomly generated seed based on BIP-39
*/
func NewSeed() ([]byte, error) {
	b := make([]byte, 64)		// create a slice of byte with length 64
	_, err := rand.Read(b)
	return b, err
}

/*
	Returns a new seed from BIP-39 mnemonic phrases.
*/
func NewSeedFromMnemonic(mnemonic string) ([]byte, error) {
	if (mnemonic == "") {
		return nil, errors.New("Mnemonic is empty, require one.")
	}
	return bip39.NewSeedWithErrorChecking(mnemonic, "")
}

/*
	Derives the private key based on the derivation path
*/
func (w *Wallet) derivePrivateKey(path accounts.DerivationPath) (*ecdsa.PrivateKey, error) {
	var err error
	key := w.masterKey
	for _, n := range path {
		key, err = key.Child(n)
		if err != nil { return nil, err }
	}

	privateKey, err := key.ECPrivKey()
	if err != nil { return nil, err }

	return privateKey.ToECDSA(), nil
}

/*
	Deriving the public key based on the derivation path
*/
func (w *Wallet) derivePublicKey(path accounts.DerivationPath) (*ecdsa.PublicKey, error) {
	privateKeyECDSA, err := w.derivePrivateKey(path)
	if err != nil { return nil, err }

	publicKey := privateKeyECDSA.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)	// non-panic ECDSA interface Type assertion
	if !ok { return nil, errors.New("Failed to obtain an ECDSA public key.") }

	return publicKeyECDSA, nil
}

/*
	Deriving the account address based on the derivation path and the public key.
*/
func (w *Wallet) deriveAddress(path accounts.DerivationPath) (common.Address, error) {
	publicKeyECDSA, err := w.derivePublicKey(path)
	if err != nil { return common.Address{}, err }

	address := crypto.PubkeyToAddress(*publicKeyECDSA)
	return address, nil
}

/*
	Removing an account from the given index.
*/
func removeAtIndex(accts []accounts.Account, index int) []accounts.Account {
	// using variadic parameter for the second arg
	return append(accts[:index], accts[index + 1:]...)
}
