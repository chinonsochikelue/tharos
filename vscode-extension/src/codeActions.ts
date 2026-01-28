import * as vscode from 'vscode';

export class TharosCodeActionProvider implements vscode.CodeActionProvider {
    public static readonly providedCodeActionKinds = [
        vscode.CodeActionKind.QuickFix
    ];

    provideCodeActions(
        document: vscode.TextDocument,
        range: vscode.Range | vscode.Selection,
        context: vscode.CodeActionContext,
        token: vscode.CancellationToken
    ): vscode.CodeAction[] {
        const codeActions: vscode.CodeAction[] = [];

        // Find Tharos diagnostics in the current range
        const tharosDiagnostics = context.diagnostics.filter(
            d => d.source === 'Tharos' || d.source === 'Tharos AI'
        );

        for (const diagnostic of tharosDiagnostics) {
            // Check if there's a suggested fix in the diagnostic
            if (diagnostic.source === 'Tharos AI') {
                const fix = this.createFixAction(document, diagnostic);
                if (fix) {
                    codeActions.push(fix);
                }
            }

            // Add "Ignore this warning" action
            const ignoreAction = this.createIgnoreAction(diagnostic);
            codeActions.push(ignoreAction);
        }

        return codeActions;
    }

    private createFixAction(
        document: vscode.TextDocument,
        diagnostic: vscode.Diagnostic
    ): vscode.CodeAction | undefined {
        // Extract suggested fix from diagnostic message
        // This is a simplified version - in production, we'd parse the actual AI insight
        const action = new vscode.CodeAction(
            '🦊 Apply Tharos AI Fix',
            vscode.CodeActionKind.QuickFix
        );

        action.diagnostics = [diagnostic];
        action.isPreferred = true;

        // TODO: Parse actual suggested_fix from AI insight and create WorkspaceEdit
        // For now, we'll show a placeholder
        action.command = {
            command: 'tharos.showFixPreview',
            title: 'Show Fix Preview',
            arguments: [document, diagnostic]
        };

        return action;
    }

    private createIgnoreAction(diagnostic: vscode.Diagnostic): vscode.CodeAction {
        const action = new vscode.CodeAction(
            'Ignore this Tharos warning',
            vscode.CodeActionKind.QuickFix
        );

        action.diagnostics = [diagnostic];

        // Add a comment to ignore this specific warning
        action.command = {
            command: 'tharos.ignoreWarning',
            title: 'Ignore Warning',
            arguments: [diagnostic]
        };

        return action;
    }
}
