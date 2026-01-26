#!/usr/bin/env node
import { Command } from 'commander';
import chalk from 'chalk';
import path from 'path';
import { fileURLToPath } from 'url';
import { initHooks, verifyHooks } from './hooks/manager.js';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const program = new Command();

program
    .name('fennec')
    .description('Intelligent, Unbreakable Code Policy Enforcement')
    .version('0.1.0');

program
    .command('init')
    .description('Initialize Fennec hooks in the current repository')
    .action(async () => {
        console.log(chalk.cyan('🦊 Initializing Fennec...'));
        try {
            await initHooks();
            console.log(chalk.green('✅ Fennec hooks installed successfully!'));
        } catch (error) {
            console.error(chalk.red('❌ Failed to initialize Fennec:'), error);
            process.exit(1);
        }
    });

program
    .command('check')
    .description('Run Fennec policy checks on staged files')
    .option('--self-heal', 'Perform self-healing if hooks are missing or tampered')
    .action(async (options) => {
        if (options.selfHeal) {
            await verifyHooks();
        }

        console.log(chalk.cyan('🦊 Fennec is analyzing your intent...'));

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
                    const corePath = path.resolve(__dirname, 'fennec-core.exe');
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
                        console.log(chalk.blue.italic('\n  🧠 Fennec AI Semantic Insights:'));
                        result.ai_insights.forEach((insight: string) => {
                            console.log(`     ✨ ${insight}`);
                        });
                    }
                } catch (e) {
                    console.error(chalk.red(`  ❌ Failed to analyze ${file}:`), e);
                }
            }

            if (globalBlock) {
                console.log(chalk.red('\n🛑 Commit blocked by Fennec policy. Please fix the issues above.'));
                process.exit(1);
            } else {
                console.log(chalk.green('\n✨ Fennec logic check passed! Proceeding...'));
            }
        } catch (error) {
            console.error(chalk.red('❌ Fennec check execution failed:'), error);
            process.exit(1);
        }
    });

program.parse();
