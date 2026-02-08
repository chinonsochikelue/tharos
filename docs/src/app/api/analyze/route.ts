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
            // Vercel / Production: Hunt for the binary in multiple potential locations
            const searchPaths = [
                path.resolve(process.cwd(), 'bin', binaryName),
                path.resolve(process.cwd(), '../dist', binaryName), // If root is repo root
                path.resolve(process.cwd(), 'dist', binaryName),    // If docs is repo root and dist was copied
                path.resolve('/tmp', binaryName)                    // Final fallback to writable /tmp
            ];

            const tmpBinaryPath = path.resolve('/tmp', binaryName);
            let foundPath = null;

            // 1. Try to find an existing executable
            for (const p of searchPaths) {
                try {
                    await fs.access(p, fs.constants.X_OK);
                    foundPath = p;
                    console.log(`🔍 Found executable binary at: ${p}`);
                    break;
                } catch { /* continue */ }
            }

            // 2. If not found or not executable, try to copy it to /tmp if we find a raw binary
            if (!foundPath) {
                for (const p of searchPaths) {
                    if (p === tmpBinaryPath) continue;
                    try {
                        await fs.access(p);
                        console.log(`📦 Found binary at ${p}, copying to /tmp...`);
                        await fs.copyFile(p, tmpBinaryPath);
                        await fs.chmod(tmpBinaryPath, 0o755);
                        foundPath = tmpBinaryPath;
                        break;
                    } catch { /* continue */ }
                }
            }

            if (!foundPath) {
                // LAST DITCH: Search the entire /var/task for anything named "tharos"
                console.log('🕵️ Binary not found in standard paths. Crawling filesystem...');
                const allFiles: string[] = [];
                async function crawl(dir: string, depth = 0) {
                    if (depth > 3) return; // Don't go too deep
                    try {
                        const entries = await fs.readdir(dir, { withFileTypes: true });
                        for (const entry of entries) {
                            const fullPath = path.resolve(dir, entry.name);
                            if (entry.isDirectory()) {
                                if (!['node_modules', '.next', '.git'].includes(entry.name)) {
                                    await crawl(fullPath, depth + 1);
                                }
                            } else {
                                allFiles.push(fullPath);
                                if (entry.name.toLowerCase().includes('tharos')) {
                                    console.log(`🎯 Found potential binary match: ${fullPath}`);
                                    // Try to use it if it's the right one
                                    if (entry.name === binaryName || entry.name === 'tharos-linux-amd64') {
                                        foundPath = fullPath;
                                    }
                                }
                            }
                        }
                    } catch (e) { /* ignore */ }
                }
                await crawl(process.cwd());

                if (!foundPath) {
                    tharosPath = searchPaths[0]; // Default for error
                    const debugInfo = {
                        error: 'Analysis binary not found',
                        cwd: process.cwd(),
                        scanned_files: allFiles.slice(0, 50), // Show first 50 files for debugging
                        searched_paths: searchPaths
                    };
                    return NextResponse.json(debugInfo, { status: 500 });
                } else {
                    tharosPath = foundPath;
                }
            } else {
                tharosPath = foundPath;
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
