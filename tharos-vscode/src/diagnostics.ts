import * as vscode from 'vscode';

export function updateDiagnostics(
    document: vscode.TextDocument,
    collection: vscode.DiagnosticCollection,
    tharosOutput: any
) {
    collection.clear();

    const result = tharosOutput.results.find((r: any) =>
        // Normalize paths for comparison
        // tharos-security-ignore
        r.file.replace(/\\/g, '/') === document.fileName.replace(/\\/g, '/')
        || document.fileName.endsWith(r.file)
    );

    if (!result || !result.findings) {
        return;
    }

    const diagnostics: vscode.Diagnostic[] = [];

    for (const finding of result.findings) {
        // Line number is 1-based in Tharos, 0-based in VS Code
        const lineIndex = finding.line - 1;
        const line = document.lineAt(lineIndex);

        // Create range (attempt to highlight specific word/token if possible)
        let range = line.range;
        if (finding.byte_offset !== undefined && finding.byte_length !== undefined) {
            try {
                // VS Code uses 0-based character offsets. 
                // Byte offsets are a bit tricky with UTF-8 but for most code it's 1:1.
                // A better way is to translate byte offset to character offset.
                const startPos = document.positionAt(finding.byte_offset);
                const endPos = document.positionAt(finding.byte_offset + finding.byte_length);
                range = new vscode.Range(startPos, endPos);
            } catch (e) {
                console.error('Failed to calculate precise range:', e);
            }
        }

        const diagnostic = new vscode.Diagnostic(
            range,
            `${finding.message} [${finding.rule}]`,
            getSeverity(finding.severity)
        );

        diagnostic.source = 'Tharos';
        diagnostic.code = finding.rule;

        // Use full finding as metadata (as much as possible)
        (diagnostic as any).tharosFinding = finding;

        diagnostic.relatedInformation = [
            new vscode.DiagnosticRelatedInformation(
                new vscode.Location(document.uri, range),
                finding.explain || "No explanation provided."
            )
        ];

        diagnostics.push(diagnostic);
    }

    collection.set(document.uri, diagnostics);
}

function getSeverity(severity: string): vscode.DiagnosticSeverity {
    switch (severity.toLowerCase()) {
        case 'critical':
        case 'block':
            return vscode.DiagnosticSeverity.Error;
        case 'high':
            return vscode.DiagnosticSeverity.Error;
        case 'medium':
        case 'warning':
            return vscode.DiagnosticSeverity.Warning;
        case 'low':
        case 'info':
        default:
            return vscode.DiagnosticSeverity.Information;
    }
}
