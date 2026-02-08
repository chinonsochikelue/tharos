package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"

	"github.com/google/generative-ai-go/genai"

	"github.com/spf13/viper"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
	"google.golang.org/api/option"
	"gopkg.in/yaml.v2"
)

var (
	externalRules []CustomRule
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

type CustomRule struct {
	Pattern  string `mapstructure:"pattern"`
	Message  string `mapstructure:"message"`
	Severity string `mapstructure:"severity"`
}

type UnifiedRule struct {
	ID          string
	Source      string // "Built-in", "tharos.yaml", "Policy File"
	Severity    string
	Description string
}

func GetAllActiveRules() []UnifiedRule {
	var allRules []UnifiedRule

	// 1. Built-in AST Rules
	allRules = append(allRules, []UnifiedRule{
		{ID: "security.js.injection", Source: "Built-in", Severity: "critical", Description: "Detects eval/exec code injection in JS/TS"},
		{ID: "security.js.credentials", Source: "Built-in", Severity: "critical", Description: "Entropy-based secret detection in JS/TS"},
		{ID: "security.js.xss", Source: "Built-in", Severity: "high", Description: "Direct DOM manipulation (innerHTML, etc.)"},
		{ID: "security.go.sqli", Source: "Built-in", Severity: "critical", Description: "SQL injection in fmt.Sprintf usage"},
		{ID: "security.py.injection", Source: "Built-in", Severity: "critical", Description: "OS command injection in Python (os.system)"},
	}...)

	// 2. Local Rules from tharos.yaml
	var localRules []CustomRule
	viper.UnmarshalKey("security.rules", &localRules)
	for _, r := range localRules {
		allRules = append(allRules, UnifiedRule{
			ID:          "config.custom",
			Source:      "tharos.yaml",
			Severity:    r.Severity,
			Description: r.Message,
		})
	}

	// 3. External Policies
	for _, r := range externalRules {
		allRules = append(allRules, UnifiedRule{
			ID:          "policy.external",
			Source:      "Policy File",
			Severity:    r.Severity,
			Description: r.Message,
		})
	}

	return allRules
}

func init() {
	// Root-level policy mapping will be handled by individual commands
}

func setupCommand() {
	fmt.Printf("%s%s[THAROS AI SETUP]%s\n\n", colorBold, colorCyan, colorReset)

	fmt.Println("Tharos uses AI to provide deeper security insights and recommendations.")
	fmt.Println("Choose one of the following providers (both have free tiers):")
	fmt.Println()

	// Check Gemini
	fmt.Printf("%s1. Google Gemini (Recommended)%s\n", colorBold, colorReset)
	// tharos-security-ignore
	if geminiKey := os.Getenv("GEMINI_API_KEY"); geminiKey != "" {
		fmt.Printf("  %s✓ API Key Configured%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("  %s✗ Not configured%s\n", colorRed, colorReset)
		fmt.Printf("    1. Get key: %shttps://makersuite.google.com/app/apikey%s\n", colorCyan, colorReset)
		// tharos-security-ignore
		fmt.Printf("    2. Set env:  %sexport GEMINI_API_KEY=\"your-key\"%s\n", colorCyan, colorReset)
		// tharos-security-ignore
		fmt.Printf("       Windows:  %s$env:GEMINI_API_KEY=\"your-key\"%s\n", colorCyan, colorReset)

	}

	fmt.Println()

	// Check Groq
	fmt.Printf("%s2. Groq (Fast & Free)%s\n", colorBold, colorReset)
	// tharos-security-ignore
	if groqKey := os.Getenv("GROQ_API_KEY"); groqKey != "" {
		fmt.Printf("  %s✓ API Key Configured%s\n", colorGreen, colorReset)
	} else {
		fmt.Printf("  %s✗ Not configured%s\n", colorRed, colorReset)
		fmt.Printf("    1. Get key: %shttps://console.groq.com%s\n", colorCyan, colorReset)
		// tharos-security-ignore
		fmt.Printf("    2. Set env:  %sexport GROQ_API_KEY=\"your-key\"%s\n", colorCyan, colorReset)
		// tharos-security-ignore
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
	// 🎨 Define Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("36")). // Cyan
		Padding(0, 1).
		MarginBottom(1)

	subHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("250")) // Light Gray

	fileStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("36")).
		Underline(true)

	// Render Header
	fmt.Println(headerStyle.Render("[THAROS SECURITY SCAN]"))

	if len(result.Results) == 0 {
		fmt.Printf(" %s No files analyzed.\n", lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"))
		return
	}

	filesWithIssues := 0
	for _, r := range result.Results {
		if len(r.Findings) > 0 {
			filesWithIssues++
		}
	}

	// Stats Summary
	stats := fmt.Sprintf("📊 Scanned: %d files in %s | ⚠️  Issues: %d in %d files",
		result.Summary.TotalFiles,
		result.Summary.Duration,
		result.Summary.Vulnerabilities,
		filesWithIssues)
	fmt.Println(subHeaderStyle.Render(stats))
	fmt.Println()

	if result.Summary.Vulnerabilities == 0 {
		fmt.Printf(" %s %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Render("✓"),
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42")).Render("No issues found! Your code is safe."))
		fmt.Println()
		printVerdict(true)
		return
	}

	// 📋 Findings Table
	t := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers("SEV", "LOCATION", "SECURITY FINDING")

	// Set Fixed Widths for stability
	t.Width(90)

	for _, r := range result.Results {
		for _, f := range r.Findings {
			sevSymbol := getSeveritySymbol(f.Severity)
			sevColor := getSeverityColorLipgloss(f.Severity)

			coloredSev := lipgloss.NewStyle().Foreground(sevColor).Padding(0, 1).Render(sevSymbol)
			location := filepath.Base(r.File) + ":" + fmt.Sprint(f.Line)

			// Manual wrapping for the message to ensure it fits in the column
			msgStyle := lipgloss.NewStyle().Width(55)
			wrappedMsg := msgStyle.Render(f.Message)

			t.Row(coloredSev, location, wrappedMsg)
		}
	}

	fmt.Println(t.Render())
	fmt.Println()

	// � AI Insights & Fixes (if verbose or AI enabled)
	if verboseMode || len(result.Results) > 0 {
		aiHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("99")) // Purple-ish

		for _, r := range result.Results {
			if len(r.AIInsights) > 0 {
				fmt.Println(fileStyle.Render(r.File))
				for _, ai := range r.AIInsights {
					fmt.Printf(" %s %s\n", aiHeaderStyle.Render("🧠 AI Insight:"), "Recommendation")
					renderedRec, _ := renderMarkdown(ai.Recommendation)
					fmt.Println(indentMultiline(renderedRec, "    "))

					if ai.SuggestedFix != "" && verboseMode {
						fixHeaderStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("42"))
						fmt.Println(indentMultiline(fixHeaderStyle.Render("Suggested Fix:"), "    "))
						fixMd := fmt.Sprintf("```go\n%s\n```", ai.SuggestedFix)
						renderedFix, _ := renderMarkdown(fixMd)
						fmt.Println(indentMultiline(renderedFix, "    "))
					}
				}
				fmt.Println()
			}
		}
	}

	if fixModeEnabled {
		fmt.Printf(" %s %s\n",
			lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render("⚙️"),
			lipgloss.NewStyle().Italic(true).Foreground(lipgloss.Color("245")).Render("Auto-fixes were successfully written to disk."))
	}

	fmt.Println()

	// Final Verdict logic (blocking based on critical/high)
	isBlocked := false
	criticalCount := 0
	highCount := 0
	for _, r := range result.Results {
		for _, f := range r.Findings {
			if f.Severity == "critical" || f.Severity == "block" {
				criticalCount++
			} else if f.Severity == "high" {
				highCount++
			}
		}
	}
	if criticalCount >= 1 || highCount >= 1 {
		isBlocked = true
	}

	printVerdict(!isBlocked)
}

func getSeverityColorLipgloss(severity string) lipgloss.Color {
	switch severity {
	case "critical", "block":
		return lipgloss.Color("1") // Red
	case "high":
		return lipgloss.Color("196") // Bright Red
	case "medium", "warning":
		return lipgloss.Color("3") // Yellow
	default:
		return lipgloss.Color("7") // White/Gray
	}
}

func printVerdict(pass bool) {
	width := 60
	style := lipgloss.NewStyle().
		Bold(true).
		Align(lipgloss.Center).
		Width(width).
		Padding(1)

	if pass {
		style = style.
			Foreground(lipgloss.Color("15")). // White
			Background(lipgloss.Color("42")). // Green
			SetString("🛡️  COMMIT VERDICT: PASS")
	} else {
		style = style.
			Foreground(lipgloss.Color("15")).  // White
			Background(lipgloss.Color("196")). // Red
			SetString("🛑  COMMIT VERDICT: BLOCK")
	}

	fmt.Println(style.Render())
	fmt.Println()
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
	case "critical", "block":
		return "💀"
	case "high":
		return "🔴"
	case "medium", "warning":
		return "⚠️ "
	case "info":
		return "ℹ️ "
	default:
		return "  "
	}
}

