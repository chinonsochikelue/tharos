import * as vscode from 'vscode';

export class TharosFixer implements vscode.CodeActionProvider {
    public static readonly providedCodeActionKinds = [
        vscode.CodeActionKind.QuickFix
    ];

    provideCodeActions(
        document: vscode.TextDocument,
        range: vscode.Range | vscode.Selection,
        context: vscode.CodeActionContext,
        token: vscode.CancellationToken
    ): vscode.CodeAction[] {
        // for each diagnostic, check if it's a tharos diagnostic and has a finding
        return context.diagnostics
            .filter(diagnostic => diagnostic.source === 'Tharos' && (diagnostic as any).tharosFinding)
            .map(diagnostic => this.createFix(document, diagnostic));
    }

    private createFix(document: vscode.TextDocument, diagnostic: vscode.Diagnostic): vscode.CodeAction {
        const finding = (diagnostic as any).tharosFinding;
        const fix = new vscode.CodeAction(`✨ Tharos: Apply Magic Fix`, vscode.CodeActionKind.QuickFix);

        if (finding.replacement) {
            fix.edit = new vscode.WorkspaceEdit();

            // If we have byte offsets, we could be more precise, but for now 
            // we use the diagnostic range which is the whole line.
            // TODO: Improve range accuracy in diagnostics.ts
            fix.edit.replace(document.uri, diagnostic.range, finding.replacement);
        } else {
            fix.title = `🔍 Tharos: Explain issue`;
            fix.command = {
                command: 'tharos.explain',
                title: 'Explain issue',
                arguments: [finding]
            };
        }

        fix.diagnostics = [diagnostic];
        fix.isPreferred = true;
        return fix;
    }
}
