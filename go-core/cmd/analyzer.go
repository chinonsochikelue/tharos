package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/glamour"

	"github.com/google/generative-ai-go/genai"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
	"google.golang.org/api/option"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorYellow = "\033[33m"
	colorGreen  = "\033[32m"
	colorCyan   = "\033[36m"
	colorGray   = "\033[90m"
	colorBold   = "\033[1m"
)

type BatchResult struct {
	Results []AnalysisResult `json:"results"`
	Summary ScanSummary      `json:"summary"`
}

type ScanSummary struct {
	TotalFiles      int    `json:"total_files"`
	Vulnerabilities int    `json:"vulnerabilities"`
	Duration        string `json:"duration"`
	DurationMs      int64  `json:"duration_ms"`
}

type AnalysisResult struct {
	File       string      `json:"file"`
	Findings   []Finding   `json:"findings"`
	AIInsights []AIInsight `json:"ai_insights"`
}

type AIInsight struct {
	Recommendation string    `json:"recommendation"`
	RiskScore      int       `json:"risk_score"`
	SuggestedFix   string    `json:"suggested_fix,omitempty"`
	Fixes          []LineFix `json:"fixes,omitempty"`
}

type LineFix struct {
	Line        int    `json:"line"`
	Type        string `json:"type"`
	Replacement string `json:"replacement"`
}

type Finding struct {
	Rule        string  `json:"rule"`
	Type        string  `json:"type"`
	Message     string  `json:"message"`
	Severity    string  `json:"severity"` // "critical", "high", "medium", "info"
	Confidence  float64 `json:"confidence"`
	Explain     string  `json:"explain"`
	Remediation string  `json:"remediation"`
	Autofix     bool    `json:"autofix"`
	Line        int     `json:"line"`
	ByteOffset  int     `json:"byte_offset,omitempty"`
	ByteLength  int     `json:"byte_length,omitempty"`
	Replacement string  `json:"replacement,omitempty"`
}

var policySeverity string

func init() {
	// Initialize policy severity based on config file
	policySeverity = "warning"
	if _, err := os.Stat("tharos.yaml"); err == nil {
		policySeverity = "block"
	}
}

func setupCommand() {
	fmt.Printf("%s%s[THAROS AI SETUP]%s\n\n", colorBold, colorCyan, colorReset)

	fmt.Println("Tharos uses AI to provide deeper security insights and recommendations.")
	fmt.Println("Choose one of the following providers (both have free tiers):")
	fmt.Println()

	// Check Gemini
	fmt.Printf("%s1. Google Gemini (Recommended)%s\n", colorBold, colorReset)
	if geminiKey := os.Getenv("GEMINI_API_KEY"); geminiKey != "" {
		fmt.Printf("  %s✓ API Key Configured%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("  %s✗ Not configured%s\n", colorRed, colorReset)
		fmt.Printf("    1. Get key: %shttps://makersuite.google.com/app/apikey%s\n", colorCyan, colorReset)
		fmt.Printf("    2. Set env:  %sexport GEMINI_API_KEY=\"your-key\"%s\n", colorCyan, colorReset)
		fmt.Printf("       Windows:  %s$env:GEMINI_API_KEY=\"your-key\"%s\n", colorCyan, colorReset)
	}
	fmt.Println()

	// Check Groq
	fmt.Printf("%s2. Groq (Fast & Free)%s\n", colorBold, colorReset)
	if groqKey := os.Getenv("GROQ_API_KEY"); groqKey != "" {
		fmt.Printf("  %s✓ API Key Configured%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("  %s✗ Not configured%s\n", colorRed, colorReset)
		fmt.Printf("    1. Get key: %shttps://console.groq.com%s\n", colorCyan, colorReset)
		fmt.Printf("    2. Set env:  %sexport GROQ_API_KEY=\"your-key\"%s\n", colorCyan, colorReset)
		fmt.Printf("       Windows:  %s$env:GROQ_API_KEY=\"your-key\"%s\n", colorCyan, colorReset)
	}
	fmt.Println()

	fmt.Printf("%sNote:%s Tharos works great without AI! These provide enhanced insights.\n", colorBold, colorReset)
	fmt.Printf("Priority: Gemini → Groq\n")
}

func printJSONOutput(result BatchResult) {
	jsonOutput, _ := json.MarshalIndent(result, "", "  ")
	fmt.Println(string(jsonOutput))
}

func printSARIFOutput(results []AnalysisResult) {
	sarif := ConvertToSARIF(results)
	sarifJSON, _ := json.MarshalIndent(sarif, "", "  ")
	fmt.Println(string(sarifJSON))
}

func renderMarkdown(content string) (string, error) {
	// Use dark theme by default, or auto-detect
	r, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(80),
	)

	out, err := r.Render(content)
	if err != nil {
		return content, err
	}
	return strings.TrimSpace(out), nil
}

func indentMultiline(text, indent string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != "" {
			lines[i] = indent + line
		}
	}
	return strings.Join(lines, "\n")
}

