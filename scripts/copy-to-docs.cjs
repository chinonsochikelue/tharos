const fs = require('fs');
const path = require('path');

const rootDir = path.join(__dirname, '..');
const distDir = path.join(rootDir, 'dist');
const docsBinDir = path.join(rootDir, 'docs', 'bin');

const hostBinary = process.platform === 'win32' ? 'tharos.exe' : 'tharos';
const binariesToCopy = [
    { src: hostBinary, dest: hostBinary },
    { src: 'tharos-linux', dest: 'tharos' } // Essential for Vercel
];

console.log('🚚 Copying Tharos binaries to docs/bin...');

try {
    if (!fs.existsSync(docsBinDir)) {
        fs.mkdirSync(docsBinDir, { recursive: true });
    }

    for (const bin of binariesToCopy) {
        const sourcePath = path.join(distDir, bin.src);
        const destPath = path.join(docsBinDir, bin.dest);

        if (fs.existsSync(sourcePath)) {
            fs.copyFileSync(sourcePath, destPath);
            // Ensure executable permissions for Linux/Unix
            if (bin.dest === 'tharos') {
                fs.chmodSync(destPath, 0o755);
            }
            console.log(`✅ Successfully copied ${bin.src} to docs/bin/${bin.dest}`);
        } else {
            console.warn(`⚠️  Source binary not found at ${sourcePath}`);
        }
    }
} catch (error) {
    console.error('❌ Failed to copy binary:', error.message);
}
