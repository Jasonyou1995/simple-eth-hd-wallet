# SKMS API Contracts and Interface Definitions

## Overview

This document defines the API contracts and interfaces for the Secure Key Management System (SKMS). All APIs are designed to be RESTful, following OpenAPI 3.0 specifications with strong typing and comprehensive error handling.

## API Design Principles

### RESTful Design

- Resource-based URLs
- HTTP verbs for operations (GET, POST, PUT, DELETE)
- Stateless requests
- Consistent response formats

### Security Requirements

- All endpoints require authentication
- Role-based access control (RBAC)
- Rate limiting on all endpoints
- Input validation and sanitization
- Audit logging for all operations

### Response Format Standards

```json
{
  "success": true|false,
  "data": { /* response payload */ },
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable message",
    "details": { /* additional error context */ }
  },
  "metadata": {
    "timestamp": "2024-06-16T12:00:00Z",
    "requestId": "uuid",
    "version": "v1"
  }
}
```

## Core API Endpoints

### 1. Authentication API (`/api/v1/auth`)

#### POST /api/v1/auth/login

**Purpose**: Authenticate user and obtain JWT token

**Request**:

```json
{
  "username": "string",
  "password": "string",
  "mfa_token": "string" // Optional for MFA
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "access_token": "jwt_token",
    "refresh_token": "refresh_token",
    "expires_in": 3600,
    "token_type": "Bearer",
    "user": {
      "id": "uuid",
      "username": "string",
      "roles": ["admin", "operator"],
      "permissions": ["key:create", "key:read"]
    }
  }
}
```

#### POST /api/v1/auth/refresh

**Purpose**: Refresh JWT token using refresh token

#### POST /api/v1/auth/logout

**Purpose**: Invalidate current session

### 2. Key Management API (`/api/v1/keys`)

#### POST /api/v1/keys

**Purpose**: Generate a new cryptographic key

**Request**:

```json
{
  "key_type": "rsa|ecdsa|ed25519",
  "key_size": 2048|256|448,
  "usage": ["signing", "encryption"],
  "metadata": {
    "name": "string",
    "description": "string",
    "tags": ["tag1", "tag2"]
  },
  "expiration": "2025-06-16T12:00:00Z", // Optional
  "hsm_required": false
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "key_id": "uuid",
    "public_key": "base64_encoded_public_key",
    "key_type": "rsa",
    "key_size": 2048,
    "fingerprint": "sha256_hash",
    "created_at": "2024-06-16T12:00:00Z",
    "expires_at": "2025-06-16T12:00:00Z",
    "status": "active",
    "metadata": {
      "name": "string",
      "description": "string",
      "tags": ["tag1", "tag2"]
    }
  }
}
```

#### GET /api/v1/keys

**Purpose**: List keys with filtering and pagination

**Query Parameters**:

- `page`: int (default: 1)
- `limit`: int (default: 100, max: 1000)
- `status`: active|expired|revoked
- `key_type`: rsa|ecdsa|ed25519
- `tag`: string (can be repeated)
- `created_after`: ISO8601 timestamp
- `created_before`: ISO8601 timestamp

#### GET /api/v1/keys/{key_id}

**Purpose**: Get specific key details

#### DELETE /api/v1/keys/{key_id}

**Purpose**: Revoke/delete a key

#### PUT /api/v1/keys/{key_id}/rotate

**Purpose**: Rotate an existing key

### 3. HD Wallet API (`/api/v1/wallets`)

#### POST /api/v1/wallets

**Purpose**: Create a new HD wallet

**Request**:

```json
{
  "name": "string",
  "blockchain": "ethereum|bitcoin|solana",
  "derivation_path": "m/44'/60'/0'/0", // BIP44 path
  "mnemonic_strength": 128|256, // bits
  "passphrase": "string", // Optional BIP39 passphrase
  "metadata": {
    "description": "string",
    "tags": ["tag1", "tag2"]
  }
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "wallet_id": "uuid",
    "name": "string",
    "blockchain": "ethereum",
    "master_public_key": "xpub...",
    "derivation_path": "m/44'/60'/0'/0",
    "created_at": "2024-06-16T12:00:00Z",
    "address_count": 0,
    "mnemonic": "word1 word2 ... word12", // Only returned on creation
    "metadata": {
      "description": "string",
      "tags": ["tag1", "tag2"]
    }
  }
}
```

#### GET /api/v1/wallets

**Purpose**: List HD wallets

#### GET /api/v1/wallets/{wallet_id}

**Purpose**: Get wallet details

#### POST /api/v1/wallets/{wallet_id}/addresses

**Purpose**: Generate new address from HD wallet

**Request**:

```json
{
  "count": 1,
  "start_index": 0, // Optional, continues from last index
  "change": false // BIP44 change address
}
```

#### GET /api/v1/wallets/{wallet_id}/addresses

**Purpose**: List addresses for wallet

