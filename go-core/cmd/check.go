package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Run Tharos policy checks on staged files",
	Long:  `Scans all files staged for commit (git add) for security vulnerabilities and policy violations.`,
	Run:   runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	repo, err := git.PlainOpen(dir)
	if err != nil {
		fmt.Printf("❌ Error: Not a git repository\n")
		os.Exit(1)
	}

	worktree, err := repo.Worktree()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	status, err := worktree.Status()
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
		os.Exit(1)
	}

	var stagedFiles []string
	excludes := viper.GetStringSlice("exclude")
	for path, s := range status {
		if s.Staging == git.Added || s.Staging == git.Modified || s.Staging == git.Renamed {
			// Skip excluded files/directories
			isExcluded := false
			for _, pattern := range excludes {
				if strings.Contains(path, pattern) {
					isExcluded = true
					break
				}
			}
			if isExcluded {
				continue
			}

			ext := strings.ToLower(filepath.Ext(path))
			if ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx" ||
				ext == ".go" || ext == ".py" {
				stagedFiles = append(stagedFiles, path)
			}
		}
	}

	if len(stagedFiles) == 0 {
		fmt.Println("No relevant files staged for commit.")
		return
	}

	fmt.Println("🛡️ Tharos is analyzing your intent...")

	start := time.Now()
	var results []AnalysisResult
	for _, file := range stagedFiles {
		results = append(results, analyzeFile(file, aiEnabled))
	}
	duration := time.Since(start)

	totalVulns := 0
	for _, r := range results {
		totalVulns += len(r.Findings)
	}

	summary := BatchResult{
		Results: results,
		Summary: ScanSummary{
			TotalFiles:      len(results),
			Vulnerabilities: totalVulns,
			Duration:        duration.String(),
			DurationMs:      duration.Milliseconds(),
		},
	}

	if jsonOutput {
		printJSONOutput(summary)
	} else {
		printRichOutput(summary, verbose, fixMode)
	}

	// Exit with 1 if blocks exist
	for _, r := range results {
		for _, f := range r.Findings {
			if f.Severity == "block" {
				fmt.Println("\n🛑 Commit blocked by Tharos policy. Please fix the issues above.")
				os.Exit(1)
			}
		}
	}

	fmt.Println("\n✨ Tharos logic check passed! Proceeding...")
}
