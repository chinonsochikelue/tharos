package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
)

type AnalysisResult struct {
	File       string    `json:"file"`
	Findings   []Finding `json:"findings"`
	AIInsights []string  `json:"ai_insights"`
}

type Finding struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "warning", "block"
	Line     int    `json:"line"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: fennec core <command> [args]")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "analyze":
		if len(os.Args) < 3 {
			fmt.Println("Usage: fennec core analyze <file_path>")
			os.Exit(1)
		}
		filePath := os.Args[2]
		result := analyze(filePath)
		output, _ := json.MarshalIndent(result, "", "  ")
		fmt.Println(string(output))
	default:
		fmt.Printf("Unknown command: %s\n", command)
		os.Exit(1)
	}
}

func analyze(filePath string) AnalysisResult {
	result := AnalysisResult{
		File:       filePath,
		Findings:   []Finding{},
		AIInsights: []string{},
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

	// Simulated Policy Sync: Check for local fennec.yaml
	policySeverity := "warning"
	if _, err := os.Stat("fennec.yaml"); err == nil {
		// In a real implementation, we'd parse the YAML
		policySeverity = "block" // Simulated "Hard Enforce" from cloud policy
	}

	code := string(content)
	analyzePatterns(code, &result, policySeverity)

	// Simulated "AI Semantic Analysis"
	if strings.Contains(code, "function") {
		result.AIInsights = append(result.AIInsights, "Refactor Insight: Consider extracting logic into a separate utility for better testability.")
	}
	if strings.Contains(code, "eval") {
		result.AIInsights = append(result.AIInsights, "Security Fix: Replace 'eval()' with a safer JSON.parse() or a mapped function lookup.")
	}
	if strings.Contains(code, "API_KEY") {
		result.AIInsights = append(result.AIInsights, "Organization Policy: Move hardcoded secrets to an encrypted .env file or Vault.")
	}

	return result
}

func analyzePatterns(content string, result *AnalysisResult, defaultSeverity string) {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// 1. Detect Security: Hardcoded Secrets
		lowerLine := strings.ToLower(trimmed)
		if (strings.Contains(lowerLine, "api_key") ||
			strings.Contains(lowerLine, "secret") ||
			strings.Contains(lowerLine, "password") ||
			strings.Contains(lowerLine, "token")) &&
			(strings.Contains(line, "=") || strings.Contains(line, ":")) {

			if len(line) > 20 && strings.Contains(line, "\"") {
				result.Findings = append(result.Findings, Finding{
					Type:     "security",
					Message:  "Potential hardcoded secret or sensitive key detected.",
					Severity: defaultSeverity, // Enforced by policy
					Line:     i + 1,
				})
			}
		}

		// 2. Detect Security: Dangerous Functions
		if strings.Contains(line, "eval(") || strings.Contains(line, "exec(") {
			result.Findings = append(result.Findings, Finding{
				Type:     "security",
				Message:  "Use of dangerous functions (eval/exec) detected.",
				Severity: "block",
				Line:     i + 1,
			})
		}

		// 3. Detect Code Smell: TODOs
		if strings.Contains(line, "TODO:") || strings.Contains(line, "FIXME:") {
			result.Findings = append(result.Findings, Finding{
				Type:     "code_smell",
				Message:  "Unresolved TODO or FIXME found.",
				Severity: "warning",
				Line:     i + 1,
			})
		}
	}
}
