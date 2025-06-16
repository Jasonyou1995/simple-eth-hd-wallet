# SKMS Database Schema Design

## Overview

The Secure Key Management System uses a hybrid data storage approach combining PostgreSQL for metadata and structured data, Redis for caching and session management, and specialized secure storage for cryptographic keys.

## Database Architecture

### Primary Database: PostgreSQL 15+

- **Purpose**: Metadata, user management, audit logs, system configuration
- **Features**: ACID compliance, strong consistency, advanced indexing
- **Security**: Row-level security (RLS), encryption at rest, SSL/TLS

### Cache Layer: Redis 7+

- **Purpose**: Session management, API rate limiting, temporary data
- **Features**: In-memory performance, persistence options, clustering
- **Security**: AUTH mechanism, SSL/TLS, key expiration

### Secure Key Storage

- **Purpose**: Encrypted cryptographic keys, sensitive data
- **Options**: HSM, cloud KMS, encrypted file system
- **Security**: Hardware-level protection, key wrapping, audit trails

## PostgreSQL Schema Design

### Core Tables

#### 1. Users and Authentication

```sql
-- Users table
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(255) NOT NULL UNIQUE,
    email VARCHAR(255) NOT NULL UNIQUE,
    password_hash VARCHAR(255) NOT NULL, -- Argon2 hash
    salt VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    is_admin BOOLEAN NOT NULL DEFAULT false,
    mfa_enabled BOOLEAN NOT NULL DEFAULT false,
    mfa_secret VARCHAR(255), -- TOTP secret
    failed_login_attempts INTEGER NOT NULL DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    last_login TIMESTAMP WITH TIME ZONE,
    password_changed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT users_username_length CHECK (length(username) >= 3),
    CONSTRAINT users_email_format CHECK (email ~* '^[A-Za-z0-9._%+-]+@[A-Za-z0-9.-]+\.[A-Za-z]{2,}$')
);

-- User roles
CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    permissions JSONB NOT NULL DEFAULT '[]',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- User role assignments
CREATE TABLE user_roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by UUID NOT NULL REFERENCES users(id),
    assigned_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    UNIQUE(user_id, role_id)
);

-- API keys for service authentication
CREATE TABLE api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE, -- SHA-256 hash
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    permissions JSONB NOT NULL DEFAULT '[]',
    rate_limit INTEGER NOT NULL DEFAULT 1000, -- requests per hour
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_used TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT api_keys_name_length CHECK (length(name) >= 3)
);
```

#### 2. Cryptographic Keys Management

```sql
-- Key metadata (no actual key material stored here)
CREATE TABLE keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    key_type VARCHAR(50) NOT NULL, -- rsa, ecdsa, ed25519, aes
    key_size INTEGER NOT NULL,
    usage_types VARCHAR(50)[] NOT NULL, -- {signing, encryption, key_agreement}
    status VARCHAR(20) NOT NULL DEFAULT 'active', -- active, expired, revoked, compromised

    -- Cryptographic properties
    algorithm VARCHAR(100) NOT NULL,
    curve_name VARCHAR(50), -- For ECC keys
    public_key TEXT, -- Base64 encoded public key
    fingerprint VARCHAR(255) NOT NULL UNIQUE, -- SHA-256 of public key

    -- Key management
    key_store_type VARCHAR(50) NOT NULL, -- hsm, software, cloud_kms
    key_store_id VARCHAR(255) NOT NULL, -- Reference to key in external store
    wrapped_key_material TEXT, -- Encrypted private key (if software)

    -- Lifecycle
    created_by UUID NOT NULL REFERENCES users(id),
    expires_at TIMESTAMP WITH TIME ZONE,
    revoked_at TIMESTAMP WITH TIME ZONE,
    revoked_by UUID REFERENCES users(id),
    revocation_reason VARCHAR(255),

    -- Metadata
    tags JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT keys_name_length CHECK (length(name) >= 1),
    CONSTRAINT keys_valid_type CHECK (key_type IN ('rsa', 'ecdsa', 'ed25519', 'aes', 'chacha20')),
    CONSTRAINT keys_valid_status CHECK (status IN ('active', 'expired', 'revoked', 'compromised')),
    CONSTRAINT keys_usage_not_empty CHECK (array_length(usage_types, 1) > 0)
);

-- Key relationships (for key hierarchies, rotation, etc.)
CREATE TABLE key_relationships (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_key_id UUID NOT NULL REFERENCES keys(id) ON DELETE CASCADE,
    child_key_id UUID NOT NULL REFERENCES keys(id) ON DELETE CASCADE,
    relationship_type VARCHAR(50) NOT NULL, -- rotation, derivation, backup
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(parent_key_id, child_key_id, relationship_type),
    CONSTRAINT key_relationships_valid_type CHECK (
        relationship_type IN ('rotation', 'derivation', 'backup', 'recovery')
    )
);
```

