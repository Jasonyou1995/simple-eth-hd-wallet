/*
	Includes table test for the address, private and public keys generation.
*/

package hdwallet

import (
	"math/big"
	"strings"
	"testing"
	"strconv"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
)

/*
	Note: 	for the Table Test, I used the Mnemonic Converter as reference
			[Mnemonic Converter website: https://iancoleman.io/bip39/]
*/
func TestWalletTable(t *testing.T) {
	// Testing different address index
	mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"

	// create a new wallet
	wallet, err := NewFromMnemonic(mnemonic)
	if err != nil {		t.Error(err)	}

	// generate a bunch of addresses with different address indices
	testNumber := 10					// number of accounts to test

	// saving slices of valid addresses, private and public keys
	validAddrs := make([]string, testNumber)
	validPrivKeys := make([]string, testNumber)
	validPubKeys := make([]string, testNumber)

	validAddrs = []string{"0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947",
							"0x8230645aC28A4EdD1b0B53E7Cd8019744E9dD559",
							"0x65c150B7eF3B1adbB9cB2b8041C892b15eDde05A",
							"0x1AebbE69459B80d4975259378577Bc01d2924Cf4",
							"0x32f48bf54DBbfCE73172e69FE563c130d536cd5f",
							"0x1c255DB352E8B3CC16EFd721C61d7b1B5952b2bb",
							"0x1a41029aEb54A8C09211539b92B2A3fD92Ea8270",
							"0x54C0897A1E281b107EeE25d4f8eEe5F6ae13f9d9",
							"0x3d503E7c3799ab9478b6c04623275fdC0ad09B1e",
							"0x2d69B45301b9B3E01c4797C7a48BBc7e7F9b355b"}

	validPrivKeys = []string{"63e21d10fd50155dbba0e7d3f7431a400b84b4c2ac1ee38872f82448fe3ecfb9",
							"b31048b0aa87649bdb9016c0ee28c788ddfc45e52cd71cc0da08c47cb4390ae7",
							"d5c561f92921a5d7eb8a91cc81cb392d1877dcc6b856260c1676cb28ef7203b0",
							"f466f6f4d2d61a11eddd10eb80aae500c7601539d08d1d55f9e5efe25ecf95bc",
							"103a9ef39d4ced2988f1d5084460ebf8ea3baea2b2ca78265b637e48d99dce82",
							"1a69b812ca32e38bcac5197a63f6c1a1fcb6ac202e524382565cef16f1b3c84c",
							"83d5a75675cc8f1be09c7d4189117fe33ee3f09d1f9b5783140f03016a35b132",
							"526db1890baf94e82162f17f25ad769eb7f981272d8d99c527ea1af443c2d0cc",
							"cae7ce30e8e07507988d43ad8907edea2fd23f848fb1b8522dee53cac43a825f",
							"7525a4c5f03fb0b22fd88862e23833d62719b609e32a9264f6e437d56520d375"}
	// using only the x-axis value for the public key and not the (x,y) point
	validPubKeys = []string{"6005c86a6718f66221713a77073c41291cc3abbfcd03aa4955e9b2b50dbf7f9b",
							"3bea344870200a06bfad8f27ceb9f81746e1c659d6c6dd427a7b9b424e224f28",
							"b97419271c674a1e593e6cc312bf5bbbd98cc8b6f89f9aeb41449d06931029a2",
							"dac726c391d6990b1c64218cf05107e7893d744ef174b36ebc3df7c469fcabf8",
							"bf26f8038c5a9026ba68d9a19593e94db2d8ecd4ca1451e6022787751c8615f9",
							"974dc5809c2c7ef5739ee1f0abbac7d9a2333f965671f8b1c459ee4ce1c5d667",
							"a5927f446136d57dc3ae1b35617633c80bd23f5fe6625ead93c949382c63cc2c",
							"c1827492fda9b42852d2aa5745c6bbb76bad233c089a6429f53af4325d1a8042",
							"173385c77cc2812f4bec788db897850abd9869b900b1765266206229093b44d7",
							"59e5bfc671e89b9b48e334a015a0cfad6245358a532c7dea6305a7f2e98bde46"}

	// generating test addresses and comparing with the valid addresses
	for i := 0; i < testNumber; i++ {
		// extract the valid addresses and keys
		validAddress := validAddrs[i]
		validPrivKey := validPrivKeys[i]
		validPubKey := validPubKeys[i]

		// generating the path, address, private and public keys
		path, err := ParseDerivationPath( "m/44'/60'/0'/0/" + strconv.Itoa(i) )
		if err != nil {		t.Error(err)	}

		account, err := wallet.Derive(path, false)
		if err != nil {		t.Error(err)	}
		address := account.Address.Hex()

		privateKeyHex, err := wallet.PrivateKeyHex(account)
		if err != nil {		t.Error(err)	}

		publicKeyHex, err := wallet.PublicKeyHex(account)
		if err != nil {		t.Error(err)	}

		// start the testing
		if address != validAddress {
			t.Errorf("Invalid address with index %d:\n\t[got: %s, want: %s]\n",
						i, address, validAddress)
		}

		if privateKeyHex != validPrivKey {
			t.Errorf("Invalid private key with index %d:\n\t[got: %s, want: %s]\n",
				i, privateKeyHex, validPrivKey)
		}


		if publicKeyHex[:64] != validPubKey {
			// testing the x-axis value of the public key is enough
			t.Errorf("Invalid public key with index %d:\n\t[got: %s, want: %s]\n",
				i, publicKeyHex[:64], validPubKey)
		}
	}
}