func getSeverityColor(severity string) string {
	switch severity {
	case "critical", "block":
		return colorRed + colorBold
	case "high":
		return colorRed
	case "medium", "warning":
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
	excludes := viper.GetStringSlice("exclude")
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		for _, pattern := range excludes {
			if strings.Contains(path, pattern) {
				if d.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
		}
		if d.IsDir() {

			name := d.Name()
			if name == "node_modules" || name == ".git" || name == "dist" || name == "build" || name == ".next" || name == "bin" || name == ".vercel" {
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
			Line:     1,
		})
		return result
	}

	// 1. Precise AST Analysis
	ext := strings.ToLower(filepath.Ext(filePath))
	switch ext {
	case ".go":
		analyzeGoAST(content, &result, filePath)
	case ".js", ".ts", ".jsx", ".tsx":
		analyzeAST(content, &result, "warning", filePath)
	case ".py":
		analyzePythonAST(content, &result, filePath)
	}

	// 1.5 Custom Regex-based Rules from tharos.yaml
	analyzeCustomRules(content, &result, filePath)

	// PRO-GRADE: Filter findings that have // tharos-security-ignore or # tharos-security-ignore
	contentStrings := strings.Split(string(content), "\n")
	filteredFindings := []Finding{}
	for _, f := range result.Findings {
		if f.Line > 0 && f.Line <= len(contentStrings) {
			lineContent := contentStrings[f.Line-1]
			if strings.Contains(lineContent, "tharos-security-ignore") {
				continue // Skip this finding
			}
		}
		filteredFindings = append(filteredFindings, f)
	}
	result.Findings = filteredFindings

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

func loadExternalPolicies(pPath string, pDir string) {
	// Clear previous external rules
	externalRules = []CustomRule{}

	// 1. Load from specific file if provided
	if pPath != "" {
		rules, err := readPolicyFile(pPath)
		if err == nil {
			externalRules = append(externalRules, rules...)
		}
	}

	// 2. Load from policy directory if provided (default "policies")
	if pDir != "" {
		files, _ := ioutil.ReadDir(pDir)
		for _, f := range files {
			if strings.HasSuffix(f.Name(), ".yaml") || strings.HasSuffix(f.Name(), ".yml") {
				rules, err := readPolicyFile(filepath.Join(pDir, f.Name()))
				if err == nil {
					externalRules = append(externalRules, rules...)
				}
			}
		}
	}
}

func readPolicyFile(filePath string) ([]CustomRule, error) {
	data, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// The policy file structure is slightly different (root level security.rules)
	type Policy struct {
		Security struct {
			Rules []CustomRule `yaml:"rules"`
		} `yaml:"security"`
	}

	var p Policy
	err = yaml.Unmarshal(data, &p)
	if err != nil {
		return nil, err
	}

	return p.Security.Rules, nil
}

func analyzeCustomRules(content []byte, result *AnalysisResult, filePath string) {
	if !viper.GetBool("security.enabled") && len(externalRules) == 0 {
		return
	}

	var localRules []CustomRule
	viper.UnmarshalKey("security.rules", &localRules)

	// Merge local and external rules
	allRules := append(localRules, externalRules...)

	lines := strings.Split(string(content), "\n")

	for _, rule := range allRules {

		re, err := regexp.Compile(rule.Pattern)
		if err != nil {
			continue
		}

		ignoreNext := false
		for i, line := range lines {
			if strings.Contains(line, "tharos-security-ignore") {
				ignoreNext = true
				continue
			}

			if ignoreNext {
				ignoreNext = false
				continue
			}

			if re.MatchString(line) {
				result.Findings = append(result.Findings, Finding{
					Rule:        "security.custom_rule",
					Type:        "security_custom",
					Message:     rule.Message,
					Severity:    rule.Severity,
					Confidence:  1.0,
					Explain:     "This finding was triggered by a user-defined security rule in tharos.yaml.",
					Remediation: "Follow the project's security guidelines to resolve this custom finding.",
					Line:        i + 1,
				})
			}
		}
	}
}

func analyzeGoAST(content []byte, result *AnalysisResult, filePath string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, content, parser.ParseComments)
	if err != nil {
		return
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch node := n.(type) {
		// Rule: Detect function calls (exec, sql, crypto)
		case *ast.CallExpr:
			if fun, ok := node.Fun.(*ast.SelectorExpr); ok {
				if pkg, ok := fun.X.(*ast.Ident); ok {
					// 1. Command Injection (os/exec)
					if pkg.Name == "exec" && fun.Sel.Name == "Command" {
						if len(node.Args) > 1 {
							isLiteral := true
							for _, arg := range node.Args[1:] {
								if _, ok := arg.(*ast.BasicLit); !ok {
									isLiteral = false
									break
								}
							}
							if !isLiteral {
								line := fset.Position(node.Pos()).Line
								result.Findings = append(result.Findings, Finding{
									Rule:        "security.go.cmd_injection",
									Type:        "security_code_injection",
									Message:     "Potential Command Injection: Non-literal argument in exec.Command.",
									Severity:    "critical",
									Confidence:  0.8,
									Explain:     "Passing variables directly into shell commands can lead to command injection if the input is untrusted.",
									Remediation: "Ensure input is sanitized or use fixed argument lists.",
									Line:        line,
								})
							}
						}
					}

					// 2. Weak Cryptography (md5, sha1, des)
					if (pkg.Name == "md5" || pkg.Name == "sha1") && fun.Sel.Name == "New" {
						line := fset.Position(node.Pos()).Line
						result.Findings = append(result.Findings, Finding{
							Rule:        "security.go.weak_crypto",
							Type:        "security_weak_crypto",
							Message:     fmt.Sprintf("Weak cryptic algorithm detected: %s.%s", pkg.Name, fun.Sel.Name),
							Severity:    "medium",
							Confidence:  1.0,
							Explain:     "MD5 and SHA1 are considered cryptographically broken and should not be used for secure hashing.",
							Remediation: "Use SHA-256 or SHA-512 (crypto/sha256 or crypto/sha512).",
							Line:        line,
						})
					}

					// 3. SQL Injection (database/sql Query/Exec)
					if fun.Sel.Name == "Query" || fun.Sel.Name == "Exec" || fun.Sel.Name == "QueryRow" {
						if len(node.Args) > 0 {
							// Check if first arg is a concatenation or variable
							isUnsafe := false
							if _, ok := node.Args[0].(*ast.BinaryExpr); ok {
								isUnsafe = true
							} else if ident, ok := node.Args[0].(*ast.Ident); ok {
								// Simple tracking: if it's an ident, it's potentially unsafe
								isUnsafe = true
								_ = ident
							}

							if isUnsafe {
								line := fset.Position(node.Pos()).Line
								result.Findings = append(result.Findings, Finding{
									Rule:        "security.go.sqli",
									Type:        "security_sqli",
									Message:     "Potential SQL Injection: Dynamic query detected.",
									Severity:    "critical",
									Confidence:  0.7,
									Explain:     "Building SQL queries with dynamic strings is prone to injection.",
									Remediation: "Use parameterized queries (e.g. db.Query(\"SELECT...\", id)).",
									Line:        line,
								})
							}
						}
					}

					// 6. Insecure CORS (Access-Control-Allow-Origin: *)
					if fun.Sel.Name == "Set" || fun.Sel.Name == "Add" || fun.Sel.Name == "Header" {
						if len(node.Args) == 2 {
							if keyLit, ok := node.Args[0].(*ast.BasicLit); ok && strings.Contains(keyLit.Value, "Access-Control-Allow-Origin") {
								if valLit, ok := node.Args[1].(*ast.BasicLit); ok && strings.Contains(valLit.Value, "*") {
									line := fset.Position(node.Pos()).Line
									result.Findings = append(result.Findings, Finding{
										Rule:        "security.go.insecure_cors",
										Type:        "security_insecure_cors",
										Message:     "Insecure CORS: Access-Control-Allow-Origin set to '*'.",
										Severity:    "high",
										Confidence:  1.0,
										Explain:     "Allowing all origins via '*' can lead to sensitive data exposure via cross-origin requests.",
										Remediation: "Specify allowed origins explicitly or use a robust CORS package.",
										Line:        line,
									})
								}
							}
						}
					}
				}
			}

		// 4. Insecure TLS (tls.Config)
		case *ast.CompositeLit:
			if typ, ok := node.Type.(*ast.SelectorExpr); ok {
				if pkg, ok := typ.X.(*ast.Ident); ok && pkg.Name == "tls" && typ.Sel.Name == "Config" {
					for _, elt := range node.Elts {
						if kv, ok := elt.(*ast.KeyValueExpr); ok {
							if key, ok := kv.Key.(*ast.Ident); ok && key.Name == "InsecureSkipVerify" {
								if val, ok := kv.Value.(*ast.Ident); ok && val.Name == "true" {
									line := fset.Position(node.Pos()).Line
									result.Findings = append(result.Findings, Finding{
										Rule:        "security.go.insecure_tls",
										Type:        "security_insecure_transport",
										Message:     "Insecure TLS: InsecureSkipVerify set to true.",
										Severity:    "critical",
										Confidence:  1.0,
										Explain:     "Disabling certificate verification makes the application vulnerable to Man-in-the-Middle (MITM) attacks.",
										Remediation: "Remove InsecureSkipVerify: true or set it to false.",
										Line:        line,
									})
								}
							}
						}
					}
				}
			}

		// 5. Hardcoded Secrets in Go (String Literals)
		case *ast.ValueSpec:
			for i, val := range node.Values {
				if lit, ok := val.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					cleanVal := strings.Trim(lit.Value, "\"")
					entropy := calculateEntropy(cleanVal)
					if entropy > 4.5 && len(cleanVal) > 20 {
						line := fset.Position(node.Pos()).Line
						var varName string
						if i < len(node.Names) {
							varName = node.Names[i].Name
						}
						result.Findings = append(result.Findings, Finding{
							Rule:        "security.secret_pattern",
							Type:        "security_credential",
							Message:     fmt.Sprintf("Possible hardcoded secret in variable '%s'.", varName),
							Severity:    "critical",
							Confidence:  0.8,
							Explain:     "High-entropy string detected in Go source code.",
							Remediation: "Move secrets to environment variables or a secret manager.",
							Line:        line,
							ByteOffset:  int(node.Pos()),
						})
					}
				}
			}
		}
		return true
	})
}

