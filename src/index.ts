#!/usr/bin/env node
import { spawn } from 'child_process';
import path from 'path';
import { fileURLToPath } from 'url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));

// Path to the Go binary
const binaryName = process.platform === 'win32' ? 'tharos.exe' : 'tharos';
const binaryPath = path.resolve(__dirname, binaryName);

// Pass all arguments to the Go binary
const args = process.argv.slice(2);

const child = spawn(binaryPath, args, {
    stdio: 'inherit',
    shell: false
});

child.on('exit', (code) => {
    process.exit(code || 0);
});

child.on('error', (err) => {
    console.error(`Failed to start Tharos binary: ${err.message}`);
    process.exit(1);
});
