import * as vscode from 'vscode';
import { TharosDiagnostics } from './diagnostics';

export class TharosHoverProvider implements vscode.HoverProvider {
    constructor(private diagnostics: TharosDiagnostics) { }

    provideHover(
        document: vscode.TextDocument,
        position: vscode.Position,
        token: vscode.CancellationToken
    ): vscode.ProviderResult<vscode.Hover> {
        const result = this.diagnostics.getResult(document.uri.toString());
        if (!result || result.ai_insights.length === 0) {
            return undefined;
        }

        const insight = result.ai_insights[0];

        // Create rich markdown content
        const markdown = new vscode.MarkdownString();
        markdown.isTrusted = true;
        markdown.supportHtml = true;

        // Header
        markdown.appendMarkdown('## 🦊 Tharos AI Insight\n\n');

        // Risk score with color coding
        const riskColor = this.getRiskColor(insight.risk_score);
        markdown.appendMarkdown(
            `**Risk Score:** \`${insight.risk_score}/100\` ${riskColor}\n\n`
        );

        // Recommendation
        markdown.appendMarkdown('### 💡 Recommendation\n\n');
        markdown.appendMarkdown(insight.recommendation + '\n\n');

        // Suggested fix if available
        if (insight.suggested_fix) {
            markdown.appendMarkdown('### 🔧 Suggested Fix\n\n');
            markdown.appendCodeblock(insight.suggested_fix, 'typescript');
        }

        // Add link to documentation
        markdown.appendMarkdown('\n\n---\n\n');
        markdown.appendMarkdown('[Learn more about Tharos](https://tharos.dev)');

        return new vscode.Hover(markdown);
    }

    private getRiskColor(score: number): string {
        if (score >= 80) {
            return '🔴 **Critical**';
        } else if (score >= 60) {
            return '🟠 **High**';
        } else if (score >= 40) {
            return '🟡 **Medium**';
        } else {
            return '🟢 **Low**';
        }
    }
}
