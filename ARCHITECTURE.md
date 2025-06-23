# SKMS Technical Architecture

## Overview

SKMS (Secure Key Management System) is a CLI application implementing modern cryptographic best practices for Ethereum HD wallet generation. This document explains the technical architecture, design decisions, and implementation patterns.

## 🏗️ Architecture Principles

### 1. Minimalist Security Design

- **Zero External Dependencies**: Uses only Go standard library to minimize attack surface
- **Single Responsibility**: Each component has a clearly defined purpose
- **Fail-Safe Defaults**: Secure configurations by default
- **Defense in Depth**: Multiple security layers protect sensitive operations

### 2. Modern Go Patterns

#### Clean Architecture

```
cmd/skms/              # Application entry point
├── main.go           # CLI argument parsing and command routing

internal/wallet/       # Core business logic
├── simple_wallet.go  # Wallet implementation with security features

bin/                  # Compiled binaries
├── skms              # Production executable
```

#### Dependency Injection & Interfaces

```go
// Wallet interface for testability and extensibility
type Wallet interface {
    GenerateMnemonic(entropyBits int) (string, error)
    DeriveAccount(mnemonic string, accountIndex uint32) (*Account, error)
}

// Account represents a derived Ethereum account
type Account struct {
    Address    string
    PrivateKey string
    PublicKey  string
    Index      uint32
}
```

### 3. Security-First Implementation

#### Memory Management

```go
type SimpleWallet struct {
    mu           sync.RWMutex
    entropy      []byte          // Secured with finalizer
    lastMnemonic []byte          // Secured with finalizer
    accounts     map[uint32]*secureAccount
}

// Automatic cleanup using Go finalizers
func NewSimpleWallet() *SimpleWallet {
    w := &SimpleWallet{
        accounts: make(map[uint32]*secureAccount),
    }
    runtime.SetFinalizer(w, (*SimpleWallet).cleanup)
    return w
}
```

#### Thread Safety

```go
// Fine-grained locking for concurrent operations
func (w *SimpleWallet) DeriveAccount(mnemonic string, accountIndex uint32) (*Account, error) {
    w.mu.Lock()
    defer w.mu.Unlock()

    // Critical section protected
    return w.deriveAccountUnsafe(mnemonic, accountIndex)
}
```

## 🔐 Cryptographic Implementation

### 1. BIP Standards Compliance

#### BIP-39 (Mnemonic Generation)

```go
// Entropy to mnemonic conversion following BIP-39
func entropyToMnemonic(entropy []byte) (string, error) {
    // 1. Create checksum
    checksum := sha256.Sum256(entropy)
    checksumBits := len(entropy) * 8 / 32

    // 2. Append checksum to entropy
    entropyWithChecksum := new(big.Int).SetBytes(entropy)
    checksumInt := new(big.Int).SetBytes(checksum[:])
    // ... bit manipulation for BIP-39 compliance
}
```

#### BIP-44 (HD Derivation)

```go
// Standard Ethereum derivation path: m/44'/60'/0'/0/index
const (
    PURPOSE       = 44   // BIP-44
    COIN_TYPE     = 60   // Ethereum
    ACCOUNT       = 0    // First account
    CHANGE        = 0    // External chain
)
```

### 2. Modern Cryptographic Practices

#### Secure Random Generation

```go
// Use crypto/rand for cryptographically secure randomness
func generateEntropy(bits int) ([]byte, error) {
    bytes := make([]byte, bits/8)
    if _, err := rand.Read(bytes); err != nil {
        return nil, fmt.Errorf("failed to generate entropy: %w", err)
    }
    return bytes, nil
}
```

#### Key Derivation

```go
// HMAC-SHA512 based key derivation
func derivePrivateKey(seed []byte, path string) (*ecdsa.PrivateKey, error) {
    // PBKDF2 with HMAC-SHA512
    key := pbkdf2.Key(seed, []byte("mnemonic"), 2048, 64, sha512.New)

    // Derive using path
    privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
    // ... proper BIP-44 derivation implementation
}
```

## 🛠️ Modern Development Practices

### 1. Error Handling Patterns

#### Structured Error Types

```go
// Custom error types for better error handling
type ValidationError struct {
    Field   string
    Value   interface{}
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error in %s: %s (got: %v)",
        e.Field, e.Message, e.Value)
}
```

#### Error Wrapping

```go
// Go 1.13+ error wrapping for context
func (w *SimpleWallet) validateMnemonic(mnemonic string) error {
    if len(words) < 12 {
        return fmt.Errorf("invalid mnemonic length: %w",
            &ValidationError{
                Field:   "mnemonic",
                Value:   len(words),
                Message: "must be at least 12 words",
            })
    }
}
```

### 2. Resource Management

#### Context-Aware Operations

```go
// Context support for cancellable operations
func (w *SimpleWallet) DeriveAccountWithContext(
    ctx context.Context,
    mnemonic string,
    accountIndex uint32,
) (*Account, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        return w.DeriveAccount(mnemonic, accountIndex)
    }
}
```

#### Graceful Cleanup

```go
// Explicit cleanup methods
func (w *SimpleWallet) cleanup() {
    w.mu.Lock()
    defer w.mu.Unlock()

    // Zero sensitive memory
    if w.entropy != nil {
        for i := range w.entropy {
            w.entropy[i] = 0
        }
    }

    // Clear accounts
    for k, v := range w.accounts {
        v.zero()
        delete(w.accounts, k)
    }
}
```

### 3. Testing Strategies

#### Table-Driven Tests