#### 3. HD Wallets

```sql
-- HD Wallet instances
CREATE TABLE hd_wallets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,
    blockchain VARCHAR(50) NOT NULL, -- ethereum, bitcoin, solana

    -- BIP specifications
    derivation_path VARCHAR(255) NOT NULL, -- m/44'/60'/0'/0
    master_public_key TEXT NOT NULL, -- xpub key
    master_key_id UUID NOT NULL REFERENCES keys(id), -- Reference to master private key

    -- Configuration
    mnemonic_strength INTEGER NOT NULL DEFAULT 128, -- 128 or 256 bits
    has_passphrase BOOLEAN NOT NULL DEFAULT false,
    address_gap_limit INTEGER NOT NULL DEFAULT 20,

    -- State
    next_external_index INTEGER NOT NULL DEFAULT 0,
    next_internal_index INTEGER NOT NULL DEFAULT 0,
    total_addresses INTEGER NOT NULL DEFAULT 0,

    -- Metadata
    tags JSONB DEFAULT '[]',
    metadata JSONB DEFAULT '{}',

    -- Audit
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT hd_wallets_name_length CHECK (length(name) >= 1),
    CONSTRAINT hd_wallets_valid_blockchain CHECK (
        blockchain IN ('ethereum', 'bitcoin', 'solana', 'cardano', 'polkadot')
    ),
    CONSTRAINT hd_wallets_valid_strength CHECK (mnemonic_strength IN (128, 256))
);

-- Generated addresses from HD wallets
CREATE TABLE hd_addresses (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    wallet_id UUID NOT NULL REFERENCES hd_wallets(id) ON DELETE CASCADE,

    -- Derivation info
    derivation_index INTEGER NOT NULL,
    is_change BOOLEAN NOT NULL DEFAULT false, -- BIP44 change address
    full_derivation_path VARCHAR(255) NOT NULL,

    -- Address info
    address VARCHAR(255) NOT NULL,
    public_key TEXT NOT NULL,
    address_type VARCHAR(50), -- legacy, segwit, native_segwit (for Bitcoin)

    -- Usage tracking
    is_used BOOLEAN NOT NULL DEFAULT false,
    first_used_at TIMESTAMP WITH TIME ZONE,
    last_used_at TIMESTAMP WITH TIME ZONE,
    transaction_count INTEGER NOT NULL DEFAULT 0,

    -- Metadata
    label VARCHAR(255),
    metadata JSONB DEFAULT '{}',

    -- Audit
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(wallet_id, derivation_index, is_change),
    UNIQUE(address), -- Addresses should be globally unique

    CONSTRAINT hd_addresses_index_positive CHECK (derivation_index >= 0)
);
```

#### 4. Digital Signatures

