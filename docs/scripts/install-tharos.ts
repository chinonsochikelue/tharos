
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

    console.log('⚙️  Building Tharos binary...');
    try {
        // Check if go is installed
        try {
            await execAsync('go version');
        } catch {
            console.error('❌ Go is not installed in the build environment!');
            console.error('   Please ensure Go is available in Vercel (e.g., by adding a go.mod in the docs root or custom build script).');
            return;
        }

        // Install dependencies just in case
        // await execAsync('go mod download', { cwd: goCoreDir });

        // Build
        await execAsync(`go build -o "${outputPath}" .`, {
            cwd: goCoreDir,
            env: { ...process.env, CGO_ENABLED: '0' }
        });
        console.log('✅ Tharos binary built successfully!');

        // Verify
        if (fs.existsSync(outputPath)) {
            console.log(`📦 Binary located at: ${outputPath}`);
        } else {
            console.error('❌ Binary file missing after build!');
            process.exit(1);
        }

    } catch (error: any) {
        console.error('❌ Failed to build Tharos binary:', error.message);
        // process.exit(1); // Don't fail build for now, might be just docs deployment
    }
}

main().catch(console.error);
