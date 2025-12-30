# Enterprise Endpoint Agent - Security Architecture

## Overview

This document describes the security architecture of the Enterprise Endpoint Agent, designed to meet enterprise security standards and pass security reviews.

## Security Principles

### 1. Zero-Trust Architecture

The agent implements zero-trust principles:
- **Never trust, always verify**: Every API call requires mTLS authentication
- **Least privilege**: Agent runs with minimal required permissions
- **Assume breach**: All data in transit is encrypted, certificates can be revoked

### 2. Defense in Depth

Multiple security layers:
1. **Network**: mTLS with certificate pinning
2. **Application**: Input validation, secure coding practices
3. **Data**: Encryption at rest (DPAPI on Windows)
4. **Audit**: Comprehensive logging of security events

### 3. Transparency

- No stealth behavior or anti-debugging
- Clear documentation of data collection
- Clean install/uninstall process
- Appears in OS service lists

---

## Identity & Authentication

### Bootstrap Flow

```
┌─────────┐                    ┌─────────┐
│ Installer│                    │ Backend │
└────┬────┘                    └────┬────┘
     │                              │
     │ 1. Provide ORG_ID + TOKEN    │
     ├─────────────────────────────>│
     │                              │
     │ 2. Validate token            │
     │    Generate agent_id         │
     │    Issue X.509 cert          │
     │<─────────────────────────────┤
     │                              │
     │ 3. Store cert securely       │
     │    Clear install token       │
     │                              │
     │ 4. All future calls use mTLS │
     ├─────────────────────────────>│
     │<─────────────────────────────┤
```

### Certificate Management

**Certificate Properties:**
- **Type**: X.509 client certificate
- **Key Size**: RSA 2048-bit minimum (4096-bit recommended)
- **Validity**: 90 days
- **Rotation**: Automatic at 60 days (30-day buffer)
- **Subject CN**: `agent_id` (UUID)
- **Extended Key Usage**: Client Authentication

**Storage Security:**

| OS      | Method                          | Protection                                    |
|---------|---------------------------------|-----------------------------------------------|
| Windows | DPAPI-encrypted file            | User/machine-specific encryption              |
| Linux   | File with 0600 permissions      | Owner-only read/write                         |
| macOS   | File with 0600 permissions      | Owner-only read/write (Keychain optional)     |

**Revocation:**
- Server maintains Certificate Revocation List (CRL)
- Agent checks CRL on policy refresh
- Revoked agents cannot authenticate

---

## Data Security

### Data in Transit

**All network communications use mTLS:**
- TLS 1.2 minimum (TLS 1.3 preferred)
- Strong cipher suites only (no RC4, 3DES, MD5)
- Certificate pinning (optional, for high-security environments)

**Cipher Suite Preferences:**
```
TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384
TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256
TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305
```

### Data at Rest

**Configuration Files:**
- Permissions: 0600 (owner read/write only)
- Sensitive fields encrypted (Windows: DPAPI)

**Certificates:**
- Private keys: 0600 permissions, DPAPI on Windows
- Never transmitted after initial bootstrap

**Buffered Telemetry:**
- Stored in local queue when offline
- Max size: 100 MB (configurable)
- Cleared after successful transmission

### Data Minimization

**Default Collection (SAFE):**
- ✅ OS version, hostname, architecture
- ✅ CPU model, core count, usage %
- ✅ Memory total/used
- ✅ Disk mount points, usage
- ✅ Network interface names, IP addresses

**Opt-In Collection (PII):**
- ⚠️ MAC addresses (disabled by default)
- ⚠️ Custom collectors (must be explicitly enabled)

**Never Collected:**
- ❌ User credentials
- ❌ File contents
- ❌ Process command lines
- ❌ Browser history
- ❌ Keystrokes or screenshots

---

## Privilege Management

### Windows

**Service Account:**
- Default: `NT AUTHORITY\SYSTEM`
- Recommended: Dedicated service account with minimal privileges
- Required Permissions:
  - Read system information (WMI)
  - Write to `%ProgramData%\unitechio\Agent\`
  - Network access

**Security Hardening:**
- Service marked as "Protected Process" (optional)
- DEP (Data Execution Prevention) enabled
- ASLR (Address Space Layout Randomization) enabled

### Linux

**User Account:**
- Dedicated user: `your-agent`
- No shell access: `/usr/sbin/nologin`
- Minimal capabilities: `CAP_NET_ADMIN` (if network stats required)

**Systemd Hardening:**
```ini
[Service]
NoNewPrivileges=true
PrivateTmp=true
ProtectSystem=strict
ProtectHome=true
ReadWritePaths=/var/lib/your-agent
```

### macOS

**Daemon:**
- Runs as root (required for system monitoring)
- Sandboxed where possible
- Entitlements limited to necessary APIs

---

## Auto-Update Security

### Update Flow

```
1. Check /api/v1/updates/metadata (mTLS)
   ↓
2. Download binary from signed URL
   ↓
3. Verify SHA256 checksum
   ↓
4. Verify code signature
   - Windows: Authenticode (signtool)
   - macOS: Developer ID + Notarization
   - Linux: GPG signature
   ↓
5. Atomic swap (rename old → .old, new → current)
   ↓
6. Restart service
   ↓
7. Health check (30s timeout)
   ↓
