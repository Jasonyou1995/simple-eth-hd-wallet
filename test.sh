#!/bin/bash

# SKMS - Test Runner Script
# Runs comprehensive tests for the Simple Ethereum HD Wallet

set -e

echo "🧪 SKMS Test Suite"
echo "=================="
echo

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_step() {
    echo -e "${BLUE}$1${NC}"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

# Build the application first
print_step "Building SKMS..."
go build -o bin/skms ./cmd/skms
print_success "Build completed"
echo

# Run all tests
print_step "Running all tests..."
go test ./...
print_success "All tests passed"
echo

# Run tests with verbose output
print_step "Running wallet tests (verbose)..."
go test ./internal/wallet -v
print_success "Verbose tests completed"
echo

# Run tests with coverage
print_step "Running tests with coverage..."
go test ./internal/wallet -cover
echo

# Generate detailed coverage report
print_step "Generating detailed coverage report..."
go test ./internal/wallet -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
print_success "Coverage report generated: coverage.html"
echo

# Run benchmarks if any exist
print_step "Running benchmarks..."
go test ./internal/wallet -bench=. -benchmem || print_warning "No benchmarks found"
echo

# Code quality checks
print_step "Running go vet..."
go vet ./...
print_success "Go vet passed"
echo

print_step "Running go fmt check..."
if [ "$(gofmt -l . | wc -l)" -eq 0 ]; then
    print_success "Code formatting is correct"
else
    print_warning "Code formatting issues found. Run 'go fmt ./...' to fix."
    gofmt -l .
fi
echo

# Manual testing examples
print_step "Manual Testing Examples:"
echo "========================"
echo

echo "1. Test mnemonic generation:"
echo "   ./bin/skms generate"
echo "   ./bin/skms generate 256"
echo

echo "2. Test account derivation:"
echo '   ./bin/skms derive "abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about" 0'
echo

echo "3. Test error handling:"
echo "   ./bin/skms generate 100  # Invalid entropy"
echo '   ./bin/skms derive "invalid words" 0  # Invalid mnemonic'
echo

echo "4. Test batch generation:"
echo "   for i in {0..4}; do ./bin/skms derive "<mnemonic>" $i; done"
echo

# Final summary
print_success "Test suite completed successfully!"
echo
print_step "Summary:"
echo "• All unit tests passed"
echo "• Code coverage report generated"
echo "• Application built successfully"
echo "• Ready for manual testing"
echo

if command -v open >/dev/null 2>&1; then
    echo "Opening coverage report in browser..."
    open coverage.html
elif command -v xdg-open >/dev/null 2>&1; then
    echo "Opening coverage report in browser..."
    xdg-open coverage.html
else
    echo "Open coverage.html in your browser to view the detailed coverage report"
fi 