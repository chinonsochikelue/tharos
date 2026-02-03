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
    // Build the Go binary for the current platform
    const buildCmd = isWindows
        ? 'go build -o ../dist/tharos.exe .'
        : 'go build -o ../dist/tharos .';

    execSync(buildCmd, {
        cwd: goCoreDir,
        stdio: 'inherit',
        env: {
            ...process.env,
            CGO_ENABLED: '0' // Disable CGO for static binary
        }
    });

    console.log('✅ Binary built successfully!');

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
