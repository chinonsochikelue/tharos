import { generateDocs } from './generate-docs.js'

async function main() {
  // Build Tharos binary
  const { fork } = await import('child_process');
  const path = await import('path');

  const installScript = path.resolve(process.cwd(), 'scripts/install-tharos.ts');

  // Run as a separate bun process
  const child = fork(installScript, [], { stdio: 'inherit', execPath: 'bun' });

  await new Promise((resolve, reject) => {
    child.on('exit', (code) => {
      if (code === 0) resolve(null);
      else reject(new Error(`install-tharos exited with code ${code}`));
    });
  });

  // comment the below to disable openapi generation
  // await Promise.all([generateDocs()])
}

await main().catch((e) => {
  console.error('Failed to run pre build script', e)
  process.exit(1)
})
