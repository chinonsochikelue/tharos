const { spawnSync } = require('child_process');
const path = require('path');
const fs = require('fs');

const isWindows = process.platform === 'win32';
const binaryName = isWindows ? 'tharos.exe' : 'tharos';
const binaryPath = path.resolve(__dirname, '../dist', binaryName);

// Ensure binary exists
if (!fs.existsSync(binaryPath)) {
    console.log('🔨 Building Tharos binary...');
    const buildRes = spawnSync('go', ['build', '-o', `../dist/${binaryName}`, '.'], {
        cwd: path.resolve(__dirname, '../go-core'),
        stdio: 'inherit'
    });
    if (buildRes.status !== 0) {
        console.error('❌ Build failed!');
        process.exit(1);
    }
}

// Create Test Cases
const testDir = path.resolve(__dirname, '../temp_verification');
if (!fs.existsSync(testDir)) fs.mkdirSync(testDir);

// 1. Critical File (Should Exit 1 normally)
const criticalFile = path.join(testDir, 'critical.js');
fs.writeFileSync(criticalFile, 'const apiKey = "sk_live_123456"; // hardcoded secret');

// 2. Safe File (Should Exit 0 normally)
const safeFile = path.join(testDir, 'safe.js');
fs.writeFileSync(safeFile, 'const apiKey = process.env.API_KEY;');

// 3. Medium File (Should Exit 0 normally, Exit 1 in strict mode)
// We need a rule that triggers medium. Let's use console.log if configured? 
// Or better: use tharos.yaml to define a custom medium rule.
const mediumFile = path.join(testDir, 'medium.js');
fs.writeFileSync(mediumFile, 'console.log("debug info");');

const tharosConfig = path.join(testDir, 'tharos.yaml');
fs.writeFileSync(tharosConfig, `
security:
  rules:
    - pattern: "console\\\\.log"
      message: "No console logs"
      severity: "medium"
`);

function runTharos(args, expectedCode, description) {
    console.log(`\n🧪 Testing: ${description}`);
    const res = spawnSync(binaryPath, args, {
        cwd: testDir, // run in testDir so it picks up tharos.yaml
        encoding: 'utf8'
    });

    if (res.status === expectedCode) {
        console.log(`✅ Passed (Got Exit Code ${res.status})`);
    } else {
        console.log(`❌ Failed! Expected ${expectedCode}, Got ${res.status}`);
        console.log(res.stdout);
        process.exit(1);
    }
}

console.log('🚀 Starting Verification...');

// Test 1: Critical File -> Exit 1
runTharos(['analyze', 'critical.js'], 1, 'Critical Issue (Standard Mode)');

// Test 2: Safe File -> Exit 0
runTharos(['analyze', 'safe.js'], 0, 'Safe File (Standard Mode)');

// Test 3: Medium File -> Exit 0 (Standard)
runTharos(['analyze', 'medium.js'], 0, 'Medium Issue (Standard Mode)');

// Test 4: Medium File + Strict -> Exit 1
runTharos(['analyze', 'medium.js', '--strict'], 1, 'Medium Issue (Strict Mode)');

console.log('\n✨ All verification tests passed!');
// Cleanup
fs.rmSync(testDir, { recursive: true, force: true });