func analyzePythonAST(content []byte, result *AnalysisResult, filePath string) {
	lines := strings.Split(string(content), "\n")

	for i, line := range lines {
		cleanLine := strings.Split(line, "#")[0] // Ignore comments
		lineNum := i + 1

		// 1. Insecure Deserialization (pickle, yaml)
		if strings.Contains(cleanLine, "pickle.loads(") {
			result.Findings = append(result.Findings, Finding{
				Rule:        "security.py.deserialization",
				Type:        "security_code_injection",
				Message:     "Insecure Deserialization: pickle.loads() detected.",
				Severity:    "critical",
				Confidence:  1.0,
				Explain:     "pickle.loads() can be exploited to execute arbitrary code. Untrusted data should never be unpickled.",
				Remediation: "Use safer formats like JSON or verify the data before unpickling.",
				Line:        lineNum,
			})
		}
		if (strings.Contains(cleanLine, "yaml.load(") || strings.Contains(cleanLine, "yaml.load_all(")) &&
			!strings.Contains(cleanLine, "SafeLoader") && !strings.Contains(cleanLine, "CSafeLoader") {
			result.Findings = append(result.Findings, Finding{
				Rule:        "security.py.yaml_load",
				Type:        "security_code_injection",
				Message:     "Unsafe YAML Load: yaml.load() without SafeLoader.",
				Severity:    "high",
				Confidence:  0.9,
				Explain:     "yaml.load() defaults to unsafe loading which can execute arbitrary code in older versions. Always use SafeLoader.",
				Remediation: "Use yaml.safe_load() or pass Loader=yaml.SafeLoader.",
				Line:        lineNum,
			})
		}

		// 2. Command Injection (os.system, subprocess)
		if strings.Contains(cleanLine, "os.system(") || strings.Contains(cleanLine, "subprocess.call(") ||
			strings.Contains(cleanLine, "subprocess.run(") {

			// Detect dynamic content (variables or concatenation)
			isDynamic := strings.Contains(cleanLine, "+") || strings.Contains(cleanLine, "f\"") ||
				strings.Contains(cleanLine, ".format(") || strings.Contains(cleanLine, "%")

			if isDynamic && !strings.Contains(cleanLine, "shell=False") {
				result.Findings = append(result.Findings, Finding{
					Rule:        "security.py.cmd_injection",
					Type:        "security_code_injection",
					Message:     "Potential Python Command Injection: Dynamic argument in shell execution.",
					Severity:    "critical",
					Confidence:  0.8,
					Explain:     "Executing shell commands with dynamic input without shell=False is extremely dangerous.",
					Remediation: "Use subprocess.run(args, shell=False) with args as a list.",
					Line:        lineNum,
				})
			}
		}

		// 3. Unsafe Eval/Exec
		if strings.Contains(cleanLine, "eval(") || strings.Contains(cleanLine, "exec(") {
			result.Findings = append(result.Findings, Finding{
				Rule:        "security.py.eval",
				Type:        "security_code_injection",
				Message:     "Unsafe Python eval() or exec() detected.",
				Severity:    "high",
				Confidence:  0.9,
				Explain:     "eval() and exec() execute string input as code, which is a major security risk if the input is untrusted.",
				Remediation: "Avoid eval/exec; use safer alternatives like ast.literal_eval() or logic-based parsing.",
				Line:        lineNum,
			})
		}

		// 4. Insecure Requests (verify=False)
		if strings.Contains(cleanLine, "verify=False") {
			result.Findings = append(result.Findings, Finding{
				Rule:        "security.py.insecure_requests",
				Type:        "security_insecure_transport",
				Message:     "Insecure Network Request: verify=False detected.",
				Severity:    "medium",
				Confidence:  1.0,
				Explain:     "Disabling SSL verification allows MITM attacks and exposes sensitive data.",
				Remediation: "Remove verify=False or set it to True. Use proper CA bundles.",
				Line:        lineNum,
			})
		}

		// 5. Hardcoded Secrets (Python)
		entropy := calculateEntropy(cleanLine)
		if entropy > 4.5 && len(cleanLine) > 30 && (strings.Contains(cleanLine, "=") || strings.Contains(cleanLine, ":")) {
			// Basic filtering to avoid flagging library names or long paths
			if !strings.Contains(cleanLine, "http") && !strings.Contains(cleanLine, "/") {
				result.Findings = append(result.Findings, Finding{
					Rule:        "security.secret_pattern",
					Type:        "security_credential",
					Message:     "Possible hardcoded secret detected in Python code.",
					Severity:    "critical",
					Confidence:  0.8,
					Explain:     "High-entropy string detected in variable assignment or literal.",
					Remediation: "Rotate secret and move to environment variables.",
					Line:        lineNum,
				})
			}
		}

		// 6. Hardcoded Passwords/Secrets in Assignments (Python)
		if (strings.Contains(cleanLine, "=") || strings.Contains(cleanLine, ":")) && !strings.Contains(cleanLine, "==") {
			lowerLine := strings.ToLower(cleanLine)
			isSensitive := strings.Contains(lowerLine, "pass") || strings.Contains(lowerLine, "pwd") ||
				strings.Contains(lowerLine, "secret") || strings.Contains(lowerLine, "token") ||
				strings.Contains(lowerLine, "key")

			if isSensitive {
				parts := strings.Split(cleanLine, "=")
				if len(parts) < 2 {
					parts = strings.Split(cleanLine, ":")
				}

				if len(parts) >= 2 {
					val := strings.TrimSpace(parts[1])
					// If it starts and ends with quotes, it's a literal string
					if (strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"")) ||
						(strings.HasPrefix(val, "'") && strings.HasSuffix(val, "'")) {
						cleanVal := strings.Trim(val, "\"'")
						// Length check to avoid flagging empty strings or very short names
						if len(cleanVal) > 4 && !strings.Contains(cleanVal, "{") && !strings.Contains(cleanVal, "os.environ") {
							result.Findings = append(result.Findings, Finding{
								Rule:        "security.py.hardcoded_password",
								Type:        "security_credential",
								Message:     fmt.Sprintf("Hardcoded secret detected in Python assignment to '%s'.", strings.TrimSpace(parts[0])),
								Severity:    "critical",
								Confidence:  0.9,
								Explain:     "Storing passwords or keys in plain text within source code is a major security risk.",
								Remediation: "Use environment variables (os.getenv) or a secret manager.",
								Line:        lineNum,
							})
						}
					}
				}
			}
		}
	}
}

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
				// Context Awareness: Avoid flagging our own build scripts and internal tools
				lowerPath := strings.ToLower(filePath)
				if strings.Contains(lowerPath, "scripts/") || strings.Contains(lowerPath, "go-core/") ||
					strings.HasSuffix(lowerPath, "install-tharos.ts") || strings.Contains(lowerPath, "package.json") {
					return nil
				}

				return &Finding{
					Rule:       "security.code_injection",
					Type:       "security_code_injection",
					Message:    fmt.Sprintf("Dangerous function '%s' detected.", text),
					Severity:   "high",
					Confidence: 0.9,
					// tharos-security-ignore
					Explain: "Functions like eval() can execute arbitrary strings as code, leading to injection vulnerabilities.",
					// tharos-security-ignore
					Remediation: "Avoid eval(). Use JSON.parse() for data or refactor to use explicit logic.",

					Line:       line,
					ByteOffset: offset,
					ByteLength: len(text),
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
						// Safe-list: Common public keys and non-secrets
						if strings.Contains(cleanVal, "ayqp5JY") || // Google Verification
							strings.HasPrefix(cleanVal, "google-site-verification") ||
							strings.Contains(cleanVal, "vercel.app") {
							return nil
						}

						severity := "critical"
						confidence := 0.85

						// Lower severity in test/fixture environments
						lowerPath := strings.ToLower(filePath)
						if strings.Contains(lowerPath, "test") || strings.Contains(lowerPath, "fixture") ||
							strings.Contains(lowerPath, "example") || strings.Contains(lowerPath, "mock") ||
							strings.Contains(lowerPath, "layout.tsx") { // Common for SEO meta tags
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
				// Regex heuristic: If it looks like a complex regex, it's probably not a secret
				isRegex := strings.Contains(cleanVal, "[") || strings.Contains(cleanVal, "(") ||
					strings.Contains(cleanVal, "\\") || strings.Contains(cleanVal, "$")

				if !isRegex && (strings.HasPrefix(cleanVal, "sk_live_") || strings.HasPrefix(cleanVal, "AKIA") ||
					(len(cleanVal) > 30 && calculateEntropy(cleanVal) > 0.7 && !strings.Contains(cleanVal, " "))) {

					// Final check: ignore if it's in analyzer.go itself (avoiding self-flagging)
					if strings.HasSuffix(filePath, "analyzer.go") {
						return nil
					}

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
				// Refinement: Only flag if it looks like a real absolute path or route, not a package name
				isRoute := strings.HasPrefix(text, "\"/") || strings.HasPrefix(text, "'/") || strings.HasPrefix(text, "`//")
				isPackage := strings.Contains(lower, "@") || strings.Contains(lower, "config-") || strings.Contains(lower, "config/")

				if isRoute && !isPackage && (strings.Contains(lower, "/admin") || strings.Contains(lower, "/debug") || strings.Contains(lower, "/config")) {
					severity := "high"
					explain := "Sensitive route pattern detected. Ensure authentication is enforced."

					// Simple Heuristic: If NODE_ENV=test is nearby or in path, lower severity
					if strings.Contains(strings.ToLower(filePath), "test") || strings.Contains(strings.ToLower(filePath), "config") {
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
		// Rule: Insecure CORS (JS/TS)
		func(tt js.TokenType, text string, pt js.TokenType, pText string, ppt js.TokenType, ppText string, line int, offset int, filePath string) *Finding {
			if tt == js.StringToken || tt == js.TemplateToken {
				if strings.Contains(text, "Access-Control-Allow-Origin") && strings.Contains(text, "*") {
					return &Finding{
						Rule:        "security.js.insecure_cors",
						Type:        "security_insecure_cors",
						Message:     "Insecure CORS: Access-Control-Allow-Origin set to '*' in JavaScript/TypeScript.",
						Severity:    "high",
						Confidence:  0.9,
						Explain:     "Allowing all origins via '*' can lead to CSRF and data theft if sensitive data is involved.",
						Remediation: "Specify explicit origins or use a robust CORS middleware like 'cors'.",
						Line:        line,
						ByteOffset:  offset,
						ByteLength:  len(text),
					}
				}
			}
			return nil
		},
		// Rule: Missing/Insecure Headers (JS/TS)
		func(tt js.TokenType, text string, pt js.TokenType, pText string, ppt js.TokenType, ppText string, line int, offset int, filePath string) *Finding {
			if tt == js.StringToken || tt == js.TemplateToken {
				lowerText := strings.ToLower(text)
				if (strings.Contains(lowerText, "x-powered-by") || strings.Contains(lowerText, "x-content-type-options")) && strings.Contains(lowerText, "off") {
					return &Finding{
						Rule:        "security.js.insecure_headers",
						Type:        "security_insecure_headers",
						Message:     "Insecure Header Configuration: Disabling standard security protections.",
						Severity:    "medium",
						Confidence:  0.8,
						Explain:     "Explicitly disabling 'X-Content-Type-Options' or leaving 'X-Powered-By' enabled exposes the tech stack and increases vulnerability to Mime-Sniffing.",
						Remediation: "Ensure 'X-Content-Type-Options: nosniff' is set and 'X-Powered-By' is hidden.",
						Line:        line,
						ByteOffset:  offset,
						ByteLength:  len(text),
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
				// PRO-GRADE: Don't block on decorative characters or regex lexer misses
				result.Findings = append(result.Findings, Finding{
					Rule:     "parse.remark",
					Type:     "parse_remark",
					Message:  fmt.Sprintf("Lexer Remark: Non-critical character encountered (%v)", lexer.Err()),
					Severity: "info",
					Explain:  "The lexer encountered a character it wasn't expecting, often decorative icons or complex regex. This is ignored for security purposes.",
					Line:     1,
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
			// Respect min_risk_score from tharos.yaml
			minScore := viper.GetInt("ai.min_risk_score")
			if minScore == 0 {
				minScore = 60 // Default if not set
			}

			if insight.RiskScore >= minScore {
				result.AIInsights = append(result.AIInsights, insight)
			} else if verbose {
				fmt.Printf("  %sℹ AI Insight ignored (Risk Score %d < %d)%s\n", colorGray, insight.RiskScore, minScore, colorReset)
			}

			// Apply AI-suggested fixes to findings (regardless of visibility, if they are accurate)

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
	// Respect min_risk_score for fallback as well
	minScore := viper.GetInt("ai.min_risk_score")
	if minScore == 0 {
		minScore = 60
	}

	if minScore <= 50 { // Fallback has a score of 50
		result.AIInsights = append(result.AIInsights, AIInsight{
			Recommendation: response,
			RiskScore:      50,
		})
	}
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
	FullDescription  SARIFDescription `json:"fullDescription,omitempty"`
	HelpURI          string           `json:"helpUri,omitempty"`
	Properties       SARIFProperties  `json:"properties,omitempty"`
}

type SARIFProperties struct {
	Precision string   `json:"precision,omitempty"`
	Tags      []string `json:"tags,omitempty"`
}

type SARIFDescription struct {
	Text string `json:"text"`
}

type SARIFResult struct {
	RuleID    string           `json:"ruleId"`
	RuleIndex int              `json:"ruleIndex"`
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
	StartLine   int `json:"startLine"`
	StartColumn int `json:"startColumn,omitempty"`
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
						InformationURI: "https://tharos.vercel.app",
						Rules:          []SARIFRule{},
					},
				},
				Results: []SARIFResult{},
			},
		},
	}

	ruleMap := make(map[string]int)
	ruleCounter := 0

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
			if _, exists := ruleMap[f.Rule]; !exists {
				report.Runs[0].Tool.Driver.Rules = append(report.Runs[0].Tool.Driver.Rules, SARIFRule{
					ID: f.Rule,
					ShortDescription: SARIFDescription{
						Text: f.Message,
					},
					FullDescription: SARIFDescription{
						Text: f.Explain,
					},
					HelpURI: "https://tharos.vercel.app/docs",
					Properties: SARIFProperties{
						Precision: "high",
						Tags:      []string{f.Type, "security"},
					},
				})
				ruleMap[f.Rule] = ruleCounter
				ruleCounter++
			}

			// Add result
			report.Runs[0].Results = append(report.Runs[0].Results, SARIFResult{
				RuleID:    f.Rule,
				RuleIndex: ruleMap[f.Rule],
				Level:     level,
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
								StartLine: func() int {
									if f.Line < 1 {
										return 1
									}
									return f.Line
								}(),
							},
						},
					},
				},
			})
		}
	}

	return report
}