### 4. Digital Signatures API (`/api/v1/signatures`)

#### POST /api/v1/signatures/sign

**Purpose**: Create digital signature

**Request**:

```json
{
  "key_id": "uuid",
  "data": "base64_encoded_data",
  "format": "raw|asn1|jose",
  "hash_algorithm": "sha256|sha512",
  "metadata": {
    "purpose": "string",
    "description": "string"
  }
}
```

**Response**:

```json
{
  "success": true,
  "data": {
    "signature": "base64_encoded_signature",
    "signature_id": "uuid",
    "algorithm": "ECDSA-SHA256",
    "format": "asn1",
    "created_at": "2024-06-16T12:00:00Z",
    "key_id": "uuid",
    "data_hash": "sha256_hash"
  }
}
```

#### POST /api/v1/signatures/verify

**Purpose**: Verify digital signature

**Request**:

```json
{
  "signature": "base64_encoded_signature",
  "data": "base64_encoded_data",
  "public_key": "base64_encoded_public_key", // Optional if key_id provided
  "key_id": "uuid", // Optional if public_key provided
  "format": "raw|asn1|jose"
}
```

#### GET /api/v1/signatures

**Purpose**: List signatures with filtering

### 5. Multi-Signature API (`/api/v1/multisig`)

#### POST /api/v1/multisig/schemes

**Purpose**: Create multi-signature scheme

**Request**:

```json
{
  "name": "string",
  "threshold": 2,
  "participants": [
    {
      "participant_id": "uuid",
      "public_key": "base64_encoded_key",
      "weight": 1
    }
  ],
  "metadata": {
    "description": "string",
    "tags": ["tag1", "tag2"]
  }
}
```

#### POST /api/v1/multisig/schemes/{scheme_id}/sign

**Purpose**: Participate in multi-signature signing

### 6. Zero-Knowledge Proofs API (`/api/v1/zkp`)

#### POST /api/v1/zkp/commitments

**Purpose**: Create cryptographic commitment

#### POST /api/v1/zkp/proofs

**Purpose**: Generate zero-knowledge proof

#### POST /api/v1/zkp/verify

**Purpose**: Verify zero-knowledge proof

### 7. Audit and Monitoring API (`/api/v1/audit`)

#### GET /api/v1/audit/logs

**Purpose**: Retrieve audit logs

**Query Parameters**:

- `start_time`: ISO8601 timestamp
- `end_time`: ISO8601 timestamp
- `event_type`: authentication|key_operation|api_access
- `user_id`: uuid
- `resource_id`: uuid
- `page`: int
- `limit`: int

**Response**:

```json
{
  "success": true,
  "data": {
    "logs": [
      {
        "event_id": "uuid",
        "timestamp": "2024-06-16T12:00:00Z",
        "event_type": "key_generation",
        "user_id": "uuid",
        "resource_id": "uuid",
        "action": "create",
        "result": "success|failure",
        "ip_address": "192.168.1.1",
        "user_agent": "string",
        "details": {
          "key_type": "rsa",
          "key_size": 2048
        }
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 100,
      "total": 1500,
      "total_pages": 15
    }
  }
}
```

#### GET /api/v1/audit/reports

**Purpose**: Generate compliance reports

### 8. System Health API (`/api/v1/health`)

#### GET /api/v1/health

**Purpose**: System health check

**Response**:

```json
{
  "success": true,
  "data": {
    "status": "healthy|degraded|unhealthy",
    "timestamp": "2024-06-16T12:00:00Z",
    "version": "1.0.0",
    "components": {
      "database": {
        "status": "healthy",
        "response_time_ms": 5
      },
      "hsm": {
        "status": "healthy",
        "response_time_ms": 10
      },
      "redis": {
        "status": "healthy",
        "response_time_ms": 2
      }
    },
    "metrics": {
      "uptime_seconds": 86400,
      "total_requests": 1000000,
      "error_rate": 0.01
    }
  }
}
```

#### GET /api/v1/health/ready

**Purpose**: Readiness probe for Kubernetes

#### GET /api/v1/health/live

**Purpose**: Liveness probe for Kubernetes

## gRPC Service Definitions

### Key Management Service

```protobuf
syntax = "proto3";

package skms.v1;

service KeyManagementService {
  rpc GenerateKey(GenerateKeyRequest) returns (GenerateKeyResponse);
  rpc GetKey(GetKeyRequest) returns (GetKeyResponse);
  rpc ListKeys(ListKeysRequest) returns (ListKeysResponse);
  rpc RevokeKey(RevokeKeyRequest) returns (RevokeKeyResponse);
  rpc RotateKey(RotateKeyRequest) returns (RotateKeyResponse);
}

message GenerateKeyRequest {
  string key_type = 1; // rsa, ecdsa, ed25519
  int32 key_size = 2;
  repeated string usage = 3; // signing, encryption
  KeyMetadata metadata = 4;
  google.protobuf.Timestamp expiration = 5;
  bool hsm_required = 6;
}

message GenerateKeyResponse {
  string key_id = 1;
  bytes public_key = 2;
  string key_type = 3;
  int32 key_size = 4;
  string fingerprint = 5;
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp expires_at = 7;
  KeyStatus status = 8;
  KeyMetadata metadata = 9;
}

message KeyMetadata {
  string name = 1;
  string description = 2;
  repeated string tags = 3;
}

enum KeyStatus {
  KEY_STATUS_UNSPECIFIED = 0;
  KEY_STATUS_ACTIVE = 1;
  KEY_STATUS_EXPIRED = 2;
  KEY_STATUS_REVOKED = 3;
}
```