func printRichOutput(result BatchResult, verboseMode bool, fixModeEnabled bool) {
	fmt.Printf("%s%s[THAROS SECURITY SCAN]%s\n", colorBold, colorCyan, colorReset)
	fmt.Println()

	if len(result.Results) == 0 {
		fmt.Printf("%s✓%s No files analyzed.\n", colorGreen, colorReset)
		return
	}

	filesWithIssues := 0
	for _, r := range result.Results {
		if len(r.Findings) > 0 {
			filesWithIssues++
		}
	}

	// Summary header
	fmt.Printf("📊 %sScanned:%s %d files in %s\n", colorBold, colorReset, result.Summary.TotalFiles, result.Summary.Duration)
	if result.Summary.Vulnerabilities > 0 {
		fmt.Printf("⚠️  %sIssues:%s %d vulnerabilities in %d files\n\n", colorBold, colorReset, result.Summary.Vulnerabilities, filesWithIssues)
	} else {
		fmt.Printf("%s✓ No issues found!%s\n", colorGreen, colorReset)
		return
	}

	// Detailed findings
	for _, r := range result.Results {
		if len(r.Findings) == 0 {
			continue
		}

		fmt.Printf("%s%s%s%s\n", colorBold, colorCyan, r.File, colorReset)
		for _, f := range r.Findings {
			severity := getSeveritySymbol(f.Severity)
			color := getSeverityColor(f.Severity)
			fmt.Printf("  %s%s Line %d:%s %s\n", color, severity, f.Line, colorReset, f.Message)

			if verboseMode && f.Replacement != "" {
				label := "💡 Fix available:"
				if strings.Contains(f.Message, "AI Insight") || f.Type == "ai_fix" { // Simple heuristic or we could add a field
					label = "🤖 AI Fix available:"
				}
				fmt.Printf("    %s%s%s %s\n", colorGreen, label, colorReset, f.Replacement)
			}
		}

		// AI Insights
		if len(r.AIInsights) > 0 {
			for _, ai := range r.AIInsights {
				fmt.Printf("  %s🤖 AI Recommendation:%s\n", colorCyan, colorReset)

				// Render recommendation as markdown
				renderedRec, _ := renderMarkdown(ai.Recommendation)
				fmt.Println(indentMultiline(renderedRec, "    "))

				if ai.SuggestedFix != "" && verboseMode {
					fmt.Printf("    %sSuggested fix:%s\n", colorGreen, colorReset)
					// Render code block or fix as markdown
					fixMd := fmt.Sprintf("```go\n%s\n```", ai.SuggestedFix)
					renderedFix, _ := renderMarkdown(fixMd)
					fmt.Println(indentMultiline(renderedFix, "    "))
				}
			}
		}
		fmt.Println()
	}

	if fixModeEnabled {
		// Just print a note that fixes were attempted if in rich mode
		fmt.Printf("%s⚙️  Auto-fixes were processed for this scan.%s\n", colorYellow, colorReset)
	}
}

