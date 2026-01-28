# Tharos VSCode Extension

AI-powered semantic code analysis with real-time security insights for TypeScript, JavaScript, Python, Go, Rust, and more.

## Features

- 🔴 **Real-time Analysis**: See security issues as you type
- 🧠 **AI-Powered Insights**: Get intelligent recommendations with risk scores
- 💡 **Quick Fixes**: Apply suggested fixes with one click
- 🌍 **Multi-Language**: Supports TypeScript, JavaScript, Python, Go, Rust, Java
- ⚡ **Fast**: Uses high-performance Go core for instant analysis

## Installation

1. Install the extension from the VSCode Marketplace
2. Open a project with supported files
3. Tharos will automatically analyze your code on save

## Configuration

- `tharos.enableAI`: Enable/disable AI-powered analysis (default: true)
- `tharos.corePath`: Custom path to tharos-core executable
- `tharos.severity`: Minimum severity level to show (block/warning/info)

## Usage

### Automatic Analysis
Tharos automatically analyzes files when you:
- Open a file
- Save a file
- Switch between files

### Manual Analysis
- **Analyze Current File**: `Ctrl+Shift+P` → "Tharos: Analyze Current File"
- **Analyze Workspace**: `Ctrl+Shift+P` → "Tharos: Analyze Entire Workspace"

### Viewing Insights
- **Diagnostics**: Red/yellow squiggles appear under issues
- **Hover**: Hover over any line to see detailed AI insights
- **Quick Fixes**: Click the lightbulb (💡) to apply suggested fixes

## Requirements

- VSCode 1.85.0 or higher
- Tharos core binary (auto-detected in workspace)

## License

MIT