### HD Wallet Service

```protobuf
service HDWalletService {
  rpc CreateWallet(CreateWalletRequest) returns (CreateWalletResponse);
  rpc GetWallet(GetWalletRequest) returns (GetWalletResponse);
  rpc ListWallets(ListWalletsRequest) returns (ListWalletsResponse);
  rpc GenerateAddress(GenerateAddressRequest) returns (GenerateAddressResponse);
  rpc ListAddresses(ListAddressesRequest) returns (ListAddressesResponse);
}
```

### Digital Signature Service

```protobuf
service DigitalSignatureService {
  rpc Sign(SignRequest) returns (SignResponse);
  rpc Verify(VerifyRequest) returns (VerifyResponse);
  rpc BatchSign(BatchSignRequest) returns (BatchSignResponse);
  rpc BatchVerify(BatchVerifyRequest) returns (BatchVerifyResponse);
}
```

## Error Codes and Handling

### Standard Error Codes

```
Authentication Errors (1000-1999)
- 1001: INVALID_CREDENTIALS
- 1002: TOKEN_EXPIRED
- 1003: TOKEN_INVALID
- 1004: MFA_REQUIRED
- 1005: ACCOUNT_LOCKED

Authorization Errors (2000-2999)
- 2001: INSUFFICIENT_PERMISSIONS
- 2002: RESOURCE_FORBIDDEN
- 2003: RATE_LIMIT_EXCEEDED

Key Management Errors (3000-3999)
- 3001: KEY_NOT_FOUND
- 3002: KEY_GENERATION_FAILED
- 3003: KEY_EXPIRED
- 3004: KEY_REVOKED
- 3005: INVALID_KEY_TYPE
- 3006: HSM_UNAVAILABLE

Wallet Errors (4000-4999)
- 4001: WALLET_NOT_FOUND
- 4002: INVALID_DERIVATION_PATH
- 4003: MNEMONIC_GENERATION_FAILED
- 4004: ADDRESS_GENERATION_FAILED

Signature Errors (5000-5999)
- 5001: SIGNATURE_GENERATION_FAILED
- 5002: SIGNATURE_VERIFICATION_FAILED
- 5003: INVALID_SIGNATURE_FORMAT
- 5004: HASH_ALGORITHM_NOT_SUPPORTED

System Errors (9000-9999)
- 9001: INTERNAL_SERVER_ERROR
- 9002: DATABASE_UNAVAILABLE
- 9003: SERVICE_UNAVAILABLE
- 9004: CONFIGURATION_ERROR
```

### Error Response Format

```json
{
  "success": false,
  "error": {
    "code": "KEY_NOT_FOUND",
    "message": "The specified key was not found",
    "details": {
      "key_id": "uuid",
      "requested_at": "2024-06-16T12:00:00Z"
    }
  },
  "metadata": {
    "timestamp": "2024-06-16T12:00:00Z",
    "requestId": "uuid",
    "version": "v1"
  }
}
```

## API Security Specifications

### Authentication

- **Bearer Token**: JWT tokens for user authentication
- **API Key**: For service-to-service communication
- **mTLS**: For highly secure environments

### Rate Limiting

- **Per User**: 1000 requests/hour for standard operations
- **Per API Key**: 10,000 requests/hour for service accounts
- **Per Endpoint**: Specific limits for resource-intensive operations

### Input Validation

- All inputs validated against OpenAPI schema
- SQL injection protection
- XSS protection
- Parameter tampering protection

### Audit Requirements

- All API calls logged with full context
- Sensitive data redacted in logs
- Tamper-evident log storage
- Real-time security monitoring

## Versioning Strategy

### URL Versioning

- Current version: `/api/v1/`
- Future versions: `/api/v2/`, `/api/v3/`, etc.

### Backward Compatibility

- Minimum 2 versions supported simultaneously
- 6-month deprecation notice for breaking changes
- Graceful degradation for removed features

### API Evolution

- Additive changes within same version
- Breaking changes require new version
- Feature flags for experimental features

---

**Document Version**: 1.0  
**Last Updated**: June 16, 2024  
**Next Review**: September 16, 2024  
**Owner**: API Team  
**Approvers**: Architecture Team, Security Team