```sql
-- Signature operations
CREATE TABLE signatures (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    key_id UUID NOT NULL REFERENCES keys(id),

    -- Signature data
    signature_value TEXT NOT NULL, -- Base64 encoded signature
    signature_algorithm VARCHAR(100) NOT NULL,
    hash_algorithm VARCHAR(50) NOT NULL,
    signature_format VARCHAR(20) NOT NULL, -- raw, asn1, jose

    -- Input data
    data_hash VARCHAR(255) NOT NULL, -- SHA-256 of signed data
    data_size INTEGER NOT NULL,

    -- Context
    purpose VARCHAR(255),
    description TEXT,

    -- Verification
    is_verified BOOLEAN NOT NULL DEFAULT false,
    verified_at TIMESTAMP WITH TIME ZONE,

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT signatures_valid_format CHECK (
        signature_format IN ('raw', 'asn1', 'jose', 'der')
    ),
    CONSTRAINT signatures_data_size_positive CHECK (data_size > 0)
);

-- Multi-signature schemes
CREATE TABLE multisig_schemes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Scheme configuration
    threshold INTEGER NOT NULL,
    total_participants INTEGER NOT NULL,
    scheme_type VARCHAR(50) NOT NULL, -- threshold, weighted

    -- Status
    status VARCHAR(20) NOT NULL DEFAULT 'active',

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT multisig_schemes_valid_threshold CHECK (threshold > 0),
    CONSTRAINT multisig_schemes_threshold_le_total CHECK (threshold <= total_participants),
    CONSTRAINT multisig_schemes_valid_type CHECK (scheme_type IN ('threshold', 'weighted'))
);

-- Multi-signature participants
CREATE TABLE multisig_participants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id UUID NOT NULL REFERENCES multisig_schemes(id) ON DELETE CASCADE,
    key_id UUID NOT NULL REFERENCES keys(id),

    -- Participant info
    participant_index INTEGER NOT NULL,
    weight INTEGER NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT true,

    -- Audit
    added_by UUID NOT NULL REFERENCES users(id),
    added_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    UNIQUE(scheme_id, participant_index),
    UNIQUE(scheme_id, key_id),

    CONSTRAINT multisig_participants_weight_positive CHECK (weight > 0)
);

-- Multi-signature operations
CREATE TABLE multisig_operations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scheme_id UUID NOT NULL REFERENCES multisig_schemes(id),

    -- Operation data
    data_hash VARCHAR(255) NOT NULL,
    operation_type VARCHAR(50) NOT NULL, -- sign, verify
    status VARCHAR(20) NOT NULL DEFAULT 'pending',

    -- Signature collection
    required_signatures INTEGER NOT NULL,
    collected_signatures INTEGER NOT NULL DEFAULT 0,
    final_signature TEXT, -- Combined signature

    -- Metadata
    metadata JSONB DEFAULT '{}',

    -- Audit
    initiated_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT multisig_operations_valid_status CHECK (
        status IN ('pending', 'completed', 'failed', 'expired')
    )
);
```

#### 5. Audit and Compliance

```sql
-- Comprehensive audit log
CREATE TABLE audit_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Event identification
    event_type VARCHAR(100) NOT NULL,
    event_category VARCHAR(50) NOT NULL, -- authentication, authorization, key_management, api_access
    action VARCHAR(100) NOT NULL, -- create, read, update, delete, sign, verify

    -- Subject and object
    user_id UUID REFERENCES users(id),
    resource_type VARCHAR(100), -- key, wallet, signature, user
    resource_id UUID,

    -- Request context
    session_id VARCHAR(255),
    request_id VARCHAR(255),
    api_endpoint VARCHAR(255),
    http_method VARCHAR(10),
    ip_address INET,
    user_agent TEXT,

    -- Result
    result VARCHAR(20) NOT NULL, -- success, failure, error
    error_code VARCHAR(100),
    error_message TEXT,

    -- Details
    details JSONB DEFAULT '{}',
    sensitive_data_hash VARCHAR(255), -- Hash of sensitive data for integrity

    -- Compliance
    retention_until TIMESTAMP WITH TIME ZONE, -- Data retention policy
    is_suspicious BOOLEAN NOT NULL DEFAULT false,
    compliance_flags VARCHAR(50)[], -- gdpr, sox, pci_dss

    -- Immutability
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    checksum VARCHAR(255) NOT NULL, -- For tamper detection

    CONSTRAINT audit_logs_valid_result CHECK (result IN ('success', 'failure', 'error')),
    CONSTRAINT audit_logs_valid_category CHECK (
        event_category IN ('authentication', 'authorization', 'key_management', 'api_access', 'system')
    )
);

-- System events and monitoring
CREATE TABLE system_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Event info
    event_type VARCHAR(100) NOT NULL, -- startup, shutdown, error, warning, info
    severity VARCHAR(20) NOT NULL, -- critical, high, medium, low, info
    component VARCHAR(100) NOT NULL, -- api, crypto_engine, hsm, database

    -- Message
    message TEXT NOT NULL,
    details JSONB DEFAULT '{}',

    -- Context
    hostname VARCHAR(255),
    process_id INTEGER,
    thread_id VARCHAR(100),

    -- Correlation
    trace_id VARCHAR(255),
    span_id VARCHAR(255),

    -- Timing
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT system_events_valid_severity CHECK (
        severity IN ('critical', 'high', 'medium', 'low', 'info')
    )
);
```

