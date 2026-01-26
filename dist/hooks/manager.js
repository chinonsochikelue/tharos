import fs from 'fs';
import path from 'path';
import { execa } from 'execa';
const HOOK_CONTENT = `#!/bin/sh
# Fennec Git Hook
// This hook is managed by Fennec. Do not modify manually.
// VERSION: 0.1.0

# Self-healing check
if ! command -v fennec > /dev/null 2>&1; then
  echo "🦊 Fennec CLI not found. Skipping checks..."
  exit 0
fi

fennec check --self-heal
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
    if (!gitDir)
        return;
    const preCommitHook = path.join(gitDir, 'hooks', 'pre-commit');
    if (!fs.existsSync(preCommitHook)) {
        console.log('⚠️ Fennec hook missing. Re-installing...');
        await initHooks();
        return;
    }
    const content = fs.readFileSync(preCommitHook, 'utf-8');
    if (!content.includes('managed by Fennec')) {
        console.log('⚠️ Fennec hook tampered with. Repairing...');
        await initHooks();
    }
}
async function findGitDir() {
    try {
        const { stdout } = await execa('git', ['rev-parse', '--git-dir']);
        return path.resolve(stdout.trim());
    }
    catch {
        return null;
    }
}