func applyFixes(results []AnalysisResult) int {
	totalApplied := 0

	for _, res := range results {
		if len(res.Findings) == 0 {
			continue
		}

		// Collect findings that have replacements
		var fixable []Finding
		for _, f := range res.Findings {
			if f.Replacement != "" && f.ByteOffset > 0 && f.ByteLength > 0 {
				fixable = append(fixable, f)
			}
		}

		if len(fixable) == 0 {
			continue
		}

		// Sort findings by byte offset in reverse order (bottom to top)
		// to prevent offset shifts after each replacement
		sort.Slice(fixable, func(i, j int) bool {
			return fixable[i].ByteOffset > fixable[j].ByteOffset
		})

		content, err := ioutil.ReadFile(res.File)
		if err != nil {
			fmt.Printf("  %s❌ Failed to read %s for fixing: %v%s\n", colorRed, res.File, err, colorReset)
			continue
		}

		newContent := make([]byte, len(content))
		copy(newContent, content)

		appliedToFile := 0
		for _, f := range fixable {
			// Basic safety check: ensure the offset and length are within bounds
			if f.ByteOffset >= len(newContent) || f.ByteOffset+f.ByteLength > len(newContent) {
				continue
			}

			// Apply the replacement
			prefix := newContent[:f.ByteOffset]
			suffix := newContent[f.ByteOffset+f.ByteLength:]

			updated := make([]byte, 0, len(prefix)+len(f.Replacement)+len(suffix))
			updated = append(updated, prefix...)
			updated = append(updated, []byte(f.Replacement)...)
			updated = append(updated, suffix...)

			newContent = updated
			appliedToFile++
		}

		if appliedToFile > 0 {
			err = ioutil.WriteFile(res.File, newContent, 0o644)
			if err != nil {
				fmt.Printf("  %s❌ Failed to write fixes to %s: %v%s\n", colorRed, res.File, err, colorReset)
			} else {
				totalApplied += appliedToFile
			}
		}
	}

	return totalApplied
}

func getSeveritySymbol(severity string) string {
	switch severity {
	case "critical":
		return "💀"
	case "high":
		return "🔴"
	case "medium":
		return "⚠️ "
	case "info":
		return "ℹ️ "
	default:
		return "  "
	}
}

func getSeverityColor(severity string) string {
	switch severity {
	case "critical":
		return colorRed + colorBold
	case "high":
		return colorRed
	case "medium":
		return colorYellow
	default:
		return colorReset
	}
}

func calculateEntropy(s string) float64 {
	m := make(map[rune]int)
	for _, r := range s {
		m[r]++
	}
	var entropy float64
	for _, count := range m {
		p := float64(count) / float64(len(s))
		entropy -= p * (p / float64(0.69314718056)) // log2 approximation
	}
	// Simplified entropy for heuristic
	return float64(len(m)) / float64(len(s))
}

func analyzePath(path string, aiFlag bool) []AnalysisResult {
	info, err := os.Stat(path)
	if err != nil {
		return []AnalysisResult{{
			File:     path,
			Findings: []Finding{{Type: "error", Message: fmt.Sprintf("Path error: %v", err), Severity: "block"}},
		}}
	}

	if !info.IsDir() {
		return []AnalysisResult{analyzeFile(path, aiFlag)}
	}

	return walkAndAnalyze(path, aiFlag)
}

func walkAndAnalyze(root string, aiFlag bool) []AnalysisResult {
	var results []AnalysisResult
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Worker pool setup
	numWorkers := runtime.NumCPU() * 2
	fileChan := make(chan string, 100)

	// Start workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range fileChan {
				res := analyzeFile(path, aiFlag)
				mu.Lock()
				results = append(results, res)
				mu.Unlock()
			}
		}()
	}

	// Walk directory
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "dist" || name == "build" {
				return filepath.SkipDir
			}
			return nil
		}

		// Included extensions
		ext := strings.ToLower(filepath.Ext(path))
		if ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx" || ext == ".go" || ext == ".py" {
			fileChan <- path
		}
		return nil
	})

	close(fileChan)
	wg.Wait()

	return results
}

func analyzeFile(filePath string, aiFlag bool) AnalysisResult {
	result := AnalysisResult{
		File:       filePath,
		Findings:   []Finding{},
		AIInsights: []AIInsight{},
	}

	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		// PRO-GRADE: Gracefully handle deleted files in staged check
		if os.IsNotExist(err) {
			return result
		}

		result.Findings = append(result.Findings, Finding{
			Type:     "error",
			Message:  fmt.Sprintf("Could not read file: %v", err),
			Severity: "block",
		})
		return result
	}

	// 1. Precise AST Analysis
	analyzeAST(content, &result, policySeverity, filePath)

	// 2. Integration with AI Providers
	if len(result.Findings) > 0 && (aiFlag || os.Getenv("THAROS_AI_AUTO_EXPLAIN") == "true") {
		aiRes := getAIInsight(string(content), result.Findings)
		if len(aiRes.AIInsights) > 0 {
			result.AIInsights = append(result.AIInsights, aiRes.AIInsights...)
		}
	}

	return result
}

