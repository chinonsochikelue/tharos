const fs = require('fs');
const path = require('path');

const rootDir = path.join(__dirname, '..');
const distDir = path.join(rootDir, 'dist');
const docsBinDir = path.join(rootDir, 'docs', 'bin');

const binaryName = process.platform === 'win32' ? 'tharos.exe' : 'tharos';
const sourcePath = path.join(distDir, binaryName);
const destPath = path.join(docsBinDir, binaryName);

console.log('🚚 Copying Tharos binary to docs/bin...');

try {
    if (!fs.existsSync(docsBinDir)) {
        fs.mkdirSync(docsBinDir, { recursive: true });
    }

    if (fs.existsSync(sourcePath)) {
        fs.copyFileSync(sourcePath, destPath);
        // On Unix-like systems, ensure it's executable
        if (process.platform !== 'win32') {
            fs.chmodSync(destPath, 0o755);
        }
        console.log(`✅ Successfully copied ${binaryName} to docs/bin`);
    } else {
        console.warn(`⚠️  Source binary not found at ${sourcePath}`);
    }
} catch (error) {
    console.error('❌ Failed to copy binary:', error.message);
}
