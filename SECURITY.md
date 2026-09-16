# Security Policy

## Supported Versions

`go-fintech-india` provides security updates and vulnerability patches for the current release series:

| Version Series | Status             | Security Updates |
| -------------- | ------------------ | ---------------- |
| 0.2.x          | **Active / Current** | :white_check_mark: Yes |
| < 0.2.0        | End-of-Life        | :x: No           |

---

## Reporting a Vulnerability

The maintainers take security, financial integrity, and data protection very seriously. If you discover a security vulnerability or potential threat within this library, **please do not disclose it publicly** via GitHub issues, discussions, or pull requests.

Instead, please report vulnerabilities through one of the following confidential channels:

### 1. GitHub Private Vulnerability Reporting (Preferred)
- Navigate to the **Security** tab of `github.com/umesh0492/go-fintech-india`.
- Click **"Report a vulnerability"** to submit a draft security advisory.
- Include a detailed description, affected functions/files, proof of concept (PoC), and potential security or compliance impact.

### 2. Direct Security Contact
- **Email**: [umesh0492@gmail.com](mailto:umesh0492@gmail.com)
- **Subject**: `[SECURITY] go-fintech-india Vulnerability Report`

### Response SLA
- **Initial Acknowledgement**: Within **48 hours**.
- **Triage & Assessment**: Within **5 business days**.
- **Fix & Coordinated Release**: Targeted within **14 business days**.

---

## Threat Model & Security Architecture

`go-fintech-india` is designed for mission-critical financial, banking, and identity workflows. It incorporates strict security safeguards against the following threat vectors:

### 1. PII (Personally Identifiable Information) Protection & Masking
- **Aadhaar Data Protection (UIDAI Compliance)**:
  - Raw 12-digit Aadhaar numbers constitute sensitive biometric-linked PII under the Aadhaar Act.
  - The library provides statutory masking (`XXXX-XXXX-1234`) ensuring that only the last 4 digits remain unmasked in logs, database displays, and receipts.
  - Memory representations of validated Aadhaar numbers should not be logged or exposed unmasked.
- **PAN (Permanent Account Number) Masking**:
  - Masking utilities (`XXXXX1234X`) preserve entity-type classification while obscuring the individual tax identifier in non-reconciliation layers.
- **Bank Account Number Masking**:
  - Account numbers (9 to 18 digits) are masked with only the last 4 digits visible to prevent unauthorized account disclosure in telemetry or customer-facing receipts.

### 2. Integer Paise Financial Safety (IEEE 754 Floating-Point Immunity)
- **Problem**: Standard IEEE 754 floating-point types (`float32`, `float64`) inherently suffer from binary representation errors (e.g., `0.1 + 0.2 != 0.3`). In financial systems, accumulating rounding errors can lead to ledger imbalance, reconciliation failure, and arbitrage.
- **Safeguard**:
  - All monetary values are represented strictly as `int64` paise (1 INR = 100 paise).
  - Remainder-safe split and proportional allocation algorithms guarantee the invariant:
    $$\sum_{i=1}^{n} \text{split}_i = \text{original\_amount}$$
  - No paise is ever created or discarded due to truncation or integer division remainders.

### 3. Supply-Chain & Dependency Hardening
- **Zero External Dependencies**:
  - The entire library relies exclusively on the Go standard library (`math`, `crypto`, `database/sql`, `encoding/json`, `time`, `regexp`, `strings`).
  - Zero transitive dependencies eliminate the risk of upstream dependency hijacking, typosquatting, and unvetted third-party code execution.
