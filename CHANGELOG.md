# Changelog

All notable changes to the Tharos project will be documented in this file.

## [1.1.1] - 2026-02-08

### Added
- **Release Automation**: Automated GitHub tagging and Release creation.
- **Community Health**: Added `SECURITY.md` and `SUPPORT.md`.
- **Expanded Security Engine**:
    - Go: Insecure CORS detection (`Access-Control-Allow-Origin: *`).
    - Python: Hardcoded Password tracking in literal assignments.
    - JS/TS: Insecure CORS and Insecure Header Configuration checks.
- **Visual Identity**: New "Fox Guardian" brand logo.

### Changed
- Improved SARIF compliance (StartLine >= 1).
- Hardened commit verdict logic (Critical/High findings now block commits).

## [1.1.0] - 2026-02-08

### 🚀 v1.1.0: The Enterprise Polyglot Update
This major release transforms Tharos into a high-fidelity, enterprise-ready security suite.

#### Added
- **🐹 Deep Go AST Analysis**: Native, structural security scanning for Go, detecting SQLi in `database/sql`, command injection in `os/exec`, and insecure TLS.
- **🐍 Deep Python AST Analysis**: Structural scanning for Python, detecting insecure deserialization (`pickle`, `yaml`), command injection, and unsafe `eval`.
- **📊 Local Security Dashboard (`tharos ui`)**: A high-fidelity, interactive web control center for browsing findings and risk scores directly from the terminal.
- **🦾 Official GitHub Action**: Introduced `tharos-action` for seamless integration with GitHub Security, populating the security tab with SARIF findings.
- **🏗️ Enterprise SARIF Exporter**: Refined SARIF engine with rich rule metadata, stable indexing, and precise location mapping.

#### Improved
- **Intelligent Gating**: Improved "Scanner Mindset" to handle complex variable flows and reduce false positives.
- **Documentation**: Comprehensive updates to the documentation site and README for all new features.

## [1.0.1] - 2026-02-05

### The VS Code & Launch Update
- **VS Code Extension (v1.0.2)**: Launched the official extension with Magic Fixes and bundled binaries.
- **Interactive Remediation**: Extended Magic Fixes to the IDE for one-click security patches.
- **Enterprise Enforcement**: Hardened exit code logic for CI/CD pipelines.

## [1.0.0] - 2026-02-01
- Initial public release of Tharos Core.
- AI-powered security analysis integration (Gemini/Groq).
- Initial Go-based security engine implementation.
