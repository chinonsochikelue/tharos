import fs from 'fs';
import path from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

const HOOK_CONTENT = `#!/bin/sh
# Tharos Git Hook
# This hook is managed by Tharos. Do not modify manually.
# VERSION: 0.1.2

# Self-healing check
if ! command -v tharos > /dev/null 2>&1; then
  echo "🦊 Tharos CLI not found. Skipping checks..."
  exit 0
fi

# Periodic setup audit & policy sync (non-blocking)
tharos sync > /dev/null 2>&1 &

# Run pre-commit security check
tharos check
`;



export async function initHooks() {
    const gitDir = await findGitDir();
    if (!gitDir) {
        throw new Error('Not a git repository');
    }

    const hooksDir = path.join(gitDir, 'hooks');
    if (!fs.existsSync(hooksDir)) {
        fs.mkdirSync(hooksDir, { recursive: true });
    }

    const preCommitHook = path.join(hooksDir, 'pre-commit');

    // Write the hook file
    fs.writeFileSync(preCommitHook, HOOK_CONTENT, { mode: 0o755 });

    if (process.platform !== 'win32') {
        // Ensure it's executable on non-windows
        fs.chmodSync(preCommitHook, '755');
    }
}

export async function verifyHooks() {
    const gitDir = await findGitDir();
    if (!gitDir) return;

    const preCommitHook = path.join(gitDir, 'hooks', 'pre-commit');
    if (!fs.existsSync(preCommitHook)) {
        console.log('⚠️ Tharos hook missing. Re-installing...');
        await initHooks();
        return;
    }

    const content = fs.readFileSync(preCommitHook, 'utf-8');
    if (!content.includes('managed by Tharos')) {
        console.log('⚠️ Tharos hook tampered with. Repairing...');
        await initHooks();
    }
}

async function findGitDir(): Promise<string | null> {
    try {
        const { stdout } = await execAsync('git rev-parse --git-dir');
        return path.resolve(stdout.trim());
    } catch {
        return null;
    }
}
