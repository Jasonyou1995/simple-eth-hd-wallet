// Package main provides the SKMS (Secure Key Management System) CLI application.
//
// SKMS is a production-ready hierarchical deterministic (HD) wallet tool
// that implements BIP-39 and BIP-44 standards for secure key management.
package main

import (
	"fmt"
	"os"
	"strconv"

	"simple-eth-hd-wallet/internal/wallet"
)

const (
	version = "1.0.0"
	appName = "SKMS - Secure Key Management System"
)

// printUsage displays the CLI usage information
func printUsage() {
	fmt.Printf(`%s v%s

Usage:
  skms <command> [arguments]

Commands:
  generate [entropy-bits]    Generate a new BIP-39 mnemonic phrase
                            entropy-bits: 128, 160, 192, 224, or 256 (default: 128)
  
  derive <mnemonic> <index>  Derive an Ethereum account from mnemonic
                            mnemonic: BIP-39 mnemonic phrase (quoted)
                            index: account index (0, 1, 2, ...)
  
  help                      Show this help message
  version                   Show version information

Examples:
  skms generate 128
  skms derive "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about" 0

Security Warning:
  This tool handles sensitive cryptographic material. Always:
  • Run on secure, trusted systems
  • Keep private keys and mnemonics secure
  • Verify the integrity of generated keys
  • Use hardware wallets for production funds

`, appName, version)
}

// printVersion displays version information
func printVersion() {
	fmt.Printf("%s v%s\n", appName, version)
	fmt.Println("Production-ready HD wallet CLI")
	fmt.Println("Implements BIP-39 and BIP-44 standards")
}

// generateMnemonic handles mnemonic generation
func generateMnemonic(args []string) error {
	entropyBits := 128 // Default entropy

	if len(args) > 0 {
		bits, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid entropy bits: %v", err)
		}
		if bits != 128 && bits != 160 && bits != 192 && bits != 224 && bits != 256 {
			return fmt.Errorf("entropy bits must be 128, 160, 192, 224, or 256")
		}
		entropyBits = bits
	}

	fmt.Printf("Generating new %d-bit mnemonic phrase...\n", entropyBits)

	mnemonic, err := wallet.GenerateMnemonic(entropyBits)
	if err != nil {
		return fmt.Errorf("failed to generate mnemonic: %v", err)
	}

	fmt.Printf("\n✅ Mnemonic generated successfully!\n\n")
	fmt.Printf("Mnemonic Phrase:\n%s\n\n", mnemonic)
	fmt.Printf("⚠️  SECURITY WARNING:\n")
	fmt.Printf("• Write down this mnemonic phrase and store it securely\n")
	fmt.Printf("• Anyone with this phrase can access your funds\n")
	fmt.Printf("• Never share it online or store it digitally\n")
	fmt.Printf("• This phrase cannot be recovered if lost\n\n")

	return nil
}

// deriveAccount handles account derivation
func deriveAccount(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("derive command requires mnemonic phrase and account index")
	}

	mnemonic := args[0]
	indexStr := args[1]

	// Parse account index
	index, err := strconv.ParseUint(indexStr, 10, 32)
	if err != nil {
		return fmt.Errorf("invalid account index: %v", err)
	}

	fmt.Printf("Deriving account at index %d...\n", index)

	// Create wallet from mnemonic
	config := wallet.DefaultConfig()
	w, err := wallet.NewFromMnemonic(mnemonic, config)
	if err != nil {
		return fmt.Errorf("failed to create wallet: %v", err)
	}
	defer w.Close()

	// Derive the account
	account, err := w.Derive(uint32(index))
	if err != nil {
		return fmt.Errorf("failed to derive account: %v", err)
	}

	// Display account information
	fmt.Printf("\n✅ Account derived successfully!\n\n")
	fmt.Printf("Account Index:    %d\n", account.Index)
	fmt.Printf("Derivation Path:  %s\n", account.Path)
	fmt.Printf("Ethereum Address: %s\n", account.Address.String())

	// Get private key (be careful with this in production!)
	privateKeyHex, err := w.GetPrivateKeyHex(account.Address)
	if err != nil {
		return fmt.Errorf("failed to get private key: %v", err)
	}
	fmt.Printf("Private Key:      0x%s\n", privateKeyHex)

	// Get public key
	publicKeyHex, err := w.GetPublicKeyHex(account.Address)
	if err != nil {
		return fmt.Errorf("failed to get public key: %v", err)
	}
	fmt.Printf("Public Key:       0x%s\n", publicKeyHex)

	fmt.Printf("\n⚠️  Warning: Keep your private key secure and never share it!\n")

	return nil
}

// main is the application entry point
func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]

	var err error

	switch command {
	case "generate":
		err = generateMnemonic(args)
	case "derive":
		err = deriveAccount(args)
	case "help", "--help", "-h":
		printUsage()
		return
	case "version", "--version", "-v":
		printVersion()
		return
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
