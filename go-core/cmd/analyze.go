package cmd

import (
	"fmt"
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
}

func runAnalyze(cmd *cobra.Command, args []string) {
	path := args[0]

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
	if fixMode {
		applyFixes(output.Results)
	}

	isMachineReadable := jsonOutput || outputFormat == "json" || outputFormat == "sarif"

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
			if f.Severity == "critical" {
				criticalCount++
			} else if f.Severity == "high" {
				highCount++
			}
		}
	}

	// 🛑 Blocking Rule: ≥1 Critical OR ≥3 High
	if criticalCount >= 1 || highCount >= 3 {
		if !isMachineReadable {
			fmt.Printf("\n%s🛑 Commit Blocked:%s Detected %d Critical and %d High risk issues.\n", colorRed+colorBold, colorReset, criticalCount, highCount)
		}
		os.Exit(1)
	}
}