// Rule definition for extensibility
type RuleCheck func(tokenType js.TokenType, text string, prevToken js.TokenType, prevText string, pPrevToken js.TokenType, pPrevText string, currentLine int, byteOffset int, filePath string) *Finding

func analyzeAST(content []byte, result *AnalysisResult, defaultSeverity string, filePath string) {
	// FIX: Handle Hashbang/Shebang lines by masking them with spaces to preserve line numbers
	if len(content) > 2 && string(content[0:2]) == "#!" {
		for i := 0; i < len(content); i++ {
			if content[i] == '\n' {
				break
			}
			content[i] = ' '
		}
	}

	input := parse.NewInputBytes(content)
	lexer := js.NewLexer(input)

	currentLine := 1
	byteOffset := 0
	var prevToken js.TokenType
	var prevText string
	var pPrevToken js.TokenType
	var pPrevText string

	var ignoreLine int

	// Define our rules
	rules := []RuleCheck{
		// Rule: Check for dangerous functions (eval, exec)
		func(tt js.TokenType, text string, pt js.TokenType, pText string, ppt js.TokenType, ppText string, line int, offset int, filePath string) *Finding {
			if tt == js.IdentifierToken && (text == "eval" || text == "exec") {
				return &Finding{
					Rule:        "security.code_injection",
					Type:        "security_code_injection",
					Message:     fmt.Sprintf("Dangerous function '%s' detected.", text),
					Severity:    "high",
					Confidence:  0.9,
					Explain:     "Functions like eval() can execute arbitrary strings as code, leading to injection vulnerabilities.",
					Remediation: "Avoid eval(). Use JSON.parse() for data or refactor to use explicit logic.",
					Line:        line,
					ByteOffset:  offset,
					ByteLength:  len(text),
				}
			}
			return nil
		},
		// Rule: PRO-GRADE Credential Detection
		func(tt js.TokenType, text string, pt js.TokenType, pText string, ppt js.TokenType, ppText string, line int, offset int, filePath string) *Finding {
			// Rule: Stop flagging variable NAMES alone.
			// Only flag if it's a hardcoded value (Literal) assigned to a sensitive key.
			if (tt == js.StringToken || tt == js.TemplateToken) && (pt == js.EqToken || pt == js.ColonToken) {
				lowerKey := strings.ToLower(ppText)
				isSensitiveKey := strings.Contains(lowerKey, "pass") || strings.Contains(lowerKey, "secret") ||
					strings.Contains(lowerKey, "token") || strings.Contains(lowerKey, "apikey") ||
					strings.Contains(lowerKey, "pwd") || strings.Contains(lowerKey, "access_key")

				if isSensitiveKey {
					cleanVal := strings.Trim(text, "\"'` ")
					entropy := calculateEntropy(cleanVal)

					// Heuristic: Length > 6 and (high entropy or specific keywords)
					if len(cleanVal) > 6 && (entropy > 0.4 || isSensitiveKey) {
						severity := "critical"
						confidence := 0.85

						// Lower severity in test/fixture environments
						lowerPath := strings.ToLower(filePath)
						if strings.Contains(lowerPath, "test") || strings.Contains(lowerPath, "fixture") ||
							strings.Contains(lowerPath, "example") || strings.Contains(lowerPath, "mock") {
							severity = "info"
							confidence = 0.5
						}

						return &Finding{
							Rule:        "security.hardcoded_credential",
							Type:        "security_credential",
							Message:     fmt.Sprintf("Hardcoded secret detected in assignment to '%s'.", ppText),
							Severity:    severity,
							Confidence:  confidence,
							Explain:     "Storing secrets in source code is high risk. Attackers can extract these to gain unauthorized access.",
							Remediation: "Move this value to an environment variable (.env) or a secret manager.",
							Line:        line,
							ByteOffset:  offset,
							ByteLength:  len(text),
						}
					}
				}
			}

			// Pattern-based detection for known formats (AWS, Stripe, etc)
			if tt == js.StringToken || tt == js.TemplateToken {
				cleanVal := strings.Trim(text, "\"'` ")
				if strings.HasPrefix(cleanVal, "sk_live_") || strings.HasPrefix(cleanVal, "AKIA") ||
					(len(cleanVal) > 30 && calculateEntropy(cleanVal) > 0.7 && !strings.Contains(cleanVal, " ")) {
					return &Finding{
						Rule:        "security.secret_pattern",
						Type:        "security_credential",
						Message:     "High-entropy string detected; possible hardcoded API key or token.",
						Severity:    "critical",
						Confidence:  0.95,
						Explain:     "This string matches patterns commonly used for high-entropy API keys or access tokens.",
						Remediation: "Rotate this secret immediately and move it to a secure vault.",
						Line:        line,
						ByteOffset:  offset,
						ByteLength:  len(text),
					}
				}
			}
			return nil
		},
		// Rule: Check for XSS
		func(tt js.TokenType, text string, pt js.TokenType, pText string, ppt js.TokenType, ppText string, line int, offset int, filePath string) *Finding {
			if tt == js.IdentifierToken {
				if text == "innerHTML" || text == "outerHTML" || text == "dangerouslySetInnerHTML" {
					severity := "high"
					// Context Awareness: Lower risk in tests
					lowerPath := strings.ToLower(filePath)
					if strings.Contains(lowerPath, "test") || strings.Contains(lowerPath, "spec") || strings.Contains(lowerPath, "mock") {
						severity = "info"
					}

					return &Finding{
						Rule:        "security.xss",
						Type:        "security_xss",
						Message:     fmt.Sprintf("Usage of '%s' can lead to XSS vulnerabilities.", text),
						Severity:    severity,
						Confidence:  0.8,
						Explain:     "Directly inserting HTML strings into the DOM bypasses sanitization and can execute malicious scripts.",
						Remediation: "Use textContent/innerText or a sanitization library like DOMPurify.",
						Line:        line,
						ByteOffset:  offset,
						ByteLength:  len(text),
						Replacement: "textContent",
					}
				}
			}
			return nil
		},
		// Rule: Advanced SQL Injection
		func(tt js.TokenType, text string, pt js.TokenType, pText string, ppt js.TokenType, ppText string, line int, offset int, filePath string) *Finding {
			if tt == js.StringToken || tt == js.TemplateToken || tt == js.TemplateStartToken || tt == js.TemplateMiddleToken || tt == js.TemplateEndToken {
				upperText := strings.ToUpper(text)
				// Look for SQL keywords
				isSQL := strings.Contains(upperText, "SELECT") || strings.Contains(upperText, "INSERT") ||
					strings.Contains(upperText, "UPDATE") || strings.Contains(upperText, "DELETE") ||
					strings.Contains(upperText, "FROM") || strings.Contains(upperText, "WHERE")

				if isSQL {
					// Check for dangerous interpolation
					// StringToken with ${ is rare but possible in some contexts; Template tokens imply interpolation if not TemplateToken
					isInterpolated := tt == js.TemplateStartToken || tt == js.TemplateMiddleToken || strings.Contains(text, "${")

					if isInterpolated {
						severity := "critical"
						// Context Awareness: Lower risk in tests
						lowerPath := strings.ToLower(filePath)
						if strings.Contains(lowerPath, "test") || strings.Contains(lowerPath, "spec") || strings.Contains(lowerPath, "mock") {
							severity = "info"
						}

						return &Finding{
							Rule:        "security.sqli",
							Type:        "security_sqli",
							Message:     "High-risk SQL Injection: Direct variable interpolation in query.",
							Severity:    severity,
							Confidence:  0.98,
							Explain:     "Detected ${...} or template literal pattern inside a SQL-like string. This allows attackers to bypass query logic via malicious input.",
							Remediation: "Use parameterized queries (e.g., db.query('...', [val])) or an ORM.",
							Line:        line,
							ByteOffset:  offset,
							ByteLength:  len(text),
						}
					}
				}
			}
			return nil
		},
		// Rule: Insecure Routes + Auth Awareness
		func(tt js.TokenType, text string, pt js.TokenType, pText string, ppt js.TokenType, ppText string, line int, offset int, filePath string) *Finding {
			if tt == js.StringToken {
				lower := strings.ToLower(text)
				if strings.Contains(lower, "/admin") || strings.Contains(lower, "/debug") || strings.Contains(lower, "/config") {
					severity := "high"
					explain := "Sensitive route pattern detected. Ensure authentication is enforced."

					// Simple Heuristic: If NODE_ENV=test is nearby or in path, lower severity
					if strings.Contains(strings.ToLower(filePath), "test") {
						severity = "info"
					}

					return &Finding{
						Rule:        "security.insecure_route",
						Type:        "security_insecure_route",
						Message:     fmt.Sprintf("Sensitive route '%s' detected.", text),
						Severity:    severity,
						Confidence:  0.7,
						Explain:     explain,
						Remediation: "Verify this route is protected by auth middleware. Avoid exposing /debug in production.",
						Line:        line,
						ByteOffset:  offset,
						ByteLength:  len(text),
					}
				}
			}
			return nil
		},
		// Rule: Data Leaks (process.env)
		func(tt js.TokenType, text string, pt js.TokenType, pText string, ppt js.TokenType, ppText string, line int, offset int, filePath string) *Finding {
			if tt == js.IdentifierToken && text == "env" && pText == "process" {
				return &Finding{
					Rule:        "security.leak",
					Type:        "security_leak",
					Message:     "Direct exposure of 'process.env' detected.",
					Severity:    "high",
					Confidence:  0.9,
					Explain:     "Exposing the entire process.env object can leak system keys, DB strings, and internal config.",
					Remediation: "Only expose specific, non-sensitive environment variables.",
					Line:        line,
					ByteOffset:  offset,
					ByteLength:  len(text),
				}
			}
			return nil
		},
	}

	for {
		tt, data := lexer.Next()
		if tt == js.ErrorToken {
			if lexer.Err() != io.EOF {
				result.Findings = append(result.Findings, Finding{
					Rule:     "parse.error",
					Type:     "parse_error",
					Message:  fmt.Sprintf("AST Lexer Error: %v", lexer.Err()),
					Severity: "medium",
				})
			}
			break
		}

		text := string(data)
		currentLine += strings.Count(text, "\n")

		// Handle security-ignore
		if tt == js.CommentToken {
			if strings.Contains(text, "tharos-security-ignore") {
				ignoreLine = currentLine + 1
			}
		}

		// Run all rules
		for _, rule := range rules {
			if f := rule(tt, text, prevToken, prevText, pPrevToken, pPrevText, currentLine, byteOffset, filePath); f != nil {
				if currentLine == ignoreLine || currentLine == ignoreLine-1 {
					continue
				}
				result.Findings = append(result.Findings, *f)
			}
		}

		// Keep track of context
		if tt != js.LineTerminatorToken && tt != js.WhitespaceToken && tt != js.CommentToken {
			pPrevToken = prevToken
			pPrevText = prevText
			prevToken = tt
			prevText = text
		}

		byteOffset += len(data)
	}
}

