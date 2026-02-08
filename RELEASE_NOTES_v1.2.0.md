# 🚀 Tharos v1.2.0 - "Intelligence & Ecosystem"

**Release Date**: February 8, 2026

---

## 🎯 Overview

Tharos v1.2.0 represents a major evolution from a powerful security scanner into a **complete security ecosystem**. This release introduces the **Functional Security Playground**, **unified professional branding**, **optimized analysis engine**, and **enhanced multi-token pattern detection**.

---

## ✨ Major Features

### 🧪 Functional Security Playground
The documentation now features a **live, interactive playground** where users can test Tharos's capabilities directly in their browser.

- **Real-time Analysis**: Paste code and get instant security feedback
- **Powered by Go Engine**: Uses the actual Tharos binary via Next.js API route
- **Magic Fix Preview**: See AI-generated patches before applying
- **Zero Installation**: Experience Tharos before installing

**Try it now**: [https://tharos.vercel.app/playground](https://tharos.vercel.app/playground)

### 🎨 Unified Professional Branding
Tharos now has a consistent, premium visual identity across all touchpoints.

- **Stylized T Logo**: Sleek, professional branding
- **Removed Mascots**: Transitioned to enterprise-grade iconography (🛡️, 🧠, ✨)
- **Consistent UI/UX**: Unified design across Docs, Dashboard, and CLI

### ⚡ Analysis Engine Optimization
Dramatically improved performance and signal-to-noise ratio.

- **Smart Filtering**: Excludes `.next`, `bin`, `.vercel`, and other build artifacts
- **98% Noise Reduction**: Dashboard file count reduced from 4,807 to ~97 relevant files
- **Faster Scans**: Focus exclusively on primary source code

### 🔍 Enhanced Multi-Token Pattern Detection
Tharos now detects real-world security anti-patterns that span multiple tokens.

- **Sliding Window Buffer**: Tracks last 10 string tokens in JS/TS lexer
- **Express.js Patterns**: Catches `res.header("Access-Control-Allow-Origin", "*")`
- **Improved Accuracy**: Significantly better detection for Node.js server code

---

## 🛡️ Security Rules Added

### New Detection Capabilities:
- **Insecure CORS** (JS/TS): Detects wildcard CORS configurations
- **Insecure Headers** (JS/TS): Identifies disabled security headers like `X-Content-Type-Options`
- **Hardcoded Credentials** (Python): Tracks secrets in literal assignments
- **Insecure CORS** (Go): Detects `Access-Control-Allow-Origin: *` in Go servers

---

## 🔧 Improvements

### Analysis Engine
- Added sliding window buffer for multi-token pattern matching
- Improved token-by-token analysis with 10-token lookahead
- Enhanced detection for Express.js and Node.js patterns

### CLI & Dashboard
- Updated branding with professional security icons
- Improved verdict display with consistent styling
- Better error messages and remediation guidance

### Documentation
- Live playground integration
- Updated engine version displays
- Enhanced visual design

---

## 📦 Installation

### NPM (Recommended)
```bash
npm install -g @collabchron/tharos@1.2.0
```

### Verify Installation
```bash
tharos --version
# Output: 1.2.0
```

---

## 🎓 Getting Started

### 1. Initialize in Your Project
```bash
cd your-project
tharos init
```

### 2. Run Analysis
```bash
# Analyze specific file
tharos analyze src/api/login.ts

# Analyze entire project
tharos analyze .

# Launch interactive dashboard
tharos ui
```

### 3. Try the Playground
Visit [https://tharos.vercel.app/playground](https://tharos.vercel.app/playground) to test Tharos without installation.

---

## 🔄 Migration from v1.1.x

No breaking changes! Simply update:

```bash
npm install -g @collabchron/tharos@1.2.0
```

Your existing `tharos.yaml` configuration will continue to work.

---

## 🙏 Acknowledgments

Thank you to everyone who provided feedback and helped shape this release!

---

## 📚 Resources

- **Documentation**: [https://tharos.vercel.app](https://tharos.vercel.app)
- **Playground**: [https://tharos.vercel.app/playground](https://tharos.vercel.app/playground)
- **GitHub**: [https://github.com/chinonsochikelue/tharos](https://github.com/chinonsochikelue/tharos)
- **Issues**: [https://github.com/chinonsochikelue/tharos/issues](https://github.com/chinonsochikelue/tharos/issues)

---

**Tharos is now a world-class, enterprise-grade security powerhouse.** 🛡️✨🚀
