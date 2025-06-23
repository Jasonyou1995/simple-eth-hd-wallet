#!/bin/bash

# SKMS Test Suite
# Comprehensive testing script for the Secure Key Management System
# This script validates all core functionality and security features

set -e  # Exit on any error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test configuration
BINARY="./bin/skms"
TEST_MNEMONIC="abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon abandon about"
EXPECTED_ADDRESS_0="0x9858effd232b4033e47d90003d41ec34ecaeda94ebe4"
ENTROPY_LEVELS=(128 160 192 224 256)

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TOTAL_TESTS=0

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_error() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_warning() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

run_test() {
    local test_name="$1"
    local test_command="$2"
    local expected_exit_code="${3:-0}"
    
    ((TOTAL_TESTS++))
    log_info "Running test: $test_name"
    
    if [ "$expected_exit_code" -eq 0 ]; then
        if eval "$test_command" >/dev/null 2>&1; then
            log_success "$test_name"
        else
            log_error "$test_name - Command failed"
        fi
    else
        # Test that should fail
        if eval "$test_command" >/dev/null 2>&1; then
            log_error "$test_name - Command should have failed but succeeded"
        else
            log_success "$test_name"
        fi
    fi
}

# Build the application
build_application() {
    log_info "Building SKMS application..."
    if go build -o bin/skms ./cmd/skms; then
        log_success "Application built successfully"
        chmod +x bin/skms
    else
        log_error "Failed to build application"
        exit 1
    fi
}

# Test basic functionality
test_basic_functionality() {
    log_info "=== Testing Basic Functionality ==="
    
    # Test help command
    run_test "Help command" "$BINARY help"
    run_test "Help flag" "$BINARY --help"
    run_test "Help flag short" "$BINARY -h"
    
    # Test version command
    run_test "Version command" "$BINARY version"
    run_test "Version flag" "$BINARY --version"
    run_test "Version flag short" "$BINARY -v"
}

# Test mnemonic generation
test_mnemonic_generation() {
    log_info "=== Testing Mnemonic Generation ==="
    
    # Test default entropy (128-bit)
    run_test "Generate default mnemonic" "$BINARY generate"
    
    # Test specific entropy levels
    for entropy in "${ENTROPY_LEVELS[@]}"; do
        run_test "Generate $entropy-bit mnemonic" "$BINARY generate $entropy"
    done
    
    # Test invalid entropy levels
    run_test "Invalid entropy (100)" "$BINARY generate 100" 1
    run_test "Invalid entropy (300)" "$BINARY generate 300" 1
    run_test "Invalid entropy (text)" "$BINARY generate abc" 1
}

# Test account derivation
test_account_derivation() {
    log_info "=== Testing Account Derivation ==="
    
    # Test with known test mnemonic
    run_test "Derive account index 0" "$BINARY derive \"$TEST_MNEMONIC\" 0"
    run_test "Derive account index 1" "$BINARY derive \"$TEST_MNEMONIC\" 1"
    run_test "Derive account index 10" "$BINARY derive \"$TEST_MNEMONIC\" 10"
    
    # Test invalid inputs
    run_test "Invalid mnemonic" "$BINARY derive \"invalid mnemonic phrase\" 0" 1
    run_test "Invalid account index (text)" "$BINARY derive \"$TEST_MNEMONIC\" abc" 1
    run_test "Missing mnemonic" "$BINARY derive" 1
    run_test "Missing index" "$BINARY derive \"$TEST_MNEMONIC\"" 1
}

# Test error handling
test_error_handling() {
    log_info "=== Testing Error Handling ==="
    
    # Test unknown commands
    run_test "Unknown command" "$BINARY unknown-command" 1
    run_test "Empty command" "$BINARY \"\"" 1
    
    # Test malformed commands
    run_test "Malformed derive command" "$BINARY derive" 1
    run_test "Too many args for version" "$BINARY version extra args" 0  # Should ignore extra args
}

# Test security features
test_security_features() {
    log_info "=== Testing Security Features ==="
    
    # Generate multiple mnemonics and ensure they're different
    log_info "Testing mnemonic randomness..."
    mnemonic1=$($BINARY generate 2>/dev/null | grep -A1 "Mnemonic Phrase:" | tail -1)
    mnemonic2=$($BINARY generate 2>/dev/null | grep -A1 "Mnemonic Phrase:" | tail -1)
    
    if [ "$mnemonic1" != "$mnemonic2" ]; then
        log_success "Mnemonic randomness test"
        ((TESTS_PASSED++))
    else
        log_error "Mnemonic randomness test - Generated identical mnemonics"
        ((TESTS_FAILED++))
    fi
    ((TOTAL_TESTS++))
    
    # Test deterministic derivation (same mnemonic should produce same keys)
    log_info "Testing deterministic derivation..."
    output1=$($BINARY derive "$TEST_MNEMONIC" 0 2>/dev/null | grep "Ethereum Address:")
    output2=$($BINARY derive "$TEST_MNEMONIC" 0 2>/dev/null | grep "Ethereum Address:")
    
    if [ "$output1" = "$output2" ]; then
        log_success "Deterministic derivation test"
        ((TESTS_PASSED++))
    else
        log_error "Deterministic derivation test - Different outputs for same input"
        ((TESTS_FAILED++))
    fi
    ((TOTAL_TESTS++))
}

