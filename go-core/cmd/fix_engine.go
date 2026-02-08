package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// FixSuggestion represents an AI-generated fix for a finding
type FixSuggestion struct {
	Line        int     `json:"line"`
	Original    string  `json:"original"`
	Replacement string  `json:"replacement"`
	Explanation string  `json:"explanation"`
	Confidence  float64 `json:"confidence"`
}

// MultiFileFix represents changes that span multiple files
type MultiFileFix struct {
	FilePath string `json:"file_path"`
	Content  string `json:"content"`
	Action   string `json:"action"` // "create", "modify", "append"
}

// FixPlan represents a complete fix strategy for a finding
type FixPlan struct {
	FindingRule       string          `json:"finding_rule"`
	PrimaryFixes      []FixSuggestion `json:"primary_fixes"`
	AdditionalChanges []MultiFileFix  `json:"additional_changes"`
	OverallConfidence float64         `json:"overall_confidence"`
	RequiresManual    bool            `json:"requires_manual"`
}

// BackupManager handles file backups and rollbacks
type BackupManager struct {
	BackupDir string
	Timestamp string
}

// NewBackupManager creates a new backup manager
func NewBackupManager() *BackupManager {
	timestamp := time.Now().Format("20060102_150405")
	backupDir := filepath.Join(".tharos-backup", timestamp)
	return &BackupManager{
		BackupDir: backupDir,
		Timestamp: timestamp,
	}
}

// BackupFile creates a backup of a file before modification
func (bm *BackupManager) BackupFile(filePath string) error {
	// Ensure backup directory exists
	if err := os.MkdirAll(bm.BackupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	// Read original file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file for backup: %w", err)
	}

	// Create backup file path (preserve directory structure)
	backupPath := filepath.Join(bm.BackupDir, filePath)
	backupDir := filepath.Dir(backupPath)
	if err := os.MkdirAll(backupDir, 0755); err != nil {
		return fmt.Errorf("failed to create backup subdirectory: %w", err)
	}

	// Write backup
	if err := ioutil.WriteFile(backupPath, content, 0644); err != nil {
		return fmt.Errorf("failed to write backup: %w", err)
	}

	return nil
}

// Rollback restores all files from the backup
func (bm *BackupManager) Rollback() error {
	if _, err := os.Stat(bm.BackupDir); os.IsNotExist(err) {
		return fmt.Errorf("backup directory not found: %s", bm.BackupDir)
	}

	restoredCount := 0
	err := filepath.Walk(bm.BackupDir, func(backupPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		// Get original file path
		relPath, err := filepath.Rel(bm.BackupDir, backupPath)
		if err != nil {
			return err
		}

		// Read backup content
		content, err := ioutil.ReadFile(backupPath)
		if err != nil {
			return fmt.Errorf("failed to read backup file: %w", err)
		}

		// Restore original file
		if err := ioutil.WriteFile(relPath, content, 0644); err != nil {
			return fmt.Errorf("failed to restore file: %w", err)
		}

		restoredCount++
		return nil
	})

	if err != nil {
		return err
	}

	fmt.Printf("✅ Restored %d files from backup\n", restoredCount)
	return nil
}

// CleanupBackup removes the backup directory
func (bm *BackupManager) CleanupBackup() error {
	return os.RemoveAll(bm.BackupDir)
}

// ApplyFix applies a single fix to a file
func ApplyFix(filePath string, fix FixSuggestion) error {
	// Read file
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	lines := strings.Split(string(content), "\n")

	// Validate line number
	if fix.Line < 1 || fix.Line > len(lines) {
		return fmt.Errorf("invalid line number: %d (file has %d lines)", fix.Line, len(lines))
	}

	// Apply fix (1-indexed to 0-indexed)
	lineIdx := fix.Line - 1
	originalLine := lines[lineIdx]

	// Verify original content matches (safety check)
	if !strings.Contains(originalLine, strings.TrimSpace(fix.Original)) {
		return fmt.Errorf("original content mismatch at line %d:\nExpected substring: %s\nActual line: %s",
			fix.Line, fix.Original, originalLine)
	}

	// Replace
	lines[lineIdx] = strings.Replace(originalLine, strings.TrimSpace(fix.Original), fix.Replacement, 1)

	// Write back
	newContent := strings.Join(lines, "\n")
	if err := ioutil.WriteFile(filePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("failed to write fixed file: %w", err)
	}

	return nil
}

