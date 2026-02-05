package cmd

import (
	"os"
	"time"

	"github.com/spf13/cobra"
)

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
	} else {
		printRichOutput(output, verbose, fixMode)
	}

	// Exit with error code if blocking issues found
	criticalCount := 0
	highCount := 0
	for _, r := range results {
		for _, f := range r.Findings {
			if f.Severity == "critical" || f.Severity == "block" {
				criticalCount++
			} else if f.Severity == "high" {
				highCount++
			}
		}
	}

	// 🛑 Blocking Rule: ≥1 Critical OR ≥3 High
	if criticalCount >= 1 || highCount >= 3 {
		os.Exit(1)
	}
}
