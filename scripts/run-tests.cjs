const { spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const isWindows = process.platform === 'win32';
const binaryName = isWindows ? 'tharos.exe' : 'tharos';
const binaryPath = path.resolve(__dirname, '../dist', binaryName);
const vulnerableDir = path.resolve(__dirname, '../audit_samples/vulnerable');
const safeDir = path.resolve(__dirname, '../audit_samples/safe');



console.log('🚀 Starting Tharos Security Test Suite...');

// 1. Build the binary
console.log('🔨 Building Tharos binary...');
const buildRes = spawnSync('go', ['build', '-o', '../dist/tharos.exe', '.'], {
    cwd: path.resolve(__dirname, '../go-core'),
    stdio: 'inherit'
});

if (buildRes.status !== 0) {
    console.error('❌ Build failed!');
    process.exit(1);
}

let passed = 0;
let failed = 0;

function runTest(testPath, expectedExitCode) {
    console.log(`\n🔍 Testing: ${path.relative(process.cwd(), testPath)}`);
    const res = spawnSync(binaryPath, ['analyze', testPath, '--verbose'], {
        encoding: 'utf8'
    });

    if (res.status === expectedExitCode) {
        console.log(`✅ Passed (Exit Code: ${res.status})`);
        passed++;
    } else {
        console.log(`❌ Failed! Expected Exit Code: ${expectedExitCode}, Got: ${res.status}`);
        console.log(res.stdout);
        console.log(res.stderr);
        failed++;
    }
}

// Run Vulnerable Tests
const vulnerableFiles = fs.readdirSync(vulnerableDir);
vulnerableFiles.forEach(file => {
    runTest(path.join(vulnerableDir, file), 1);
});

// Run Safe Tests
const safeFiles = fs.readdirSync(safeDir);
safeFiles.forEach(file => {
    runTest(path.join(safeDir, file), 0);
});

console.log('\n--- Test Summary ---');
console.log(`Total: ${passed + failed}`);
console.log(`Passed: ${passed}`);
console.log(`Failed: ${failed}`);

if (failed > 0) {
    process.exit(1);
} else {
    process.exit(0);
}
