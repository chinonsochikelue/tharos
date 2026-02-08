
import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs';
import path from 'path';

const execAsync = promisify(exec);

async function main() {
    console.log('🏗️  Setting up Tharos binary...');

    const isVercel = process.env.VERCEL === '1';
    const rootDir = process.cwd(); // docs/
    const goCoreDir = path.resolve(rootDir, '../go-core');
    const binDir = path.resolve(rootDir, 'bin');

    console.log(`📂 Current Dir: ${rootDir}`);
    console.log(`📂 Go Core Dir: ${goCoreDir}`);

    if (!fs.existsSync(binDir)) {
        fs.mkdirSync(binDir, { recursive: true });
    }

    // Check if go-core exists
    if (!fs.existsSync(goCoreDir)) {
        console.warn('⚠️  go-core directory not found. Skipping binary build.');
        console.warn('   This is expected if "Root Directory" is set to "docs" in Vercel without including root files.');
        return;
    }

    const binaryName = process.platform === 'win32' ? 'tharos.exe' : 'tharos';
    const outputPath = path.resolve(binDir, binaryName);
    const rootDistPath = path.resolve(rootDir, '../dist', binaryName);

    console.log(`🔍 Checking Root Dist: ${rootDistPath}`);
    console.log(`🎯 Target Output: ${outputPath}`);

    // 1. Try to use existing build from root dist (User's suggestion)
    if (fs.existsSync(rootDistPath)) {
        console.log(`📦 Found pre-built binary in root dist: ${rootDistPath}`);
        try {
            fs.copyFileSync(rootDistPath, outputPath);
            fs.chmodSync(outputPath, 0o755);
            console.log(`✅ Successfully copied binary to ${outputPath}`);
            return;
        } catch (err: any) {
            console.warn(`⚠️  Failed to copy binary from root dist: ${err.message}`);
        }
    }

    // 2. Fallback to building from source
    console.log('⚙️  Building Tharos binary from source...');
    try {
        // Check if go is installed
        try {
            const { stdout: goVer } = await execAsync('go version');
            console.log(`📟 Go version: ${goVer.trim()}`);
        } catch {
            console.error('❌ Go is not installed in the build environment!');
            console.error('   Please ensure Go is available in Vercel.');
            return;
        }

        // Build
        console.log(`🚀 Executing: go build -o "${outputPath}" . in ${goCoreDir}`);
        await execAsync(`go build -o "${outputPath}" .`, {
            cwd: goCoreDir,
            env: { ...process.env, CGO_ENABLED: '0' }
        });

        if (fs.existsSync(outputPath)) {
            console.log(`✅ Binary built successfully at: ${outputPath}`);
            fs.chmodSync(outputPath, 0o755);
        } else {
            console.error('❌ Binary built but not found at output path!');
        }

    } catch (error: any) {
        console.error('❌ Failed to build Tharos binary:', error.message);
    }
}

main().catch(console.error);
