package cli

import (
	"fmt"

	"github.com/jasony/simple-eth-hd-wallet/internal/wallet"
	"github.com/spf13/cobra"
)

var deriveCmd = &cobra.Command{
	Use:   "derive",
	Short: "Derive addresses from mnemonic",
	Long: `Derive Ethereum addresses from a mnemonic phrase using BIP-44 derivation paths.
	
Default derivation path follows BIP-44 standard: m/44'/60'/0'/0/0
where 44' is the purpose, 60' is Ethereum's coin type, and the last numbers
represent account, change, and address index.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		mnemonic, _ := cmd.Flags().GetString("mnemonic")
		derivationPath, _ := cmd.Flags().GetString("path")
		count, _ := cmd.Flags().GetInt("count")
		showPrivate, _ := cmd.Flags().GetBool("private")

		if mnemonic == "" {
			return fmt.Errorf("mnemonic phrase is required")
		}

		w, err := wallet.NewFromMnemonic(mnemonic)
		if err != nil {
			return fmt.Errorf("failed to create wallet from mnemonic: %w", err)
		}

		fmt.Printf("Derivation Path: %s\n", derivationPath)
		fmt.Printf("Deriving %d address(es):\n\n", count)

		for i := 0; i < count; i++ {
			pathWithIndex := fmt.Sprintf("%s/%d", derivationPath, i)

			path, err := wallet.ParseDerivationPath(pathWithIndex)
			if err != nil {
				return fmt.Errorf("failed to parse derivation path %s: %w", pathWithIndex, err)
			}

			account, err := w.Derive(path, false)
			if err != nil {
				return fmt.Errorf("failed to derive account for path %s: %w", pathWithIndex, err)
			}

			fmt.Printf("Index %d:\n", i)
			fmt.Printf("  Path:    %s\n", pathWithIndex)
			fmt.Printf("  Address: %s\n", account.Address.Hex())

			if showPrivate {
				privateKeyHex, err := w.PrivateKeyHex(account)
				if err != nil {
					return fmt.Errorf("failed to get private key: %w", err)
				}
				fmt.Printf("  Private: %s\n", privateKeyHex)
			}

			publicKeyHex, err := w.PublicKeyHex(account)
			if err != nil {
				return fmt.Errorf("failed to get public key: %w", err)
			}
			fmt.Printf("  Public:  %s\n", publicKeyHex)
			fmt.Println()
		}

		if showPrivate {
			fmt.Printf("⚠️  WARNING: Private keys are shown above.\n")
			fmt.Printf("Keep them secure and never share them.\n")
		}

		return nil
	},
}

func init() {
	deriveCmd.Flags().StringP("mnemonic", "m", "", "Mnemonic phrase (required)")
	deriveCmd.Flags().StringP("path", "p", "m/44'/60'/0'/0", "Base derivation path (default: m/44'/60'/0'/0)")
	deriveCmd.Flags().IntP("count", "c", 1, "Number of addresses to derive")
	deriveCmd.Flags().Bool("private", false, "Show private keys (use with caution)")

	deriveCmd.MarkFlagRequired("mnemonic")
	rootCmd.AddCommand(deriveCmd)
}
