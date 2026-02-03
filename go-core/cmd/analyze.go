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

	if jsonOutput {
		printJSONOutput(output)
	} else {
		printRichOutput(output, verbose, fixMode)
	}

	// Exit with error code if blocking issues found
	if totalVulns > 0 {
		for _, r := range results {
			for _, f := range r.Findings {
				if f.Severity == "block" {
					os.Exit(1)
				}
			}
		}
	}
}