/*
	General functions and methods tests
*/
func TestWallet(t *testing.T) {
	mnemonic := "tag volcano eight thank tide danger coast health above argue embrace heavy"
	wallet, err := NewFromMnemonic(mnemonic)
	if err != nil {
		t.Error(err)
	}

	path, err := ParseDerivationPath("m/44'/60'/0'/0/0")
	if err != nil {
		t.Error(err)
	}

	account, err := wallet.Derive(path, false)
	if err != nil {
		t.Error(err)
	}

	if account.Address.Hex() != "0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947" {
		t.Error("wrong address")
	}

	if len(wallet.Accounts()) != 0 {
		t.Error("expected 0")
	}

	account, err = wallet.Derive(path, true)
	if err != nil {
		t.Error(err)
	}

	if len(wallet.Accounts()) != 1 {
		t.Error("expected 1")
	}

	if !wallet.Contains(account) {
		t.Error("expected to contain account")
	}

	url := wallet.URL()
	if url.String() != "" {
		t.Error("expected empty url")
	}

	if err := wallet.Open(""); err != nil {
		t.Error(err)
	}

	if err := wallet.Close(); err != nil {
		t.Error(err)
	}

	status, err := wallet.Status()
	if err != nil {
		t.Error(err)
	}

	if status != "ok" {
		t.Error("expected status ok")
	}

	accountPath, err := wallet.Path(account)
	if err != nil {
		t.Error(err)
	}

	if accountPath != `m/44'/60'/0'/0/0` {
		t.Error("wrong hdpath")
	}

	privateKeyHex, err := wallet.PrivateKeyHex(account)
	if err != nil {
		t.Error(err)
	}

	if privateKeyHex != "63e21d10fd50155dbba0e7d3f7431a400b84b4c2ac1ee38872f82448fe3ecfb9" {
		t.Error("wrong private key")
	}

	publicKeyHex, err := wallet.PublicKeyHex(account)
	if err != nil {
		t.Error(err)
	}

	if publicKeyHex != "6005c86a6718f66221713a77073c41291cc3abbfcd03aa4955e9b2b50dbf7f9b6672dad0d46ade61e382f79888a73ea7899d9419becf1d6c9ec2087c1188fa18" {
		t.Error("wrong public key")
	}

	addressHex, err := wallet.AddressHex(account)
	if err != nil {
		t.Error(err)
	}

	if addressHex != "0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947" {
		t.Error("wrong address")
	}

	nonce := uint64(0)
	value := big.NewInt(1000000000000000000)
	toAddress := common.HexToAddress("0x0")
	gasLimit := uint64(21000)
	gasPrice := big.NewInt(21000000000)
	data := []byte{}

	tx := types.NewTransaction(nonce, toAddress, value, gasLimit, gasPrice, data)

	signedTx, err := wallet.SignTx(account, tx, nil)
	if err != nil {
		t.Error(err)
	}

	v, r, s := signedTx.RawSignatureValues()
	if v.Cmp(big.NewInt(0)) != 1 {
		t.Error("expected v value")
	}
	if r.Cmp(big.NewInt(0)) != 1 {
		t.Error("expected r value")
	}
	if s.Cmp(big.NewInt(0)) != 1 {
		t.Error("expected s value")
	}

	signedTx2, err := wallet.SignTxWithPassphrase(account, "", tx, nil)
	if err != nil {
		t.Error(err)
	}
	if signedTx.Hash() != signedTx2.Hash() {
		t.Error("expected match")
	}

	data = []byte("hello")
	hash := crypto.Keccak256Hash(data)
	sig, err := wallet.SignHash(account, hash.Bytes())
	if err != nil {
		t.Error(err)
	}
	if len(sig) == 0 {
		t.Error("expected signature")
	}

	sig2, err := wallet.SignHashWithPassphrase(account, "", hash.Bytes())
	if err != nil {
		t.Error(err)
	}
	if len(sig2) == 0 {
		t.Error("expected signature")
	}
	if hexutil.Encode(sig) != hexutil.Encode(sig2) {
		t.Error("expected match")
	}

	err = wallet.Unpin(account)
	if err != nil {
		t.Error(err)
	}

	if wallet.Contains(account) {
		t.Error("expected to not contain account")
	}

	// seed test

	seed, err := NewSeedFromMnemonic(mnemonic)
	if err != nil {
		t.Error(err)
	}

	wallet, err = NewFromSeed(seed)
	if err != nil {
		t.Error(err)
	}

	if account.Address.Hex() != "0xC49926C4124cEe1cbA0Ea94Ea31a6c12318df947" {
		t.Error("wrong address")
	}

	seed, err = NewSeed()
	if err != nil {
		t.Error(err)
	}

	if len(seed) != 64 {
		t.Error("expected size of 64")
	}

	seed, err = NewSeedFromMnemonic(mnemonic)
	if err != nil {
		t.Error(err)
	}

	if len(seed) != 64 {
		t.Error("expected size of 64")
	}

	mnemonic, err = NewMnemonic(128)
	if err != nil {
		t.Error(err)
	}

	words := strings.Split(mnemonic, " ")
	if len(words) != 12 {
		t.Error("expected 12 words")
	}
}

