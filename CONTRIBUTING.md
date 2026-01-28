# Contributing to Tharos

Thank you for your interest in contributing to Tharos! 🦊

## Ways to Contribute

### 1. Report Bugs
- Use GitHub Issues
- Include reproduction steps
- Provide code samples
- Specify your environment (OS, Node version, etc.)

### 2. Suggest Features
- Open a GitHub Discussion
- Explain the use case
- Provide examples

### 3. Submit Code
- Fork the repository
- Create a feature branch
- Write tests
- Submit a pull request

## Development Setup

```bash
# Clone your fork
git clone https://github.com/YOUR_USERNAME/tharos.git
cd tharos

# Install dependencies
npm install

# Build Go core
cd go-core
go build -o tharos-core.exe main.go
cd ..

# Build CLI
npm run build

# Run tests
npm test
```

## Code Style

- **TypeScript**: Follow existing patterns, use ESLint
- **Go**: Use `gofmt` and `golint`
- **Commits**: Use conventional commits (feat:, fix:, docs:, etc.)

## Adding a New Policy

1. Create YAML file in `policies/`
2. Follow existing policy format
3. Add documentation to `policies/README.md`
4. Include compliance references
5. Add tests

## Pull Request Process

1. Update documentation
2. Add tests for new features
3. Ensure all tests pass
4. Update CHANGELOG.md
5. Request review

## Code of Conduct

Be respectful, inclusive, and professional. We're all here to build great software together.

## Questions?

Open a GitHub Discussion or join our Discord community.

Thank you for making Tharos better! 🦊
