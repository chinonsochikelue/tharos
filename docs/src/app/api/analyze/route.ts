import { NextRequest, NextResponse } from 'next/server';
import { exec } from 'child_process';
import { promisify } from 'util';
import fs from 'fs/promises';
import path from 'path';
import os from 'os';

const execAsync = promisify(exec);

export async function POST(req: NextRequest) {
    try {
        const { code, ai, fix } = await req.json();

        if (!code) {
            return NextResponse.json({ error: 'No code provided' }, { status: 400 });
        }

        // Create a temporary file
        const tempDir = os.tmpdir();
        const tempFile = path.resolve(tempDir, `tharos-playground-${Date.now()}.ts`);
        await fs.writeFile(tempFile, code);

        // Path to tharos binary - adjust based on environment
        const isProd = process.env.NODE_ENV === 'production';
        const binaryName = process.platform === 'win32' ? 'tharos.exe' : 'tharos';
        let tharosPath: string;

        if (isProd) {
            // Vercel / Production: Expect binary in local bin/ or specific path
            const bundledBinaryPath = path.resolve(process.cwd(), 'bin', binaryName);
            const tmpBinaryPath = path.resolve('/tmp', binaryName);

            try {
                // Check if binary is already in /tmp and executable
                await fs.access(tmpBinaryPath, fs.constants.X_OK);
                tharosPath = tmpBinaryPath;
            } catch {
                // Not in /tmp or not executable, try to copy it from the bundle
                try {
                    await fs.access(bundledBinaryPath);
                    await fs.copyFile(bundledBinaryPath, tmpBinaryPath);
                    await fs.chmod(tmpBinaryPath, 0o755); // Make it executable
                    tharosPath = tmpBinaryPath;
                    console.log('✅ Copied Tharos binary to /tmp and set permissions');
                } catch (copyErr: any) {
                    console.error('❌ Failed to copy or chmod binary:', copyErr);
                    // Fallback to bundled path just in case
                    tharosPath = bundledBinaryPath;
                }
            }
        } else {
            // Local Dev
            tharosPath = path.resolve(process.cwd(), '../dist/tharos.exe');
        }

        let command = `"${tharosPath}" analyze "${tempFile}" --json`;
        if (ai) {
            command += ' --ai';
        }
        if (fix) {
            command += ' --fix';
        }

        try {
            const { stdout } = await execAsync(command, { env: { ...process.env } });
            const results = JSON.parse(stdout);

            let fixedCode = null;
            if (fix) {
                fixedCode = await fs.readFile(tempFile, 'utf-8');
            }

            // Cleanup
            await fs.unlink(tempFile);

            return NextResponse.json({ ...results, fixedCode });
        } catch (execError: any) {
            // tharos exits with 1 if issues are found, which captures as an error in execAsync
            if (execError.stdout) {
                try {
                    const results = JSON.parse(execError.stdout);

                    let fixedCode = null;
                    if (fix) {
                        fixedCode = await fs.readFile(tempFile, 'utf-8');
                    }

                    await fs.unlink(tempFile);
                    return NextResponse.json({ ...results, fixedCode });
                } catch (parseError) {
                    console.error('Failed to parse tharos output:', execError.stdout);
                }
            }

            try { await fs.unlink(tempFile); } catch (e) { }
            console.error('Tharos execution failed:', execError);
            return NextResponse.json({
                error: 'Analysis failed',
                details: execError.message,
                path: tharosPath,
                cwd: process.cwd(),
            }, { status: 500 });
        }
    } catch (err: any) {
        console.error('API Error:', err);
        return NextResponse.json({
            error: 'Internal server error',
            details: err.message
        }, { status: 500 });
    }
}
