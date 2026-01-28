# Tharos Policy Library

This directory contains pre-defined security and compliance policies that can be used with Tharos.

## Available Policies

### Security Standards
- `owasp-top10.yaml` - OWASP Top 10 security risks
- `cwe-top25.yaml` - CWE Top 25 most dangerous software weaknesses
- `sans-top25.yaml` - SANS Top 25 programming errors

### Compliance Frameworks
- `soc2.yaml` - SOC 2 Type II compliance
- `gdpr.yaml` - GDPR data protection requirements
- `pci-dss.yaml` - PCI-DSS payment card security
- `hipaa.yaml` - HIPAA healthcare data protection

### Code Quality
- `code-quality.yaml` - General code quality best practices
- `performance.yaml` - Performance anti-patterns
- `accessibility.yaml` - Web accessibility (WCAG)

## Usage

1. **Copy a policy to your project:**
   ```bash
   cp policies/owasp-top10.yaml tharos.yaml
   ```

2. **Customize for your needs:**
   Edit the YAML file to adjust severity levels or add custom rules.

3. **Run Tharos:**
   ```bash
   tharos check
   ```

## Policy Format

Each policy file follows this structure:

```yaml
name: "Policy Name"
version: "1.0"
description: "Policy description"

rules:
  - id: "rule-001"
    name: "Rule Name"
    severity: "block" # or "warning"
    pattern: "regex pattern"
    message: "Error message"
    category: "security" # or "quality", "performance"
```

## Contributing

To add a new policy:
1. Create a new YAML file in this directory
2. Follow the policy format above
3. Add documentation to this README
4. Submit a pull request