func getAIInsight(code string, findings []Finding) AnalysisResult {
	res := AnalysisResult{AIInsights: []AIInsight{}}

	// Try Gemini (Personal)
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey != "" {
		insightStr := getGeminiInsight(code, findings, geminiKey)
		if !strings.Contains(insightStr, "AI Insight Unavailable") {
			parseAIResponse(insightStr, &res)
			return res
		}
	}

	// Fallback: Groq (Personal)
	groqKey := os.Getenv("GROQ_API_KEY")
	if groqKey != "" {
		insightStr := getGroqInsight(code, findings, groqKey)
		if !strings.Contains(insightStr, "AI Insight Unavailable") {
			parseAIResponse(insightStr, &res)
			return res
		}
	}

	return res
}

func getGeminiInsight(code string, findings []Finding, apiKey string) string {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return fmt.Sprintf("AI Insight Unavailable: (Gemini Client Error %v)", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	prompt := generatePrompt(code, findings)

	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return fmt.Sprintf("AI Insight Unavailable: (Gemini Error %v)", err)
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "AI Insight Unavailable: Gemini returned no content."
	}

	return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
}

func getGroqInsight(code string, findings []Finding, apiKey string) string {
	url := "https://api.groq.com/openai/v1/chat/completions"
	prompt := generatePrompt(code, findings)

	payload := map[string]interface{}{
		"model": "llama-3.3-70b-versatile",
		"messages": []map[string]string{
			{"role": "user", "content": prompt},
		},
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Sprintf("AI Insight Unavailable: (Groq HTTP Error %v)", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Sprintf("AI Insight Unavailable: (Groq Status %d: %s)", resp.StatusCode, string(body))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return fmt.Sprintf("AI Insight Unavailable: (Groq Decode Error %v)", err)
	}

	if len(result.Choices) == 0 {
		return "AI Insight Unavailable: Groq returned no choices."
	}

	return result.Choices[0].Message.Content
}

func parseAIResponse(response string, result *AnalysisResult) {
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]
		var insight AIInsight
		if err := json.Unmarshal([]byte(jsonStr), &insight); err == nil {
			result.AIInsights = append(result.AIInsights, insight)

			// Apply AI-suggested fixes to findings
			for _, fix := range insight.Fixes {
				for i := range result.Findings {
					// Match by line or type
					if result.Findings[i].Line == fix.Line && (fix.Type == "" || result.Findings[i].Type == fix.Type) {
						if result.Findings[i].Replacement == "" || os.Getenv("THAROS_AI_PRIORITY") == "true" {
							result.Findings[i].Replacement = fix.Replacement
							result.Findings[i].Message += " (AI Optimized Fix Available)"
						}
					}
				}
			}
			return
		}
	}

	// Fallback if not valid JSON
	result.AIInsights = append(result.AIInsights, AIInsight{
		Recommendation: response,
		RiskScore:      50,
	})
}

