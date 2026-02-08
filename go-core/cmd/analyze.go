package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var strictMode bool

var analyzeCmd = &cobra.Command{
	Use:   "analyze [path]",
	Short: "Analyze code for security vulnerabilities",
	Long:  `Scan the specified file or directory for security issues and quality problems.`,
	Args:  cobra.ExactArgs(1),
	Run:   runAnalyze,
}

func init() {
	rootCmd.AddCommand(analyzeCmd)

	analyzeCmd.Flags().BoolVar(&aiEnabled, "ai", false, "force AI analysis for all files")
	analyzeCmd.Flags().BoolVar(&fixMode, "fix", false, "attempt to auto-fix issues")
	analyzeCmd.Flags().StringVar(&policyPath, "policy", "", "path to external policy file (YAML)")
	analyzeCmd.Flags().StringVar(&policyDir, "policy-dir", "policies", "directory for policy files")
	analyzeCmd.Flags().BoolVarP(&interactiveMode, "interactive", "i", false, "interactive review and fix mode")
	analyzeCmd.Flags().BoolVar(&strictMode, "strict", false, "fail on any non-info finding")
}

func runAnalyze(cmd *cobra.Command, args []string) {
	path := args[0]

	// Load external policies if flags are set
	loadExternalPolicies(policyPath, policyDir)

	start := time.Now()

	results := analyzePath(path, aiEnabled)
	duration := time.Since(start)

	totalVulns := 0
	for _, r := range results {
		totalVulns += len(r.Findings)
	}

	output := BatchResult{
		Results: results,
		Summary: ScanSummary{
			TotalFiles:      len(results),
			Vulnerabilities: totalVulns,
			Duration:        duration.String(),
			DurationMs:      duration.Milliseconds(),
		},
	}

	// Apply fixes if requested, regardless of output format
	if interactiveMode {
		runInteractiveFixes(&output)
	} else if fixMode {
		applyFixes(output.Results)
	}

	if jsonOutput || outputFormat == "json" {
		printJSONOutput(output)
	} else if outputFormat == "sarif" {
		printSARIFOutput(output.Results)
	} else if outputFormat == "html" {
		printHTMLOutput(output)
	} else {
		printRichOutput(output, verbose, fixMode)
	}

	// Exit Enforcement Logic
	failed := false

	// Count findings by severity
	criticalCount := 0
	highCount := 0
	mediumCount := 0
	lowCount := 0

	for _, r := range results {
		for _, f := range r.Findings {
			switch strings.ToLower(f.Severity) {
			case "block", "critical":
				criticalCount++
			case "high":
				highCount++
			case "medium", "warning":
				mediumCount++
			case "low", "info":
				lowCount++
			}
		}
	}

	// 1. Strict Mode: Fail on ANY non-info issue
	if strictMode {
		if criticalCount > 0 || highCount > 0 || mediumCount > 0 {
			if !jsonOutput && outputFormat != "sarif" && outputFormat != "html" {
				fmt.Printf("\n%s🛑 STRICT MODE: Failing build due to %d issues.%s\n", colorRed, criticalCount+highCount+mediumCount, colorReset)
			}
			failed = true
		}
	} else {
		// 2. Standard Mode: Fail on BLOCK/CRITICAL or HIGH
		if criticalCount > 0 || highCount > 0 {
			if !jsonOutput && outputFormat != "sarif" && outputFormat != "html" {
				fmt.Printf("\n%s🛑 BUILD FAILED: Security issues detected.%s\n", colorRed, colorReset)
			}
			failed = true
		}
	}

	if failed {
		os.Exit(1)
	}
}
