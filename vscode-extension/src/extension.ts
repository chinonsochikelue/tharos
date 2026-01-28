import * as vscode from 'vscode';
import { TharosDiagnostics } from './diagnostics';
import { TharosCodeActionProvider } from './codeActions';
import { TharosHoverProvider } from './hover';

let diagnostics: TharosDiagnostics;
let statusBarItem: vscode.StatusBarItem;

export function activate(context: vscode.ExtensionContext) {
    console.log('🛡️ Tharos extension is now active!');

    // Create status bar item
    statusBarItem = vscode.window.createStatusBarItem(
        vscode.StatusBarAlignment.Right,
        100
    );
    statusBarItem.text = '$(shield) Tharos';
    statusBarItem.tooltip = 'Tharos Security & Quality Analysis';
    statusBarItem.show();
    context.subscriptions.push(statusBarItem);

    // Initialize diagnostic provider
    const diagnosticCollection = vscode.languages.createDiagnosticCollection('tharos');
    diagnostics = new TharosDiagnostics(diagnosticCollection, statusBarItem);
    context.subscriptions.push(diagnosticCollection);

    // Register providers for all supported languages
    const supportedLanguages = [
        'typescript',
        'javascript',
        'typescriptreact',
        'javascriptreact',
        'python',
        'go',
        'rust',
        'java'
    ];

    // Code Actions Provider (Quick Fixes)
    const codeActionProvider = new TharosCodeActionProvider();
    for (const lang of supportedLanguages) {
        context.subscriptions.push(
            vscode.languages.registerCodeActionsProvider(
                lang,
                codeActionProvider,
                {
                    providedCodeActionKinds: TharosCodeActionProvider.providedCodeActionKinds
                }
            )
        );
    }

    // Hover Provider (Show AI insights on hover)
    const hoverProvider = new TharosHoverProvider(diagnostics);
    for (const lang of supportedLanguages) {
        context.subscriptions.push(
            vscode.languages.registerHoverProvider(lang, hoverProvider)
        );
    }

    // Analyze on save
    context.subscriptions.push(
        vscode.workspace.onDidSaveTextDocument((document) => {
            if (supportedLanguages.includes(document.languageId)) {
                diagnostics.analyze(document);
            }
        })
    );

    // Analyze on open
    context.subscriptions.push(
        vscode.workspace.onDidOpenTextDocument((document) => {
            if (supportedLanguages.includes(document.languageId)) {
                diagnostics.analyze(document);
            }
        })
    );

    // Analyze current file on activation
    if (vscode.window.activeTextEditor) {
        const document = vscode.window.activeTextEditor.document;
        if (supportedLanguages.includes(document.languageId)) {
            diagnostics.analyze(document);
        }
    }

    // Register commands
    context.subscriptions.push(
        vscode.commands.registerCommand('tharos.analyzeFile', () => {
            const editor = vscode.window.activeTextEditor;
            if (editor) {
                diagnostics.analyze(editor.document);
            }
        })
    );

    context.subscriptions.push(
        vscode.commands.registerCommand('tharos.analyzeWorkspace', async () => {
            statusBarItem.text = '$(sync~spin) Analyzing...';
            const files = await vscode.workspace.findFiles('**/*.{ts,js,py,go,rs,java}');
            for (const file of files) {
                const document = await vscode.workspace.openTextDocument(file);
                await diagnostics.analyze(document);
            }
            statusBarItem.text = '$(shield) Tharos';
        })
    );
}

export function deactivate() {
    console.log('🛡️ Tharos extension deactivated');
}
