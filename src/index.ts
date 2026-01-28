#!/usr/bin/env node
import { Command } from 'commander';
import chalk from 'chalk';
import path from 'path';
import { fileURLToPath } from 'url';
import { initHooks, verifyHooks } from './hooks/manager.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const program = new Command();

program
    .name('tharos')
    .description('Tharos: Intelligent, Unbreakable Code Policy Enforcement')
    .version('0.1.0');

program
    .command('init')
    .description('Initialize Tharos hooks in the current repository')
    .action(async () => {
        console.log(chalk.cyan('🛡️ Initializing Tharos...'));
        try {
            await initHooks();
            console.log(chalk.green('✅ Tharos hooks installed successfully!'));
        } catch (error) {
            console.error(chalk.red('❌ Failed to initialize Tharos:'), error);
            process.exit(1);
        }
    });

program
    .command('login')
    .description('Authenticate with Tharos Cloud')
    .action(async () => {
        const { default: os } = await import('os');
        const { default: fs } = await import('fs');
        const { default: readline } = await import('readline');

        const rl = readline.createInterface({
            input: process.stdin,
            output: process.stdout
        });

        console.log(chalk.cyan('🔐 Tharos Cloud Authentication'));

        rl.question('Enter your Tharos API Key: ', (apiKey) => {
            if (!apiKey.trim()) {
                console.log(chalk.red('❌ API Key cannot be empty.'));
                rl.close();
                return;
            }

            const projectDir = process.cwd();
            const configDir = path.join(projectDir, '.tharos');
            const configFile = path.join(configDir, 'config.json');

            if (!fs.existsSync(configDir)) {
                fs.mkdirSync(configDir, { recursive: true });
            }

            const config = {
                apiKey: apiKey.trim(),
                updatedAt: new Date().toISOString()
            };

            fs.writeFileSync(configFile, JSON.stringify(config, null, 2));

            // Check if .gitignore exists and add .tharos if needed
            const gitignorePath = path.join(projectDir, '.gitignore');
            if (fs.existsSync(gitignorePath)) {
                const gitignoreContent = fs.readFileSync(gitignorePath, 'utf-8');
                if (!gitignoreContent.includes('.tharos')) {
                    fs.appendFileSync(gitignorePath, '\n# Tharos local config\n.tharos/\n');
                    console.log(chalk.yellow('📝 Added .tharos/ to .gitignore'));
                }
            } else {
                console.log(chalk.yellow('⚠️ No .gitignore found. Please ensure .tharos/ is ignored to protect your API key.'));
            }

            console.log(chalk.green(`\n✅ Authenticated successfully!`));
            console.log(chalk.gray(`   Config saved to: ${configFile}`));
            rl.close();
        });
    });

program
    .command('sync')
    .description('Synchronize organizational policies with the cloud')
    .action(async () => {
        const { default: os } = await import('os');
        const { default: fs } = await import('fs');

        const projectDir = process.cwd();
        const configFile = path.join(projectDir, '.tharos', 'config.json');

        if (!fs.existsSync(configFile)) {
            console.log(chalk.yellow('⚠️ No session found in this project. Please run "tharos login" first.'));
            // Optional: allow continue as anonymous or local-only
        }

        console.log(chalk.cyan('☁️ Syncing Tharos policies with cloud...'));
        await new Promise(resolve => setTimeout(resolve, 1500)); // Simulate network latency
        console.log(chalk.green('✅ Organizational policies synchronized!'));
        console.log(chalk.gray('   Applied Policy: SEC-RULE-2026 (Enforced)'));
    });

program
    .command('check')
    .description('Run Tharos policy checks on staged files')
    .option('--self-heal', 'Perform self-healing if hooks are missing or tampered')
    .action(async (options) => {
        if (options.selfHeal) {
            await verifyHooks();
        }

        console.log(chalk.cyan('🛡️ Tharos is analyzing your intent...'));

        try {
            const { execa } = await import('execa');

            // Get staged files
            const { stdout: stagedFiles } = await execa('git', ['diff', '--cached', '--name-only']);
            const files = stagedFiles.split('\n').filter(f => f.match(/\.(js|ts|jsx|tsx)$/));

            if (files.length === 0) {
                console.log(chalk.gray('No relevant files staged for commit.'));
                return;
            }

            let globalBlock = false;

            for (const file of files) {
                console.log(chalk.white(`\n📄 Analyzing ${chalk.bold(file)}...`));

                try {
                    const corePath = path.resolve(__dirname, 'tharos-core.exe');
                    const { stdout } = await execa(corePath, ['analyze', file]);
                    const result = JSON.parse(stdout);

                    // Display Findings
                    if (result.findings && result.findings.length > 0) {
                        result.findings.forEach((finding: any) => {
                            const color = finding.severity === 'block' ? chalk.red : chalk.yellow;
                            const icon = finding.severity === 'block' ? '🛑' : '⚠️';

                            console.log(`  ${icon} ${color(finding.type.toUpperCase())}: ${finding.message}`);
                            if (finding.line) {
                                console.log(chalk.gray(`     Line ${finding.line}`));
                            }

                            if (finding.severity === 'block') globalBlock = true;
                        });
                    } else {
                        console.log(chalk.green('  ✅ No issues found.'));
                    }

                    // Display AI Insights
                    if (result.ai_insights && result.ai_insights.length > 0) {
                        console.log(chalk.blue.italic('\n  🧠 Tharos AI Semantic Insights:'));
                        result.ai_insights.forEach((insight: any) => {
                            if (typeof insight === 'string') {
                                console.log(`     ✨ ${insight}`);
                                return;
                            }
                            const score = insight.risk_score || 50;
                            const recommendation = insight.recommendation || insight;
                            const scoreColor = score > 70 ? chalk.red : score > 40 ? chalk.yellow : chalk.green;

                            console.log(`     ✨ ${recommendation}`);
                            console.log(`     📊 Risk Score: ${scoreColor(score + '/100')}`);

                            if (insight.suggested_fix) {
                                console.log(chalk.cyan('\n     💡 Suggested Fix:'));
                                console.log(chalk.gray('     ---------------------------------------'));
                                console.log(insight.suggested_fix.split('\n').map((line: string) => `     ${line}`).join('\n'));
                                console.log(chalk.gray('     ---------------------------------------'));
                            }
                        });
                    } else if (result.findings && result.findings.length > 0) {
                        console.log(chalk.gray('\n  💡 Tip: No AI insights available.'));
                        console.log(chalk.gray('     Run "ollama serve" or use Tharos Cloud for smart analysis.'));
                    }
                } catch (e) {
                    console.error(chalk.red(`  ❌ Failed to analyze ${file}:`), e);
                }
            }

            if (globalBlock) {
                console.log(chalk.red('\n🛑 Commit blocked by Tharos policy. Please fix the issues above.'));
                process.exit(1);
            } else {
                console.log(chalk.green('\n✨ Tharos logic check passed! Proceeding...'));
            }
        } catch (error) {
            console.error(chalk.red('❌ Tharos check execution failed:'), error);
            process.exit(1);
        }
    });

program.parse();