#### 6. Configuration and Settings

```sql
-- System configuration
CREATE TABLE system_config (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Configuration key
    config_key VARCHAR(255) NOT NULL UNIQUE,
    config_value JSONB NOT NULL,

    -- Metadata
    description TEXT,
    is_sensitive BOOLEAN NOT NULL DEFAULT false,
    is_readonly BOOLEAN NOT NULL DEFAULT false,

    -- Validation
    value_type VARCHAR(50) NOT NULL, -- string, integer, boolean, json
    validation_rules JSONB,

    -- Versioning
    version INTEGER NOT NULL DEFAULT 1,
    previous_value JSONB,

    -- Audit
    updated_by UUID REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT system_config_valid_type CHECK (
        value_type IN ('string', 'integer', 'boolean', 'json', 'array')
    )
);

-- Feature flags for controlled rollouts
CREATE TABLE feature_flags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),

    -- Flag definition
    flag_name VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,
    is_enabled BOOLEAN NOT NULL DEFAULT false,

    -- Targeting
    enabled_for_users UUID[], -- Specific users
    enabled_for_roles UUID[], -- Specific roles
    enabled_percentage INTEGER NOT NULL DEFAULT 0, -- Percentage rollout

    -- Metadata
    created_by UUID NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),

    CONSTRAINT feature_flags_valid_percentage CHECK (
        enabled_percentage >= 0 AND enabled_percentage <= 100
    )
);
```

## Indexes and Performance Optimization

### Essential Indexes

```sql
-- Users table indexes
CREATE INDEX idx_users_username ON users(username);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_active ON users(is_active) WHERE is_active = true;
CREATE INDEX idx_users_last_login ON users(last_login DESC);

-- Keys table indexes
CREATE INDEX idx_keys_status ON keys(status);
CREATE INDEX idx_keys_type ON keys(key_type);
CREATE INDEX idx_keys_created_by ON keys(created_by);
CREATE INDEX idx_keys_expires_at ON keys(expires_at) WHERE expires_at IS NOT NULL;
CREATE INDEX idx_keys_fingerprint ON keys(fingerprint);
CREATE INDEX idx_keys_tags ON keys USING GIN(tags);

-- HD Wallets indexes
CREATE INDEX idx_hd_wallets_blockchain ON hd_wallets(blockchain);
CREATE INDEX idx_hd_wallets_created_by ON hd_wallets(created_by);
CREATE INDEX idx_hd_addresses_wallet_id ON hd_addresses(wallet_id);
CREATE INDEX idx_hd_addresses_address ON hd_addresses(address);
CREATE INDEX idx_hd_addresses_used ON hd_addresses(is_used);

-- Signatures indexes
CREATE INDEX idx_signatures_key_id ON signatures(key_id);
CREATE INDEX idx_signatures_created_by ON signatures(created_by);
CREATE INDEX idx_signatures_created_at ON signatures(created_at DESC);
CREATE INDEX idx_signatures_algorithm ON signatures(signature_algorithm);

-- Audit logs indexes
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_created_at ON audit_logs(created_at DESC);
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_resource ON audit_logs(resource_type, resource_id);
CREATE INDEX idx_audit_logs_result ON audit_logs(result);
CREATE INDEX idx_audit_logs_ip_address ON audit_logs(ip_address);

-- System events indexes
CREATE INDEX idx_system_events_created_at ON system_events(created_at DESC);
CREATE INDEX idx_system_events_severity ON system_events(severity);
CREATE INDEX idx_system_events_component ON system_events(component);
```