// ApplyMultiFileFix applies changes to additional files (e.g., creating .env)
func ApplyMultiFileFix(fix MultiFileFix) error {
	switch fix.Action {
	case "create":
		// Check if file already exists
		if _, err := os.Stat(fix.FilePath); err == nil {
			return fmt.Errorf("file already exists: %s", fix.FilePath)
		}
		// Create file
		return ioutil.WriteFile(fix.FilePath, []byte(fix.Content), 0644)

	case "append":
		// Append to existing file
		f, err := os.OpenFile(fix.FilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()
		_, err = f.WriteString(fix.Content)
		return err

	case "modify":
		// Replace entire file content
		return ioutil.WriteFile(fix.FilePath, []byte(fix.Content), 0644)

	default:
		return fmt.Errorf("unknown action: %s", fix.Action)
	}
}

// GenerateFixPlanFromAI generates a fix plan using AI
func GenerateFixPlanFromAI(finding Finding, codeContext string) (*FixPlan, error) {
	// Construct AI prompt for fix generation
	prompt := fmt.Sprintf(`You are a security code remediation expert. Generate a precise fix for this security finding.

Finding:
- Rule: %s
- Message: %s
- Severity: %s
- Line: %d

Code Context:
%s

Generate a fix in this EXACT JSON format (no markdown, no extra text):
{
  "finding_rule": "%s",
  "primary_fixes": [
    {
      "line": %d,
      "original": "exact code to replace",
      "replacement": "fixed code",
      "explanation": "why this fix works",
      "confidence": 0.95
    }
  ],
  "additional_changes": [],
  "overall_confidence": 0.95,
  "requires_manual": false
}

CRITICAL: Return ONLY valid JSON. No markdown code blocks, no explanations outside JSON.`,
		finding.Rule,
		finding.Message,
		finding.Severity,
		finding.Line,
		codeContext,
		finding.Rule,
		finding.Line,
	)

	// Get AI response
	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		groqKey := os.Getenv("GROQ_API_KEY")
		if groqKey == "" {
			return nil, fmt.Errorf("no AI API key found (GEMINI_API_KEY or GROQ_API_KEY)")
		}
		return getFixFromGroq(prompt, groqKey)
	}
	return getFixFromGemini(prompt, geminiKey)
}

// getFixFromGemini gets fix plan from Gemini
func getFixFromGemini(prompt string, apiKey string) (*FixPlan, error) {
	// Use existing Gemini client
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, err
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-2.0-flash")
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return nil, err
	}

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("no response from Gemini")
	}

	responseText := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Clean response (remove markdown if present)
	responseText = strings.TrimPrefix(responseText, "```json")
	responseText = strings.TrimPrefix(responseText, "```")
	responseText = strings.TrimSuffix(responseText, "```")
	responseText = strings.TrimSpace(responseText)

	// Parse JSON
	var fixPlan FixPlan
	if err := json.Unmarshal([]byte(responseText), &fixPlan); err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w\nResponse: %s", err, responseText)
	}

	return &fixPlan, nil
}

// getFixFromGroq gets fix plan from Groq
func getFixFromGroq(prompt string, apiKey string) (*FixPlan, error) {
	// Similar to getGroqInsight but returns FixPlan
	// Implementation similar to Gemini
	return nil, fmt.Errorf("Groq fix generation not yet implemented")
}
