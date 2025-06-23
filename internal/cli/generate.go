package cli

import (
	"fmt"

	"github.com/jasony/simple-eth-hd-wallet/internal/wallet"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new mnemonic phrase",
	Long: `Generate a new cryptographically secure mnemonic phrase that can be used 
to create a hierarchical deterministic wallet.

The mnemonic follows BIP-39 standard and can be used to deterministically 
generate private keys and addresses.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		bits, _ := cmd.Flags().GetInt("bits")

		if bits != 128 && bits != 160 && bits != 192 && bits != 224 && bits != 256 {
			return fmt.Errorf("invalid entropy bits: %d (must be 128, 160, 192, 224, or 256)", bits)
		}

		mnemonic, err := wallet.NewMnemonic(bits)
		if err != nil {
			return fmt.Errorf("failed to generate mnemonic: %w", err)
		}

		fmt.Printf("Generated mnemonic phrase:\n%s\n", mnemonic)
		fmt.Printf("\nEntropy: %d bits\n", bits)
		fmt.Printf("Words: %d\n", len(fmt.Fields(mnemonic)))

		fmt.Printf("\n⚠️  SECURITY WARNING:\n")
		fmt.Printf("Store this mnemonic phrase safely and securely.\n")
		fmt.Printf("Anyone with access to this phrase can control your wallet.\n")

		return nil
	},
}

func init() {
	generateCmd.Flags().IntP("bits", "b", 256, "Entropy bits (128, 160, 192, 224, or 256)")
	rootCmd.AddCommand(generateCmd)
}
