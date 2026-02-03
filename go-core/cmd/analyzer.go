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
	Type        string `json:"type"`
	Message     string `json:"message"`
	Severity    string `json:"severity"` // "warning", "block", "info"
	Line        int    `json:"line"`
	ByteOffset  int    `json:"byte_offset,omitempty"`
	ByteLength  int    `json:"byte_length,omitempty"`
	Replacement string `json:"replacement,omitempty"`
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
		fmt.Printf("%s⚙️  Auto-fix mode enabled - applying fixes...%s\n", colorYellow, colorReset)
		appliedCount := applyFixes(result.Results)
		if appliedCount > 0 {
			fmt.Printf("%s✅ Applied %d fixes across analyzed files.%s\n", colorGreen, appliedCount, colorReset)
		} else {
			fmt.Printf("%sℹ️  No auto-fixes were available for the detected issues.%s\n", colorGray, colorReset)
		}
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
			err = ioutil.WriteFile(res.File, newContent, 0644)
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
	case "block":
		return "🔴"
	case "warning":
		return "⚠️ "
	case "info":
		return "ℹ️ "
	default:
		return "  "
	}
}

func getSeverityColor(severity string) string {
	switch severity {
	case "block":
		return colorRed
	case "warning":
		return colorYellow
	default:
		return colorReset
	}
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
		result.Findings = append(result.Findings, Finding{
			Type:     "error",
			Message:  fmt.Sprintf("Could not read file: %v", err),
			Severity: "block",
		})
		return result
	}

	// 1. Precise AST Analysis
	analyzeAST(content, &result, policySeverity)

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
type RuleCheck func(tokenType js.TokenType, text string, prevToken js.TokenType, prevText string, currentLine int, byteOffset int) *Finding

