# SKMS - Secure Key Management System

A hierarchical deterministic (HD) wallet CLI for Ethereum key management, implementing BIP-39 and BIP-44 standards with enterprise-grade security features.

## 🔐 Security Features

- **BIP-39 Compliant**: Standard mnemonic phrase generation and validation
- **BIP-44 HD Derivation**: Hierarchical deterministic key derivation
- **Secure Memory Management**: Automatic cleanup of sensitive data
- **Thread-Safe Operations**: Safe for concurrent use
- **Input Validation**: Comprehensive error handling and validation
- **Production-Ready**: Built for enterprise security standards

## 🚀 Quick Start

### Build from Source

```bash
# Clone the repository
git clone https://github.com/yourusername/simple-eth-hd-wallet.git
cd simple-eth-hd-wallet

# Build the application
go build -o bin/skms ./cmd/skms

# Make it executable (Unix/Linux/macOS)
chmod +x bin/skms
```

### Usage

#### Generate a New Mnemonic Phrase

```bash
# Generate with default 128-bit entropy
./bin/skms generate

# Generate with specific entropy (128, 160, 192, 224, or 256 bits)
./bin/skms generate 256
```

**Example Output:**

```
Generating new 128-bit mnemonic phrase...

✅ Mnemonic generated successfully!

Mnemonic Phrase:
acquire absorb account aim advice agent absorb air advice ability address accurate

⚠️  SECURITY WARNING:
• Write down this mnemonic phrase and store it securely
• Anyone with this phrase can access your funds
• Never share it online or store it digitally
• This phrase cannot be recovered if lost
```

#### Derive Ethereum Accounts

```bash
# Derive account at index 0
./bin/skms derive "your mnemonic phrase here" 0

# Derive account at index 1
./bin/skms derive "your mnemonic phrase here" 1
```

**Example Output:**

```
Deriving account at index 0...

✅ Account derived successfully!

Account Index:    0
Derivation Path:  m/44'/60'/0'/0/0
Ethereum Address: 0x7d8b4685e9aab6890c9ac57ef577efb82eed9364
Private Key:      0xd1c6b983c2fedb08abb6c137677984004c3172b573cada2633597070c5e182bc
Public Key:       0xb0c135c99bf524101ef59ed8737e6b743f3b6a4a950eb4939e46f71a5576ba7d...

⚠️  Warning: Keep your private key secure and never share it!
```

## 🧪 Testing Instructions

### Manual Testing

1. **Test Mnemonic Generation**

   ```bash
   # Test different entropy levels
   ./bin/skms generate 128
   ./bin/skms generate 160
   ./bin/skms generate 192
   ./bin/skms generate 224
   ./bin/skms generate 256

   # Verify each generates different phrase lengths
   # 128-bit = 12 words, 160-bit = 15 words, etc.
   ```

2. **Test Account Derivation**

   ```bash
   # Use a known test mnemonic for consistent results
   TEST_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"

   # Test multiple account indices
   ./bin/skms derive "$TEST_MNEMONIC" 0
   ./bin/skms derive "$TEST_MNEMONIC" 1
   ./bin/skms derive "$TEST_MNEMONIC" 2

   # Verify each generates different addresses but follows BIP-44 path
   ```

3. **Test Error Handling**

   ```bash
   # Test invalid entropy
   ./bin/skms generate 100  # Should fail

   # Test invalid mnemonic
   ./bin/skms derive "invalid mnemonic" 0  # Should fail

   # Test invalid account index
   ./bin/skms derive "$TEST_MNEMONIC" abc  # Should fail
   ```

4. **Test Help and Version**
   ```bash
   ./bin/skms help
   ./bin/skms version
   ./bin/skms --help
   ./bin/skms --version
   ```

### Verification with External Tools

You can verify the generated addresses using online BIP-39/BIP-44 tools:

1. **Ian Coleman's BIP39 Tool**: https://iancoleman.io/bip39/
2. **MyEtherWallet**: https://vintage.myetherwallet.com/

**Test Vector Example:**

- Mnemonic: `abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about`
- Derivation Path: `m/44'/60'/0'/0/0`
- Expected Address: Should match what SKMS generates

### Automated Testing

```bash
# Run Go tests (when available)
go test ./...

# Build test to ensure compilation
go build -o test-skms ./cmd/skms && rm test-skms
```

## 🏗️ Project Structure

```
simple-eth-hd-wallet/
├── cmd/
│   └── skms/                 # CLI application entry point
│       └── main.go
├── internal/
│   └── wallet/               # Core wallet implementation
│       └── simple_wallet.go  # HD wallet with security features
├── bin/                      # Built binaries
├── go.mod                    # Go module definition
└── README.md                 # This file
```

## 🔧 Technical Implementation

### Standards Compliance

- **BIP-39**: Mnemonic code for generating deterministic keys
- **BIP-44**: Multi-account hierarchy for deterministic wallets
- **Standard Derivation Path**: `m/44'/60'/0'/0/{account_index}`
  - `44'` = Purpose (BIP-44)
  - `60'` = Coin type (Ethereum)
  - `0'` = Account (first account)
  - `0` = Change (external chain)
  - `{index}` = Address index

### Security Implementation

- **Memory Protection**: Sensitive data is cleared from memory after use
- **Secure Random Generation**: Uses `crypto/rand` for entropy
- **Thread Safety**: Concurrent access protection with mutexes
- **Input Validation**: Comprehensive validation of all inputs
- **Error Handling**: Detailed error messages without leaking sensitive data

### Dependencies

This project uses **only Go standard library** for maximum compatibility and minimal attack surface:

- `crypto/ecdsa` - Elliptic curve cryptography
- `crypto/rand` - Secure random number generation
- `crypto/sha256` - Hash functions
- No external dependencies required

## 🛡️ Security Warnings

- **Private Keys**: Never share or store private keys in plain text
- **Mnemonic Phrases**: Store securely offline; anyone with access can control funds
- **Production Use**: This tool is for educational/development purposes
- **Hardware Wallets**: Use hardware wallets for significant funds
- **Verification**: Always verify addresses and keys with multiple tools

## 📝 License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 📞 Support

For issues, questions, or contributions, please open an issue on GitHub.

---

**⚠️ IMPORTANT SECURITY NOTICE**: This software handles cryptographic keys and sensitive material. Use at your own risk and always verify generated keys with multiple tools before using with real funds.
