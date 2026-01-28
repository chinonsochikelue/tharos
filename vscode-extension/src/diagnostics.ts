import * as vscode from 'vscode';
import * as path from 'path';
import { exec } from 'child_process';
import { promisify } from 'util';

const execAsync = promisify(exec);

interface TharosFinding {
    type: string;
    message: string;
    severity: 'block' | 'warning' | 'info';
    line: number;
}

interface AIInsight {
    recommendation: string;
    risk_score: number;
    suggested_fix?: string;
}

interface TharosResult {
    file: string;
    findings: TharosFinding[];
    ai_insights: AIInsight[];
}

export class TharosDiagnostics {
    private diagnosticCollection: vscode.DiagnosticCollection;
    private statusBarItem: vscode.StatusBarItem;
    private resultsCache: Map<string, TharosResult> = new Map();

    constructor(
        diagnosticCollection: vscode.DiagnosticCollection,
        statusBarItem: vscode.StatusBarItem
    ) {
        this.diagnosticCollection = diagnosticCollection;
        this.statusBarItem = statusBarItem;
    }

    async analyze(document: vscode.TextDocument): Promise<void> {
        if (document.uri.scheme !== 'file') {
            console.log('[Tharos] Skipping non-file document:', document.uri.toString());
            return;
        }

        console.log('[Tharos] Starting analysis for:', document.uri.fsPath);
        this.statusBarItem.text = '$(sync~spin) Analyzing...';

        try {
            const corePath = this.getCorePath();
            const filePath = document.uri.fsPath;

            console.log('[Tharos] Using core path:', corePath);
            console.log('[Tharos] Analyzing file:', filePath);

            // Call tharos-core
            const { stdout, stderr } = await execAsync(`"${corePath}" analyze "${filePath}"`);

            if (stderr) {
                console.log('[Tharos] Core stderr:', stderr);
            }

            console.log('[Tharos] Core stdout:', stdout);

            const result: TharosResult = JSON.parse(stdout);
            console.log('[Tharos] Parsed result:', result);

            // Cache results for hover provider
            this.resultsCache.set(document.uri.toString(), result);

            // Convert to VSCode diagnostics
            const diagnostics = this.convertToDiagnostics(result, document);
            this.diagnosticCollection.set(document.uri, diagnostics);

            console.log('[Tharos] Created diagnostics:', diagnostics.length);

            // Update status bar
            const errorCount = diagnostics.filter(d => d.severity === vscode.DiagnosticSeverity.Error).length;
            const warningCount = diagnostics.filter(d => d.severity === vscode.DiagnosticSeverity.Warning).length;

            if (errorCount > 0) {
                this.statusBarItem.text = `$(error) ${errorCount} $(warning) ${warningCount}`;
            } else if (warningCount > 0) {
                this.statusBarItem.text = `$(warning) ${warningCount}`;
            } else {
                this.statusBarItem.text = '$(shield) Tharos ✓';
            }
        } catch (error) {
            console.error('[Tharos] Analysis error:', error);
            this.statusBarItem.text = '$(shield) Tharos ⚠';

            // Show error to user
            vscode.window.showErrorMessage(`Tharos analysis failed: ${error}`);
        }
    }

    private convertToDiagnostics(result: TharosResult, document: vscode.TextDocument): vscode.Diagnostic[] {
        const diagnostics: vscode.Diagnostic[] = [];

        // Add findings
        for (const finding of result.findings) {
            const line = Math.max(0, finding.line - 1); // Convert to 0-indexed
            const range = document.lineAt(line).range;

            const severity = this.mapSeverity(finding.severity);
            const diagnostic = new vscode.Diagnostic(
                range,
                finding.message,
                severity
            );

            diagnostic.source = 'Tharos';
            diagnostic.code = `tharos-${finding.type}`;

            diagnostics.push(diagnostic);
        }

        // Add AI insights as informational diagnostics
        if (result.ai_insights.length > 0) {
            const insight = result.ai_insights[0];
            const range = new vscode.Range(0, 0, 0, 0);

            const message = `🧠 AI Insight (Risk: ${insight.risk_score}/100): ${insight.recommendation}`;
            const diagnostic = new vscode.Diagnostic(
                range,
                message,
                vscode.DiagnosticSeverity.Information
            );

            diagnostic.source = 'Tharos AI';
            diagnostic.code = 'tharos-ai-insight';

            diagnostics.push(diagnostic);
        }

        return diagnostics;
    }

    private mapSeverity(severity: string): vscode.DiagnosticSeverity {
        switch (severity) {
            case 'block':
                return vscode.DiagnosticSeverity.Error;
            case 'warning':
                return vscode.DiagnosticSeverity.Warning;
            case 'info':
                return vscode.DiagnosticSeverity.Information;
            default:
                return vscode.DiagnosticSeverity.Warning;
        }
    }

    private getCorePath(): string {
        const config = vscode.workspace.getConfiguration('tharos');
        const customPath = config.get<string>('corePath');

        if (customPath) {
            return customPath;
        }

        // Auto-detect core path
        const workspaceRoot = vscode.workspace.workspaceFolders?.[0]?.uri.fsPath;
        if (workspaceRoot) {
            const localCore = path.join(workspaceRoot, 'dist', 'tharos-core.exe');
            return localCore;
        }

        return 'tharos-core.exe'; // Assume it's in PATH
    }

    getResult(uri: string): TharosResult | undefined {
        return this.resultsCache.get(uri);
    }
}