### Partitioning Strategy

```sql
-- Partition audit_logs by month for better performance
CREATE TABLE audit_logs_y2024m06 PARTITION OF audit_logs
FOR VALUES FROM ('2024-06-01') TO ('2024-07-01');

-- Partition system_events by severity and date
CREATE TABLE system_events_critical PARTITION OF system_events
FOR VALUES IN ('critical', 'high');
```

## Redis Schema Design

### Key Patterns

```
# User sessions
session:{session_id} -> {user_data, permissions, expires_at}

# API rate limiting
ratelimit:{api_key}:{endpoint}:{hour} -> {request_count}
ratelimit:{user_id}:{endpoint}:{hour} -> {request_count}

# Temporary data
temp:{operation_id} -> {operation_data, expires_in}

# Cache
cache:key:{key_id} -> {key_metadata}
cache:user:{user_id} -> {user_profile}
cache:wallet:{wallet_id} -> {wallet_info}

# Locks for critical operations
lock:{resource_type}:{resource_id} -> {lock_holder, expires_at}

# MFA tokens
mfa:{user_id}:{token} -> {expires_at}

# Password reset tokens
reset:{token} -> {user_id, expires_at}
```

### TTL Policies

```
# Sessions: 8 hours
# Rate limit counters: 1 hour
# Temporary data: Variable (1 minute to 24 hours)
# Cache: 30 minutes to 24 hours
# Locks: 5 minutes maximum
# MFA tokens: 30 seconds
# Reset tokens: 15 minutes
```

## Data Security Measures

### Encryption

- **Database encryption at rest**: PostgreSQL transparent data encryption
- **Connection encryption**: SSL/TLS for all database connections
- **Field-level encryption**: Sensitive fields encrypted using envelope encryption
- **Key material**: Never stored in PostgreSQL, always in HSM or encrypted

### Access Control

- **Row-level security**: PostgreSQL RLS for multi-tenant isolation
- **Column-level permissions**: Restrict access to sensitive columns
- **Database users**: Separate users for different application components
- **Connection pooling**: PgBouncer for efficient connection management

### Backup and Recovery

- **Continuous archiving**: PostgreSQL WAL archiving
- **Point-in-time recovery**: Full PITR capability
- **Encrypted backups**: All backups encrypted at rest
- **Geographic replication**: Cross-region backup storage

### Compliance

- **Audit trail**: Complete audit trail for all data changes
- **Data retention**: Automated data retention policies
- **Right to erasure**: GDPR-compliant data deletion
- **Data anonymization**: PII anonymization for non-production environments

## Migration Strategy

### Version Control

```sql
-- Schema version tracking
CREATE TABLE schema_migrations (
    version VARCHAR(255) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    execution_time INTERVAL,
    checksum VARCHAR(255)
);
```

### Migration Process

1. **Schema changes**: DDL migrations with rollback scripts
2. **Data migrations**: Separate data transformation scripts
3. **Index creation**: Non-blocking index creation
4. **Validation**: Data integrity checks after migration

---

**Document Version**: 1.0  
**Last Updated**: June 16, 2024  
**Next Review**: September 16, 2024  
**Owner**: Database Team  
**Approvers**: Architecture Team, Security Team
