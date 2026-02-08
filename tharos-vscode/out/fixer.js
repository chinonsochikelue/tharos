"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
exports.TharosFixer = void 0;
const vscode = require("vscode");
class TharosFixer {
    provideCodeActions(document, range, context, token) {
        // for each diagnostic, check if it's a tharos diagnostic and has a finding
        return context.diagnostics
            .filter(diagnostic => diagnostic.source === 'Tharos' && diagnostic.tharosFinding)
            .map(diagnostic => this.createFix(document, diagnostic));
    }
    createFix(document, diagnostic) {
        const finding = diagnostic.tharosFinding;
        const fix = new vscode.CodeAction(`✨ Tharos: Apply Magic Fix`, vscode.CodeActionKind.QuickFix);
        if (finding.replacement) {
            fix.edit = new vscode.WorkspaceEdit();
            // If we have byte offsets, we could be more precise, but for now 
            // we use the diagnostic range which is the whole line.
            // TODO: Improve range accuracy in diagnostics.ts
            fix.edit.replace(document.uri, diagnostic.range, finding.replacement);
        }
        else {
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
exports.TharosFixer = TharosFixer;
TharosFixer.providedCodeActionKinds = [
    vscode.CodeActionKind.QuickFix
];
//# sourceMappingURL=fixer.js.map