func generatePrompt(code string, findings []Finding) string {
	prompt := `Analyze the following code snippet and its security/quality findings. 
Use a "Professional Security Scanner Mindset":
- Real secrets (high entropy, known patterns) are Critical.
- Logic vulnerabilities are High.
- Variable names alone (without hardcoded values) are usually Info/Safe.
- Understand context: code in test folders or mocks should have lower risk profiles.

Provide your response in RAW JSON format ONLY. Do not use markdown code blocks.
Ensure all newlines in the "suggested_fix" are escaped as \n.

JSON Keys:
- "recommendation": A professional, actionable advice for the engineer.
- "risk_score": An integer 0-100.
- "suggested_fix": (Optional) A code snippet fixing the issue.
- "fixes": (Optional) A list of specific replacements: [{"line": N, "type": "T", "replacement": "new_text"}]

Code Context:
%s

Issues Found by AST Engine:
`
	for _, f := range findings {
		prompt += fmt.Sprintf("- [%s] Rule: %s - %s (Line %d, Severity: %s)\n", f.Type, f.Rule, f.Message, f.Line, f.Severity)
	}

	return fmt.Sprintf(prompt, code)
}

// --- SARIF Export Support ---

type SARIFReport struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name           string      `json:"name"`
	InformationURI string      `json:"informationUri"`
	Rules          []SARIFRule `json:"rules"`
}