func analyzeAST(content []byte, result *AnalysisResult, defaultSeverity string) {
	input := parse.NewInputBytes(content)
	lexer := js.NewLexer(input)

	currentLine := 1
	byteOffset := 0
	var prevToken js.TokenType
	var prevText string

	// Define our rules
	rules := []RuleCheck{
		// Rule: Check for dangerous functions (eval, exec)
		func(tt js.TokenType, text string, pt js.TokenType, pText string, line int, offset int) *Finding {
			if tt == js.IdentifierToken && (text == "eval" || text == "exec") {
				return &Finding{
					Type:       "security_code_injection",
					Message:    fmt.Sprintf("Dangerous function '%s' detected.", text),
					Severity:   "block",
					Line:       line,
					ByteOffset: offset,
					ByteLength: len(text),
				}
			}
			return nil
		},
		// Rule: Check for hardcoded credentials
		func(tt js.TokenType, text string, pt js.TokenType, pText string, line int, offset int) *Finding {
			if tt == js.IdentifierToken {
				lowerId := strings.ToLower(text)
				if strings.Contains(lowerId, "api") || strings.Contains(lowerId, "key") || strings.Contains(lowerId, "secret") || strings.Contains(lowerId, "password") || strings.Contains(lowerId, "token") {
					// Only flag if it's likely a variable assignment or similar
					return &Finding{
						Type:       "security_credential",
						Message:    fmt.Sprintf("Identifier '%s' might contain sensitive data. Ensure it is not a hardcoded secret.", text),
						Severity:   "warning",
						Line:       line,
						ByteOffset: offset,
						ByteLength: len(text),
					}
				}
			}
			if tt == js.StringToken {
				val := text
				if len(val) > 35 && (strings.Contains(val, "-") || strings.Contains(val, "_")) && !strings.Contains(val, " ") {
					return &Finding{
						Type:       "security_credential",
						Message:    "Suspiciously long string literal detected; possible hardcoded token.",
						Severity:   defaultSeverity,
						Line:       line,
						ByteOffset: offset,
						ByteLength: len(text),
					}
				}
			}
			return nil
		},
		// Rule: Check for XSS
		func(tt js.TokenType, text string, pt js.TokenType, pText string, line int, offset int) *Finding {
			if tt == js.IdentifierToken {
				if text == "innerHTML" || text == "outerHTML" {
					return &Finding{
						Type:        "security_xss",
						Message:     fmt.Sprintf("Usage of '%s' can lead to XSS vulnerabilities. Use 'innerText' or 'textContent' instead.", text),
						Severity:    "warning",
						Line:        line,
						ByteOffset:  offset,
						ByteLength:  len(text),
						Replacement: "textContent",
					}
				}
				if text == "dangerouslySetInnerHTML" {
					return &Finding{
						Type:       "security_xss",
						Message:    "Usage of 'dangerouslySetInnerHTML' must be carefully reviewed.",
						Severity:   "warning",
						Line:       line,
						ByteOffset: offset,
						ByteLength: len(text),
					}
				}
				if text == "write" && pText == "document" {
					return &Finding{
						Type:       "security_xss",
						Message:    "Usage of 'document.write' is strongly discouraged.",
						Severity:   "warning",
						Line:       line,
						ByteOffset: offset,
						ByteLength: len(text),
					}
				}
			}
			return nil
		},
		// Rule: Check for Weak Cryptography
		func(tt js.TokenType, text string, pt js.TokenType, pText string, line int, offset int) *Finding {
			if tt == js.IdentifierToken {
				if text == "random" && pText == "." {
					return &Finding{
						Type:        "security_crypto",
						Message:     "Usage of .random() detected. Ensure this is not used for security-critical randomness.",
						Severity:    "warning",
						Line:        line,
						ByteOffset:  offset,
						ByteLength:  len(text),
						Replacement: "getRandomValues",
					}
				}
				if text == "md5" || text == "sha1" {
					return &Finding{
						Type:        "security_crypto",
						Message:     fmt.Sprintf("Weak hashing algorithm '%s' detected. Use SHA-256 or better.", text),
						Severity:    "warning",
						Line:        line,
						ByteOffset:  offset,
						ByteLength:  len(text),
						Replacement: "sha256",
					}
				}
			}
			return nil
		},
		// Rule: Basic SQL Injection Heuristic
		func(tt js.TokenType, text string, pt js.TokenType, pText string, line int, offset int) *Finding {
			if tt == js.StringToken {
				upperText := strings.ToUpper(text)
				if strings.Contains(upperText, "SELECT ") || strings.Contains(upperText, "UPDATE ") || strings.Contains(upperText, "INSERT INTO") {
					return &Finding{
						Type:       "security_sqli",
						Message:    "Possible SQL command detected. ensure variables are not concatenated directly.",
						Severity:   "info",
						Line:       line,
						ByteOffset: offset,
						ByteLength: len(text),
					}
				}
			}
			return nil
		},
		// Rule: Check for TODOs
		func(tt js.TokenType, text string, pt js.TokenType, pText string, line int, offset int) *Finding {
			if tt == js.CommentToken {
				if strings.Contains(text, "TODO") || strings.Contains(text, "FIXME") {
					return &Finding{
						Type:       "code_quality",
						Message:    "Unresolved task found in comments.",
						Severity:   "info",
						Line:       line,
						ByteOffset: offset,
						ByteLength: len(text),
					}
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
					Type:     "parse_error",
					Message:  fmt.Sprintf("AST Lexer Error: %v", lexer.Err()),
					Severity: "warning",
				})
			}
			break
		}

		text := string(data)
		currentLine += strings.Count(text, "\n")

		// Run all rules
		for _, rule := range rules {
			if f := rule(tt, text, prevToken, prevText, currentLine, byteOffset); f != nil {
				result.Findings = append(result.Findings, *f)
			}
		}

		// Keep track of limited context
		if tt != js.LineTerminatorToken && tt != js.WhitespaceToken && tt != js.CommentToken {
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
			if len(insight.Fixes) > 0 && os.Getenv("THAROS_VERBOSE") == "true" {
				fmt.Printf("[DEBUG] AI suggesting %d fixes\n", len(insight.Fixes))
			}
			for _, fix := range insight.Fixes {
				for i := range result.Findings {
					if result.Findings[i].Line == fix.Line && (fix.Type == "" || result.Findings[i].Type == fix.Type) {
						result.Findings[i].Replacement = fix.Replacement
						// Mark finding as AI-enhanced for the label
						result.Findings[i].Message += " (AI Enhanced)"
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
Provide your response in RAW JSON format ONLY. Do not use markdown code blocks.
Ensure all newlines in the "suggested_fix" are escaped as \n.

JSON Keys:
- "recommendation": A professional, actionable advice for the engineer.
- "risk_score": An integer 0-100.
- "suggested_fix": (Optional) A code snippet fixing the issue.
- "fixes": (Optional) A list of specific replacements for individual findings. Use the format: [{"line": N, "type": "T", "replacement": "new_text"}]
  The "replacement" should be a direct replacement for the specific vulnerable token/expression identified on that line.

Code Context:
%s

Issues Found:
`
	for _, f := range findings {
		prompt += fmt.Sprintf("- [%s] %s (Line %d)\n", f.Type, f.Message, f.Line)
	}

	return fmt.Sprintf(prompt, code)
}
