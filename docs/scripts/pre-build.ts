import { generateDocs } from './generate-docs.js'

async function main() {
  // Build Tharos binary
  const { fork } = await import('child_process');
  const path = await import('path');

  const installScript = path.resolve(process.cwd(), 'scripts/install-tharos.ts');

  console.log('🚀 Running Tharos installation script...');
  const { execSync } = await import('child_process');
  execSync(`bun ${installScript}`, { stdio: 'inherit' });

  // comment the below to disable openapi generation
  // await Promise.all([generateDocs()])
}

await main().catch((e) => {
  console.error('Failed to run pre build script', e)
  process.exit(1)
})
