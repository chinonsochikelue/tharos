const { execSync } = require('child_process');
const fs = require('fs');
const path = require('path');

const isWindows = process.platform === 'win32';
const goCoreDir = path.join(__dirname, '..', 'go-core');
const distDir = path.join(__dirname, '..', 'dist');

// Ensure dist directory exists
if (!fs.existsSync(distDir)) {
    fs.mkdirSync(distDir, { recursive: true });
}

console.log('🔨 Building Tharos binary...');

try {
    // 1. Build host binary
    console.log(`🔨 Building binary for host platform (${process.platform})...`);
    const hostBinary = isWindows ? 'tharos.exe' : 'tharos';
    execSync(`go build -o ../dist/${hostBinary} .`, {
        cwd: goCoreDir,
        stdio: 'inherit',
        env: { ...process.env, CGO_ENABLED: '0' }
    });

    // 2. Build Linux binary (Always needed for Vercel/Production)
    console.log('🔨 Building binary for Linux (Vercel target)...');
    execSync('go build -o ../dist/tharos-linux .', {
        cwd: goCoreDir,
        stdio: 'inherit',
        env: { ...process.env, GOOS: 'linux', GOARCH: 'amd64', CGO_ENABLED: '0' }
    });

    console.log('✅ Binaries built successfully!');

    // Make executable on Unix systems
    if (!isWindows) {
        const binaryPath = path.join(distDir, 'tharos');
        fs.chmodSync(binaryPath, '755');
        console.log('✅ Made binary executable');
    }

} catch (error) {
    console.error('❌ Failed to build binary:', error.message);
    process.exit(1);
}