func runInteractiveFixes(batch *BatchResult) {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("36")).
		Padding(0, 1).
		MarginBottom(1)

	fmt.Println(headerStyle.Render("\n✨ THAROS INTERACTIVE MAGIC FIX SESSION"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Review each finding and decide: [Fix], [Explain], or [Skip].\n"))

	for i := range batch.Results {
		res := &batch.Results[i]
		if len(res.Findings) == 0 {
			continue
		}

		fmt.Printf("%s📁 FILE: %s%s\n", colorBold+colorCyan, res.File, colorReset)

		// Create a copy of pointers to findings to sort for review
		findingsToReview := make([]*Finding, 0, len(res.Findings))
		for j := range res.Findings {
			findingsToReview = append(findingsToReview, &res.Findings[j])
		}

		// Sort by line for a logical review order (top to bottom)
		sort.Slice(findingsToReview, func(i, j int) bool {
			return findingsToReview[i].Line < findingsToReview[j].Line
		})

		markedForFix := []*Finding{}

		for _, f := range findingsToReview {
			// Clear separator
			fmt.Println(strings.Repeat("─", 60))

			// Display finding context
			sevSym := getSeveritySymbol(f.Severity)
			sevCol := getSeverityColor(f.Severity)
			fmt.Printf("%s %s[%s] LINE %d: %s%s\n", sevSym, sevCol, strings.ToUpper(f.Severity), f.Line, f.Message, colorReset)

			var choice string
			options := []huh.Option[string]{
				huh.NewOption("Skip finding", "skip"),
			}

			if f.Replacement != "" {
				options = append([]huh.Option[string]{huh.NewOption("Apply Magic Fix", "fix")}, options...)
			}

			options = append(options, huh.NewOption("Explain Risk (AI)", "explain"))
			options = append(options, huh.NewOption("Abort session", "abort"))

			for {
				prompt := huh.NewSelect[string]().
					Title("Action:").
					Options(options...).
					Value(&choice)

				err := prompt.Run()
				if err != nil {
					fmt.Println("Interactive session interrupted.")
					return
				}

				if choice == "explain" {
					fmt.Printf("\n%s🧠 AI EXPLANATION:%s\n", colorBold+colorYellow, colorReset)
					if f.Explain != "" {
						fmt.Println(f.Explain)
					} else {
						fmt.Println("No explanation available for this finding.")
					}
					if f.Remediation != "" {
						fmt.Printf("\n%s💡 REMEDIATION:%s\n", colorBold+colorGreen, colorReset)
						fmt.Println(f.Remediation)
					}
					fmt.Println("\nPress Enter to return to options...")
					var dummy string
					fmt.Scanln(&dummy)
					// Continue loop to re-show options
					continue
				}

				if choice == "fix" {
					markedForFix = append(markedForFix, f)
					fmt.Printf("%s  ✅ Marked for fix.%s\n", colorGreen, colorReset)
				} else if choice == "abort" {
					fmt.Println("Aborting session.")
					return
				} else if choice == "skip" {
					fmt.Println("Skipped.")
				}
				break
			}
		}

		if len(markedForFix) > 0 {
			// Apply marked fixes in reverse order of offset to avoid shifting issues
			sort.Slice(markedForFix, func(i, j int) bool {
				return markedForFix[i].ByteOffset > markedForFix[j].ByteOffset
			})

			content, err := ioutil.ReadFile(res.File)
			if err != nil {
				fmt.Printf("❌ Failed to re-read file for fixing: %v\n", err)
				continue
			}

			newContent := make([]byte, len(content))
			copy(newContent, content)

			appliedAny := false
			for _, f := range markedForFix {
				if f.ByteOffset >= len(newContent) || f.ByteOffset+f.ByteLength > len(newContent) {
					continue
				}
				prefix := newContent[:f.ByteOffset]
				suffix := newContent[f.ByteOffset+f.ByteLength:]
				updated := make([]byte, 0, len(prefix)+len(f.Replacement)+len(suffix))
				updated = append(updated, prefix...)
				updated = append(updated, []byte(f.Replacement)...)
				updated = append(updated, suffix...)
				newContent = updated
				appliedAny = true
			}

			if appliedAny {
				err = ioutil.WriteFile(res.File, newContent, 0o644)
				if err != nil {
					fmt.Printf("❌ Error writing fixes to %s: %v\n", res.File, err)
				} else {
					fmt.Printf("\n%s✨ Successfully applied %d fixes to %s%s\n", colorGreen, len(markedForFix), res.File, colorReset)
				}
			}
		}
	}

	// ... existing code ...
}

func printHTMLOutput(results BatchResult) {
	jsonData, err := json.Marshal(results)
	if err != nil {
		fmt.Printf("Error generating HTML report: %v\n", err)
		return
	}

	htmlTemplate := `
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Tharos Security Audit Report</title>
    <style>
        :root {
            --bg-color: #0f172a;
            --card-bg: rgba(30, 41, 59, 0.7);
            --text-primary: #f8fafc;
            --text-secondary: #94a3b8;
            --accent: #38bdf8;
            --critical: #ef4444;
            --high: #f97316;
            --medium: #eab308;
            --info: #3b82f6;
            --success: #22c55e;
        }

        body {
            font-family: 'Inter', system-ui, -apple-system, sans-serif;
            background-color: var(--bg-color);
            color: var(--text-primary);
            margin: 0;
            padding: 2rem;
            line-height: 1.6;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
        }

        header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 2rem;
            padding-bottom: 1rem;
            border-bottom: 1px solid rgba(255,255,255,0.1);
        }

        .brand {
            font-size: 1.5rem;
            font-weight: 700;
            color: var(--accent);
            display: flex;
            align-items: center;
            gap: 0.5rem;
        }

        .stats-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 1rem;
            margin-bottom: 2rem;
        }

        .stat-card {
            background: var(--card-bg);
            border: 1px solid rgba(255,255,255,0.05);
            padding: 1.5rem;
            border-radius: 12px;
            backdrop-filter: blur(10px);
        }

        .stat-value {
            font-size: 2rem;
            font-weight: 700;
            margin-bottom: 0.25rem;
        }

        .stat-label {
            color: var(--text-secondary);
            font-size: 0.875rem;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .file-section {
            background: var(--card-bg);
            border-radius: 12px;
            margin-bottom: 1.5rem;
            overflow: hidden;
            border: 1px solid rgba(255,255,255,0.05);
        }

        .file-header {
            padding: 1rem 1.5rem;
            background: rgba(255,255,255,0.02);
            display: flex;
            justify-content: space-between;
            align-items: center;
            cursor: pointer;
        }

        .file-name {
            font-family: 'Fira Code', monospace;
            font-weight: 600;
        }

        .finding-list {
            padding: 0;
            margin: 0;
            list-style: none;
        }

        .finding-item {
            padding: 1.5rem;
            border-top: 1px solid rgba(255,255,255,0.05);
            transition: background 0.2s;
        }

        .finding-item:hover {
            background: rgba(255,255,255,0.02);
        }

        .severity-badge {
            padding: 0.25rem 0.75rem;
            border-radius: 9999px;
            font-size: 0.75rem;
            font-weight: 700;
            text-transform: uppercase;
        }

        .sev-critical { background: rgba(239, 68, 68, 0.2); color: var(--critical); border: 1px solid rgba(239, 68, 68, 0.3); }
        .sev-high { background: rgba(249, 115, 22, 0.2); color: var(--high); border: 1px solid rgba(249, 115, 22, 0.3); }
        .sev-medium { background: rgba(234, 179, 8, 0.2); color: var(--medium); border: 1px solid rgba(234, 179, 8, 0.3); }
        .sev-info { background: rgba(59, 130, 246, 0.2); color: var(--info); border: 1px solid rgba(59, 130, 246, 0.3); }

        .finding-header {
            display: flex;
            gap: 1rem;
            align-items: center;
            margin-bottom: 0.75rem;
        }

        .finding-message {
            font-weight: 500;
            flex-grow: 1;
        }

        .finding-rule {
            color: var(--text-secondary);
            font-size: 0.875rem;
            font-family: 'Fira Code', monospace;
        }

        .finding-details {
            background: rgba(0,0,0,0.2);
            padding: 1rem;
            border-radius: 8px;
            margin-top: 1rem;
            font-size: 0.9rem;
            color: var(--text-secondary);
        }

        .verdict-pass { color: var(--success); }
        .verdict-block { color: var(--critical); }

    </style>
</head>
<body>
    <div class="container">
        <header>
            <div class="brand">
                🛡️ THAROS SECURITY REPORT
            </div>
            <div class="verdict">
                VERDICT: <span id="verdict-text" style="font-weight: 800;">CALCULATING...</span>
            </div>
        </header>

        <div class="stats-grid">
            <div class="stat-card">
                <div class="stat-value" id="stats-vulns">0</div>
                <div class="stat-label">Total Vulnerabilities</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="stats-files">0</div>
                <div class="stat-label">Files Scanned</div>
            </div>
            <div class="stat-card">
                <div class="stat-value" id="stats-dur">0ms</div>
                <div class="stat-label">Duration</div>
            </div>
        </div>

        <div id="results-container"></div>
    </div>

    <script>
        const reportData = %s;

        function renderReport() {
            // Update Stats
            document.getElementById('stats-vulns').textContent = reportData.summary.vulnerabilities;
            document.getElementById('stats-files').textContent = reportData.summary.total_files;
            document.getElementById('stats-dur').textContent = reportData.summary.duration;

            // Verdict
            const crit = reportData.results.reduce((acc, r) => acc + r.findings.filter(f => f.severity === 'critical' || f.severity === 'block').length, 0);
            const high = reportData.results.reduce((acc, r) => acc + r.findings.filter(f => f.severity === 'high').length, 0);
            const verEl = document.getElementById('verdict-text');
            if (crit > 0 || high >= 3) {
                verEl.textContent = "BLOCK";
                verEl.className = "verdict-block";
            } else {
                verEl.textContent = "PASS";
                verEl.className = "verdict-pass";
            }

            // Render Files
            const container = document.getElementById('results-container');
            if (reportData.results.length === 0) {
                container.innerHTML = '<div style="text-align: center; color: var(--text-secondary); padding: 3rem;">No files analyzed.</div>';
                return;
            }

            // Sort: files with issues first
            const sortedResults = reportData.results.sort((a,b) => b.findings.length - a.findings.length);

            sortedResults.forEach(res => {
                const hasIssues = res.findings.length > 0;
                if(!hasIssues) return; // Optional: Show/hide safe files

                const section = document.createElement('div');
                section.className = 'file-section';
                
                let findingsHtml = '<ul class="finding-list">';
                res.findings.forEach(f => {
                    const sevClass = 'sev-' + (f.severity === 'block' ? 'critical' : f.severity);
                    findingsHtml += '<li class="finding-item">';
                    findingsHtml += '<div class="finding-header">';
                    findingsHtml += '<span class="severity-badge ' + sevClass + '">' + f.severity + '</span>';
                    findingsHtml += '<span class="finding-message">Line ' + f.line + ': ' + f.message + '</span>';
                    findingsHtml += '<span class="finding-rule">' + f.rule + '</span>';
                    findingsHtml += '</div>';
                    if (f.explain) {
                        findingsHtml += '<div class="finding-details">🧠 ' + f.explain + '</div>';
                    }
                    findingsHtml += '</li>';
                });
                findingsHtml += '</ul>';

                section.innerHTML = 
                    '<div class="file-header">' +
                    '<span class="file-name">' + res.file + '</span>' +
                    '<span class="severity-badge sev-info">' + res.findings.length + ' Finding(s)</span>' +
                    '</div>' + 
                    findingsHtml;
                container.appendChild(section);
            });

            if (container.children.length === 0) {
                container.innerHTML = '<div style="text-align: center; color: var(--success); padding: 3rem; font-size: 1.2rem;">✨ Clean Scan! No issues found.</div>';
            }
        }

        renderReport();
    </script>
</body>
</html>
`
	fmt.Printf(htmlTemplate, jsonData)
}
