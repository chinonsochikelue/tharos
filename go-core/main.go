package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"github.com/ollama/ollama/api"
	"github.com/tdewolff/parse/v2"
	"github.com/tdewolff/parse/v2/js"
	"google.golang.org/api/option"
)

type AnalysisResult struct {
	File       string      `json:"file"`
	Findings   []Finding   `json:"findings"`
	AIInsights []AIInsight `json:"ai_insights"`
}

type AIInsight struct {
	Recommendation string `json:"recommendation"`
	RiskScore      int    `json:"risk_score"`
	SuggestedFix   string `json:"suggested_fix,omitempty"`
}

type Finding struct {
	Type     string `json:"type"`
	Message  string `json:"message"`
	Severity string `json:"severity"` // "warning", "block"
	Line     int    `json:"line"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: tharos core <command> [args]")
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "analyze":
		if len(os.Args) < 3 {
			fmt.Println("Usage: tharos core analyze <file_path>")
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

	// Simulated Policy Severity
	policySeverity := "warning"
	if _, err := os.Stat("tharos.yaml"); err == nil {
		policySeverity = "block"
	}

	// 1. Precise AST Analysis
	analyzeAST(content, &result, policySeverity)

	// 2. Integration with AI Providers (Ollama/Gemini/Groq)
	if len(result.Findings) > 0 {
		aiRes := getAIInsight(string(content), result.Findings)
		if len(aiRes.AIInsights) > 0 {
			result.AIInsights = append(result.AIInsights, aiRes.AIInsights...)
		}
	}

	return result
}

func analyzeAST(content []byte, result *AnalysisResult, defaultSeverity string) {
	input := parse.NewInputBytes(content)
	lexer := js.NewLexer(input)

	// We need to track line manually if lexer offset is tricky
	currentLine := 1

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

		// Update line count based on consumed data
		currentLine += strings.Count(string(data), "\n")

		switch tt {
		case js.IdentifierToken:
			id := string(data)
			if id == "eval" || id == "exec" {
				result.Findings = append(result.Findings, Finding{
					Type:     "security",
					Message:  fmt.Sprintf("Dangerous function '%s' detected via AST analysis.", id),
					Severity: "block",
					Line:     currentLine,
				})
			}
			lowerId := strings.ToLower(id)
			if strings.Contains(lowerId, "api_key") || strings.Contains(lowerId, "secret") || strings.Contains(lowerId, "password") {
				// Semantic check: If this is process.env.STUFF, it's actually GOOD practice
				// In a lexer, we can check the previous token if we track it.
				// For now, let's just make identifier-only findings a WARNING instead of a BLOCK
				// because naming a variable correctly is GOOD, but hardcoding is BAD.
				result.Findings = append(result.Findings, Finding{
					Type:     "security_warning",
					Message:  fmt.Sprintf("Identifier '%s' might contain sensitive data. Ensure it is sourced from environment variables.", id),
					Severity: "warning",
					Line:     currentLine,
				})
			}
		case js.StringToken:
			val := string(data)
			if len(val) > 35 && (strings.Contains(val, "-") || strings.Contains(val, "_")) {
				result.Findings = append(result.Findings, Finding{
					Type:     "security",
					Message:  "Suspiciously long string literal detected; possible hardcoded token.",
					Severity: defaultSeverity,
					Line:     currentLine,
				})
			}
		case js.CommentToken:
			comment := string(data)
			if strings.Contains(comment, "TODO") || strings.Contains(comment, "FIXME") {
				result.Findings = append(result.Findings, Finding{
					Type:     "code_smell",
					Message:  "Static analysis found unresolved TODO in comment.",
					Severity: "warning",
					Line:     currentLine,
				})
			}
		}
	}
}

func getAIInsight(code string, findings []Finding) AnalysisResult {
	res := AnalysisResult{AIInsights: []AIInsight{}}

	// Try Ollama (Local Privacy)
	insightStr := getOllamaInsight(code, findings)
	if !strings.Contains(insightStr, "AI Insight Unavailable") {
		parseAIResponse(insightStr, &res)
		return res
	}

	// Try Managed AI (Default Easy Experience)
	insightStr = getManagedInsight(code, findings)
	if !strings.Contains(insightStr, "AI Insight Unavailable") {
		parseAIResponse(insightStr, &res)
		return res
	}

	// Fallback 1: Gemini (Personal)
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey != "" {
		insightStr = getGeminiInsight(code, findings, geminiKey)
		if !strings.Contains(insightStr, "AI Insight Unavailable") {
			parseAIResponse(insightStr, &res)
			return res
		}
	}

	// Fallback 2: Groq (Personal)
	groqKey := os.Getenv("GROQ_API_KEY")
	if groqKey != "" {
		insightStr = getGroqInsight(code, findings, groqKey)
		if !strings.Contains(insightStr, "AI Insight Unavailable") {
			parseAIResponse(insightStr, &res)
			return res
		}
	}

	return res
}

func getManagedInsight(code string, findings []Finding) string {
	// In a real scenario, this would call our hosted cloud endpoint.
	// For this implementation, we'll simulate a call to a hosted Gemini endpoint
	// providing a "Public" Tharos API Key experience.

	tharosPublicApiKey := os.Getenv("THAROS_MANAGED_KEY")
	if tharosPublicApiKey == "" {
		// Simulation: If no key is provided, we simulate a gracefully failing cloud call
		return "AI Insight Unavailable: Managed Service requires THAROS_MANAGED_KEY or local Ollama."
	}

	// For simulation purposes, we'll use the Gemini logic but with the Managed Key
	return getGeminiInsight(code, findings, tharosPublicApiKey)
}

func getOllamaInsight(code string, findings []Finding) string {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		return "AI Insight Unavailable: Could not connect to Ollama."
	}

	ctx := context.Background()
	prompt := generatePrompt(code, findings)

	req := &api.GenerateRequest{
		Model:  "llama3",
		Prompt: prompt,
		Stream: nil,
	}

	var aiResponse string
	respFunc := func(resp api.GenerateResponse) error {
		aiResponse = resp.Response
		return nil
	}

	err = client.Generate(ctx, req, respFunc)
	if err != nil {
		return fmt.Sprintf("AI Insight Unavailable: (Ollama %v)", err)
	}

	return aiResponse
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
	// Groq is OpenAI-compatible, we'll use a direct HTTP request to avoid extra fat dependencies
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
	// AI might return Markdown blocks or extra text, try to extract JSON
	jsonStart := strings.Index(response, "{")
	jsonEnd := strings.LastIndex(response, "}")

	if jsonStart != -1 && jsonEnd != -1 && jsonEnd > jsonStart {
		jsonStr := response[jsonStart : jsonEnd+1]
		var insight AIInsight
		if err := json.Unmarshal([]byte(jsonStr), &insight); err == nil {
			result.AIInsights = append(result.AIInsights, insight)
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

Code Context:
%s

Issues Found:
`
	for _, f := range findings {
		prompt += fmt.Sprintf("- [%s] %s (Line %d)\n", f.Type, f.Message, f.Line)
	}

	return fmt.Sprintf(prompt, code)
}