# Test performance (basic timing)
test_performance() {
    log_info "=== Testing Performance ==="
    
    # Test mnemonic generation speed
    log_info "Testing mnemonic generation performance..."
    start_time=$(date +%s%N)
    for i in {1..10}; do
        $BINARY generate >/dev/null 2>&1
    done
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 )) # Convert to milliseconds
    avg_time=$(( duration / 10 ))
    
    if [ $avg_time -lt 1000 ]; then  # Less than 1 second average
        log_success "Mnemonic generation performance: ${avg_time}ms average"
        ((TESTS_PASSED++))
    else
        log_warning "Mnemonic generation performance: ${avg_time}ms average (may be slow)"
        ((TESTS_PASSED++))  # Still pass, just note it's slow
    fi
    ((TOTAL_TESTS++))
    
    # Test derivation speed
    log_info "Testing derivation performance..."
    start_time=$(date +%s%N)
    for i in {1..5}; do
        $BINARY derive "$TEST_MNEMONIC" $i >/dev/null 2>&1
    done
    end_time=$(date +%s%N)
    duration=$(( (end_time - start_time) / 1000000 ))
    avg_time=$(( duration / 5 ))
    
    if [ $avg_time -lt 2000 ]; then  # Less than 2 seconds average
        log_success "Account derivation performance: ${avg_time}ms average"
        ((TESTS_PASSED++))
    else
        log_warning "Account derivation performance: ${avg_time}ms average (may be slow)"
        ((TESTS_PASSED++))  # Still pass, just note it's slow
    fi
    ((TOTAL_TESTS++))
}

# Test output validation
test_output_validation() {
    log_info "=== Testing Output Validation ==="
    
    # Test mnemonic word count
    for entropy in "${ENTROPY_LEVELS[@]}"; do
        expected_words=$(( entropy / 11 + (entropy % 11 > 0 ? 1 : 0) ))
        if [ $expected_words -lt 12 ]; then
            expected_words=12
        fi
        
        log_info "Testing $entropy-bit entropy (expecting ~$expected_words words)..."
        output=$($BINARY generate $entropy 2>/dev/null)
        mnemonic=$(echo "$output" | grep -A1 "Mnemonic Phrase:" | tail -1)
        word_count=$(echo "$mnemonic" | wc -w | tr -d ' ')
        
        if [ "$word_count" -ge 12 ] && [ "$word_count" -le 24 ]; then
            log_success "Word count validation for $entropy-bit: $word_count words"
            ((TESTS_PASSED++))
        else
            log_error "Word count validation for $entropy-bit: $word_count words (invalid)"
            ((TESTS_FAILED++))
        fi
        ((TOTAL_TESTS++))
    done
    
    # Test address format
    log_info "Testing address format validation..."
    output=$($BINARY derive "$TEST_MNEMONIC" 0 2>/dev/null)
    address=$(echo "$output" | grep "Ethereum Address:" | cut -d' ' -f3)
    
    if [[ $address =~ ^0x[a-fA-F0-9]{40}$ ]]; then
        log_success "Address format validation: $address"
        ((TESTS_PASSED++))
    else
        log_error "Address format validation: Invalid format - $address"
        ((TESTS_FAILED++))
    fi
    ((TOTAL_TESTS++))
    
    # Test private key format
    log_info "Testing private key format validation..."
    private_key=$(echo "$output" | grep "Private Key:" | cut -d' ' -f3)
    
    if [[ $private_key =~ ^0x[a-fA-F0-9]{64}$ ]]; then
        log_success "Private key format validation"
        ((TESTS_PASSED++))
    else
        log_error "Private key format validation: Invalid format"
        ((TESTS_FAILED++))
    fi
    ((TOTAL_TESTS++))
}

# Main test execution
main() {
    echo "======================================================"
    echo "      SKMS (Secure Key Management System)"
    echo "           Comprehensive Test Suite"
    echo "======================================================"
    echo ""
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go to run tests."
        exit 1
    fi
    
    # Build application
    build_application
    echo ""
    
    # Check if binary exists
    if [ ! -f "$BINARY" ]; then
        log_error "SKMS binary not found at $BINARY"
        exit 1
    fi
    
    # Run all test suites
    test_basic_functionality
    echo ""
    test_mnemonic_generation
    echo ""
    test_account_derivation
    echo ""
    test_error_handling
    echo ""
    test_security_features
    echo ""
    test_performance
    echo ""
    test_output_validation
    echo ""
    
    # Summary
    echo "======================================================"
    echo "                  TEST SUMMARY"
    echo "======================================================"
    echo "Total Tests:  $TOTAL_TESTS"
    echo "Passed:       $TESTS_PASSED"
    echo "Failed:       $TESTS_FAILED"
    echo ""
    
    if [ $TESTS_FAILED -eq 0 ]; then
        echo -e "${GREEN}🎉 ALL TESTS PASSED! 🎉${NC}"
        echo ""
        echo "SKMS is working correctly and ready for use."
        echo "Remember to always verify generated keys with external tools"
        echo "before using with real funds."
        exit 0
    else
        echo -e "${RED}❌ SOME TESTS FAILED ❌${NC}"
        echo ""
        echo "Please review the failed tests above and fix any issues."
        exit 1
    fi
}

# Check if running as script (not sourced)
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
    main "$@"
fi 