8. Rollback if health check fails
```

### Signature Verification

**Windows:**
```powershell
signtool verify /pa /v agent.exe
```
- Must be signed by trusted publisher
- Certificate must be valid and not revoked

**macOS:**
```bash
codesign --verify --deep --strict agent
spctl --assess --verbose agent
```
- Must be signed with Developer ID
- Must be notarized by Apple

**Linux:**
```bash
gpg --verify agent.sig agent
```
- Must be signed by trusted GPG key
- Key must be in agent's keyring

---

## Audit Logging

### Security Events

All security-relevant events are logged:

| Event                  | Severity | Details                                      |
|------------------------|----------|----------------------------------------------|
| Agent start/stop       | INFO     | Timestamp, version                           |
| Bootstrap success      | INFO     | Agent ID, org ID                             |
| Bootstrap failure      | ERROR    | Reason, source IP                            |
| Certificate rotation   | INFO     | Old/new expiry dates                         |
| Policy change          | INFO     | Old/new policy versions                      |
| Update installed       | INFO     | Old/new versions                             |
| Update failed          | ERROR    | Reason, rollback status                      |
| Authentication failure | WARNING  | Endpoint, reason                             |
| Health check failure   | WARNING  | Metrics, thresholds                          |

### Log Format

**Structured JSON:**
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "event": "bootstrap_success",
  "agent_id": "550e8400-e29b-41d4-a716-446655440000",
  "org_id": "acme-corp",
  "version": "1.0.0"
}
```

### Log Retention

- **Local**: 90 days (rotated daily, max 100 MB)
- **Server**: Indefinite (compliance requirement)

---

## Compliance

### GDPR Readiness

**Data Subject Rights:**
- **Right to Access**: API endpoint to export agent data
- **Right to Deletion**: Uninstall removes all local data
- **Right to Portability**: Data exported in JSON format
- **Right to Object**: Collectors can be disabled per-endpoint

**Lawful Basis:**
- **Legitimate Interest**: System monitoring for security/performance
- **Consent**: Installation implies consent (documented in EULA)

**Data Processing Agreement:**
- Agent is a "data processor"
- Customer is "data controller"
- DPA defines data handling obligations

### SOC 2 Alignment

**Security Principles:**
- ✅ **Confidentiality**: mTLS, encryption at rest
- ✅ **Integrity**: Code signing, audit logs
- ✅ **Availability**: Auto-restart, health monitoring
- ✅ **Privacy**: Data minimization, opt-in PII
- ✅ **Processing Integrity**: Input validation, error handling

---

## Threat Model

### Threats Mitigated

| Threat                          | Mitigation                                   |
|---------------------------------|----------------------------------------------|
| Man-in-the-middle (MITM)        | mTLS with certificate pinning                |
| Credential theft                | No long-lived tokens, certificate-based auth |
| Unauthorized data access        | File permissions, DPAPI encryption           |
| Malicious updates               | Code signing, signature verification         |
| Privilege escalation            | Least privilege, systemd hardening           |
| Data exfiltration               | Policy-controlled collection, audit logs     |

### Residual Risks

| Risk                            | Likelihood | Impact | Mitigation Plan                              |
|---------------------------------|------------|--------|----------------------------------------------|
| Compromised CA                  | Low        | High   | Certificate revocation, key rotation         |
| Zero-day in Go runtime          | Medium     | Medium | Regular updates, vulnerability scanning      |
| Insider threat (admin access)   | Low        | High   | Audit logs, anomaly detection                |

---

## Security Review Checklist

### Code Security

- [x] No hardcoded secrets (API keys, passwords)
- [x] Input validation on all external data
- [x] Parameterized queries (if using SQL)
- [x] No use of `unsafe` package
- [x] Error messages don't leak sensitive info
- [x] Dependency scanning (Dependabot, Snyk)

### Authentication

- [x] mTLS for all API calls
- [x] Certificate validation (chain, expiry, revocation)
- [x] No fallback to insecure auth

### Authorization

- [x] Policy-based access control
- [x] Least privilege principle
- [x] Audit logging of policy changes

### Data Protection

- [x] Encryption in transit (TLS 1.2+)
- [x] Encryption at rest (sensitive data only)
- [x] Secure key storage (DPAPI, file permissions)
- [x] Data minimization

### Operational Security

- [x] Graceful shutdown (no data loss)
- [x] Health monitoring
- [x] Auto-restart on failure
- [x] Rollback on update failure
- [x] Clean uninstall (removes all data)

---

## Incident Response

### Security Incident Procedures

1. **Detection**: Anomaly in audit logs or health metrics
2. **Containment**: Revoke agent certificate
3. **Investigation**: Analyze logs, collect forensics
4. **Remediation**: Patch vulnerability, rotate keys
5. **Recovery**: Re-bootstrap agent with new certificate
6. **Lessons Learned**: Update threat model, improve controls

### Contact

Security issues: security@unitechio.com  
PGP Key: [Link to public key]

---

## Conclusion

This agent is designed to meet enterprise security standards through:
- Zero-trust architecture with mTLS
- Defense in depth with multiple security layers
- Transparency in operation and data collection
- Compliance with GDPR and SOC 2 principles

All design decisions prioritize security without sacrificing functionality or user experience.
