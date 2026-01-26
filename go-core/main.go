package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/ollama/ollama/api"
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

	code := string(content)

	// Simulated Policy Sync
	policySeverity := "warning"
	if _, err := os.Stat("fennec.yaml"); err == nil {
		policySeverity = "block"
	}

	analyzePatterns(code, &result, policySeverity)

	// Integration with Local Ollama
	if len(result.Findings) > 0 {
		insight := getAIInsight(code, result.Findings)
		if insight != "" {
			result.AIInsights = append(result.AIInsights, insight)
		}
	}

	return result
}

func getAIInsight(code string, findings []Finding) string {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return "AI Insight Unavailable: Could not connect to Ollama."
	}

	ctx := context.Background()

	prompt := fmt.Sprintf("Analyze the following code snippet and findings. Provide a single, concise expert recommendation for the developer.\n\nCode:\n%s\n\nFindings:\n", code)
	for _, f := range findings {
		prompt += fmt.Sprintf("- %s: %s (Line %d)\n", f.Type, f.Message, f.Line)
	}

	req := &api.GenerateRequest{
		Model:  "llama3", // Default to llama3, can be configured
		Prompt: prompt,
		Stream: nil, // We want the full response
	}

	var aiResponse string
	respFunc := func(resp api.GenerateResponse) error {
		aiResponse = resp.Response
		return nil
	}

	err = client.Generate(ctx, req, respFunc)
	if err != nil {
		return fmt.Sprintf("AI Insight Unavailable: Ensure Ollama is running with 'llama3' model (%v)", err)
	}

	return aiResponse
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
					Severity: defaultSeverity,
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
