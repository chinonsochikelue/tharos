import * as vscode from 'vscode';
import * as path from 'path';
import { execFile } from 'child_process';
import { updateDiagnostics } from './diagnostics';
import { TharosFixer } from './fixer';

const DIAGNOSTIC_COLLECTION_NAME = 'tharos';

export function activate(context: vscode.ExtensionContext) {
    console.log('Tharos Security Extension is now active!');

    const diagnosticCollection = vscode.languages.createDiagnosticCollection(DIAGNOSTIC_COLLECTION_NAME);
    context.subscriptions.push(diagnosticCollection);

    // Register Quick Fix Provider
    context.subscriptions.push(
        vscode.languages.registerCodeActionsProvider(
            { pattern: '**/*' },
            new TharosFixer(),
            {
                providedCodeActionKinds: TharosFixer.providedCodeActionKinds
            }
        )
    );

    // Command: Scan Current File
    context.subscriptions.push(
        vscode.commands.registerCommand('tharos.scanFile', () => {
            const editor = vscode.window.activeTextEditor;
            if (editor) {
                scanFile(editor.document, diagnosticCollection, context);
            }
        })
    );

    // Listen for Save and Open Events
    context.subscriptions.push(
        vscode.workspace.onDidSaveTextDocument((document) => {
            scanFile(document, diagnosticCollection, context);
        })
    );

    context.subscriptions.push(
        vscode.workspace.onDidOpenTextDocument((document) => {
            scanFile(document, diagnosticCollection, context);
        })
    );

    // Initial Scan of Active File
    if (vscode.window.activeTextEditor) {
        scanFile(vscode.window.activeTextEditor.document, diagnosticCollection, context);
    }
}

function scanFile(
    document: vscode.TextDocument,
    collection: vscode.DiagnosticCollection,
    context: vscode.ExtensionContext
) {
    // Only scan relevant files
    const supportedLangs = ['javascript', 'typescript', 'javascriptreact', 'typescriptreact', 'go', 'python'];
    if (!supportedLangs.includes(document.languageId)) {
        return;
    }

    const config = vscode.workspace.getConfiguration('tharos');
    const strictMode = config.get<boolean>('strictMode', false);

    // Resolve Binary path (Bundled vs System)
    // For this POC, we look in the 'bin' folder of the extension
    let binaryPath = path.join(context.extensionPath, 'bin', process.platform === 'win32' ? 'tharos-win32-x64.exe' : 'tharos');

    // Allow override
    const configPath = config.get<string>('binaryPath');
    if (configPath) {
        binaryPath = configPath;
    }

    const args = ['analyze', document.fileName, '--format', 'json'];
    if (strictMode) {
        args.push('--strict');
    }

    console.log(`Running Tharos: ${binaryPath} ${args.join(' ')}`);

    execFile(binaryPath, args, (error, stdout, stderr) => {
        if (stderr) {
            console.error(`Tharos Stderr: ${stderr}`);
        }

        // Parse JSON output
        try {
            const output = JSON.parse(stdout);
            updateDiagnostics(document, collection, output);
        } catch (e) {
            console.error('Failed to parse Tharos output:', e);
            console.log('Raw output:', stdout);
        }
    });
}

export function deactivate() { }