type SARIFRule struct {
	ID               string           `json:"id"`
	ShortDescription SARIFDescription `json:"shortDescription"`
}

type SARIFDescription struct {
	Text string `json:"text"`
}

type SARIFResult struct {
	RuleID    string           `json:"ruleId"`
	Level     string           `json:"level"`
	Message   SARIFDescription `json:"message"`
	Locations []SARIFLocation  `json:"locations"`
}

type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

type SARIFRegion struct {
	StartLine int `json:"startLine"`
}

func ConvertToSARIF(results []AnalysisResult) SARIFReport {
	report := SARIFReport{
		Version: "2.1.0",
		Schema:  "https://schemastore.azurewebsites.net/schemas/json/sarif-2.1.0-rtm.5.json",
		Runs: []SARIFRun{
			{
				Tool: SARIFTool{
					Driver: SARIFDriver{
						Name:           "Tharos",
						InformationURI: "https://tharos.dev",
						Rules:          []SARIFRule{},
					},
				},
				Results: []SARIFResult{},
			},
		},
	}

	ruleMap := make(map[string]bool)

	for _, res := range results {
		for _, f := range res.Findings {
			// Map severity to SARIF level
			level := "warning"
			switch f.Severity {
			case "critical", "high":
				level = "error"
			case "medium":
				level = "warning"
			case "info":
				level = "note"
			}

			// Add rule to driver if not exists
			if !ruleMap[f.Rule] {
				report.Runs[0].Tool.Driver.Rules = append(report.Runs[0].Tool.Driver.Rules, SARIFRule{
					ID: f.Rule,
					ShortDescription: SARIFDescription{
						Text: f.Explain,
					},
				})
				ruleMap[f.Rule] = true
			}

			// Add result
			report.Runs[0].Results = append(report.Runs[0].Results, SARIFResult{
				RuleID: f.Rule,
				Level:  level,
				Message: SARIFDescription{
					Text: f.Message,
				},
				Locations: []SARIFLocation{
					{
						PhysicalLocation: SARIFPhysicalLocation{
							ArtifactLocation: SARIFArtifactLocation{
								URI: res.File,
							},
							Region: SARIFRegion{
								StartLine: f.Line,
							},
						},
					},
				},
			})
		}
	}

	return report
}