```go
func TestEntropyLevels(t *testing.T) {
    tests := []struct {
        name        string
        entropyBits int
        expectError bool
        expectWords int
    }{
        {"128-bit entropy", 128, false, 12},
        {"160-bit entropy", 160, false, 15},
        {"192-bit entropy", 192, false, 18},
        {"224-bit entropy", 224, false, 21},
        {"256-bit entropy", 256, false, 24},
        {"Invalid entropy", 100, true, 0},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

## 🚀 Performance Optimizations

### 1. Memory Efficiency

#### Object Pooling for Frequent Operations

```go
var (
    // Pool for byte slices to reduce allocations
    bytePool = sync.Pool{
        New: func() interface{} {
            return make([]byte, 32)
        },
    }
)

func (w *SimpleWallet) deriveKey() {
    buf := bytePool.Get().([]byte)
    defer bytePool.Put(buf)

    // Use buf for temporary calculations
}
```

#### Lazy Loading

```go
// Initialize BIP-39 word list only when needed
var (
    bip39Words []string
    bip39Once  sync.Once
)

func getBIP39Words() []string {
    bip39Once.Do(func() {
        bip39Words = strings.Split(bip39WordList, "\n")
    })
    return bip39Words
}
```

### 2. Computational Efficiency

#### Precomputed Values

```go
// Precompute common derivation paths
var commonPaths = map[uint32]string{
    0: "m/44'/60'/0'/0/0",
    1: "m/44'/60'/0'/0/1",
    // ... up to reasonable limit
}
```

#### Batch Operations

```go
// Derive multiple accounts efficiently
func (w *SimpleWallet) DeriveAccounts(
    mnemonic string,
    indices []uint32,
) ([]*Account, error) {
    // Validate once, derive many
    if err := w.validateMnemonic(mnemonic); err != nil {
        return nil, err
    }

    accounts := make([]*Account, len(indices))
    for i, idx := range indices {
        accounts[i], _ = w.deriveAccountUnsafe(mnemonic, idx)
    }
    return accounts, nil
}
```

## 🔍 Quality Assurance

### 1. Static Analysis Integration

#### Go Vet & Linting

```bash
# Comprehensive static analysis
go vet ./...
golangci-lint run
gosec ./...
staticcheck ./...
```

#### Security Scanning

```bash
# Dependency vulnerability scanning
go list -json -deps ./... | nancy sleuth
govulncheck ./...
```

### 2. Comprehensive Testing

#### Test Categories

- **Unit Tests**: Individual function testing
- **Integration Tests**: Component interaction testing
- **Security Tests**: Cryptographic validation
- **Performance Tests**: Benchmarking critical paths
- **Fuzz Tests**: Random input validation

#### Test Coverage

```bash
# Generate coverage reports
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## 🔄 DevOps & Deployment

### 1. Build Process

#### Reproducible Builds

```bash
# Deterministic compilation
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -ldflags="-s -w -X main.version=${VERSION}" \
    -o bin/skms-linux-amd64 ./cmd/skms
```

#### Multi-Platform Support

```bash
# Cross-compilation for multiple platforms
platforms=("linux/amd64" "darwin/amd64" "windows/amd64")
for platform in "${platforms[@]}"; do
    # Build for each platform
done
```

### 2. Security Hardening

#### Binary Hardening

```bash
# Strip symbols and debug info
go build -ldflags="-s -w" ./cmd/skms

# Enable stack protection (when available)
CGO_ENABLED=1 go build -ldflags="-linkmode external -extldflags '-static'" ./cmd/skms
```

## 📊 Observability

### 1. Logging Strategy

#### Structured Logging

```go
type Logger struct {
    level LogLevel
    out   io.Writer
}

func (l *Logger) SecurityEvent(event string, details map[string]interface{}) {
    entry := map[string]interface{}{
        "timestamp": time.Now().UTC(),
        "level":     "SECURITY",
        "event":     event,
        "details":   details,
    }
    // Log without sensitive data
}
```

### 2. Metrics Collection

#### Performance Metrics

```go
var (
    mnemonicGenerationDuration = prometheus.NewHistogram(
        prometheus.HistogramOpts{
            Name: "mnemonic_generation_duration_seconds",
            Help: "Time taken to generate mnemonic phrases",
        },
    )
)
```

## 🔮 Future Architecture Considerations

### 1. Extensibility Patterns

#### Plugin Architecture

```go
// Plugin interface for additional functionality
type KeyDerivationPlugin interface {
    Name() string
    DerivePath() string
    DeriveKey(seed []byte, index uint32) (*ecdsa.PrivateKey, error)
}
```

### 2. Scalability Preparations

#### Horizontal Scaling

- Stateless design enables multiple instances
- Shared-nothing architecture
- Event-driven updates for coordination

#### Performance Scaling

- Worker pool patterns for batch operations
- Caching strategies for computed values
- Async operations for non-critical paths

## 📚 Technology Stack

### Core Technologies

- **Language**: Go 1.21+ (latest stable)
- **Cryptography**: Go standard library (`crypto/`)
- **Concurrency**: Go routines and channels
- **CLI**: Standard library flag parsing
- **Testing**: Go testing framework with custom extensions

### Development Tools

- **Linting**: golangci-lint, gosec, staticcheck
- **Testing**: go test, testify (if needed)
- **Profiling**: go tool pprof
- **Documentation**: Go doc, markdown

### Build & Deployment

- **Build**: Go compiler with cross-compilation
- **CI/CD**: GitHub Actions (when applicable)
- **Security**: Vulnerability scanning, SAST tools
- **Distribution**: Single binary distribution

---

This architecture demonstrates modern Go development practices while maintaining the highest security standards for cryptographic key management. The design prioritizes security, performance, and maintainability in equal measure.
