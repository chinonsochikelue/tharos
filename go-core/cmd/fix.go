package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"
)

var (
	autoFix             bool
	confidenceThreshold float64
	rollbackTimestamp   string
)

var fixCmd = &cobra.Command{
	Use:   "fix [path]",
	Short: "Interactively fix security vulnerabilities with AI",
	Long: `Analyze code and apply AI-generated fixes for security vulnerabilities.

Examples:
  tharos fix src/api/login.ts          # Interactive fix mode for one file
  tharos fix . --auto                  # Auto-fix all high-confidence issues
  tharos fix --rollback 20240208_153045  # Rollback to backup`,
	Args: cobra.MaximumNArgs(1),
	Run:  runFix,
}

func init() {
	rootCmd.AddCommand(fixCmd)
	fixCmd.Flags().BoolVar(&autoFix, "auto", false, "automatically apply high-confidence fixes")
	fixCmd.Flags().Float64Var(&confidenceThreshold, "confidence", 0.9, "minimum confidence for auto-fix")
	fixCmd.Flags().StringVar(&rollbackTimestamp, "rollback", "", "rollback to specific backup timestamp")
}

func runFix(cmd *cobra.Command, args []string) {
	// Handle rollback
	if rollbackTimestamp != "" {
		handleRollback(rollbackTimestamp)
		return
	}

	// Determine scan path
	scanPath := "."
	if len(args) > 0 {
		scanPath = args[0]
	}

	// Create backup manager
	backupMgr := NewBackupManager()
	fmt.Printf("🛡️ Creating backup at: %s\n", backupMgr.BackupDir)

	// Analyze code
	fmt.Printf("🔍 Analyzing %s...\n\n", scanPath)
	results := analyzePath(scanPath, true) // Enable AI for fix generation

	// Count fixable findings
	totalFindings := 0
	fixableFindings := 0
	for _, res := range results {
		totalFindings += len(res.Findings)
		for _, f := range res.Findings {
			if f.Replacement != "" || f.Severity == "high" || f.Severity == "critical" {
				fixableFindings++
			}
		}
	}

	if totalFindings == 0 {
		fmt.Println("✅ No security issues found!")
		return
	}

	fmt.Printf("📊 Found %d issues, %d potentially fixable\n\n", totalFindings, fixableFindings)

	// Auto-fix mode
	if autoFix {
		runAutoFix(results, backupMgr)
		return
	}

	// Interactive mode
	runInteractiveFix(results, backupMgr)
}

func runInteractiveFix(results []AnalysisResult, backupMgr *BackupManager) {
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("36")).
		Padding(0, 1).
		MarginBottom(1)

	fmt.Println(headerStyle.Render("✨ THAROS INTERACTIVE FIX SESSION"))
	fmt.Println(lipgloss.NewStyle().Foreground(lipgloss.Color("245")).Render("Review each finding and decide: [Fix], [Skip], or [Explain].\n"))

	fixedCount := 0
	skippedCount := 0

	for _, res := range results {
		if len(res.Findings) == 0 {
			continue
		}

		fmt.Printf("%s📁 FILE: %s%s\n", colorBold+colorCyan, res.File, colorReset)

		for _, finding := range res.Findings {
			// Display finding
			sevSym := getSeveritySymbol(finding.Severity)
			sevCol := getSeverityColor(finding.Severity)
			fmt.Println(strings.Repeat("─", 60))
			fmt.Printf("%s %s[%s] LINE %d: %s%s\n", sevSym, sevCol, strings.ToUpper(finding.Severity), finding.Line, finding.Message, colorReset)

			// Get code context
			codeContext := getCodeContext(res.File, finding.Line, 3)

			// Generate AI fix with spinner
			fmt.Printf("\n%s🧠 Generating AI fix...%s ", colorYellow, colorReset)

			// Simple spinner animation
			spinChars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
			spinDone := make(chan bool)
			go func() {
				i := 0
				for {
					select {
					case <-spinDone:
						fmt.Print("\r" + strings.Repeat(" ", 50) + "\r")
						return
					default:
						fmt.Printf("\r%s🧠 Generating AI fix... %s%s", colorYellow, spinChars[i%len(spinChars)], colorReset)
						i++
						time.Sleep(100 * time.Millisecond)
					}
				}
			}()

			fixPlan, err := GenerateFixPlanFromAI(finding, codeContext)
			spinDone <- true

			if err != nil {
				fmt.Printf("\r%s❌ Failed to generate fix: %v%s\n", colorRed, err, colorReset)
				skippedCount++
				continue
			}

			if fixPlan.RequiresManual {
				fmt.Printf("%s⚠️  This fix requires manual intervention%s\n", colorYellow, colorReset)
			}

			// Show confidence meter
			confidencePct := int(fixPlan.OverallConfidence * 100)
			confidenceBar := renderConfidenceMeter(fixPlan.OverallConfidence)
			fmt.Printf("\n%s📊 Confidence: %s %d%%%s\n", colorCyan, confidenceBar, confidencePct, colorReset)

			// Show proposed fix with enhanced diff
			if len(fixPlan.PrimaryFixes) > 0 {
				fmt.Printf("\n%s📝 Proposed Fix:%s\n", colorGreen, colorReset)
				for _, fix := range fixPlan.PrimaryFixes {
					fmt.Printf("  %sLine %d:%s\n", colorBold, fix.Line, colorReset)

					// Enhanced diff display
					fmt.Printf("    %s-%s %s\n", colorRed+colorBold, colorReset, strings.TrimSpace(fix.Original))
					fmt.Printf("    %s+%s %s\n", colorGreen+colorBold, colorReset, strings.TrimSpace(fix.Replacement))

					// Explanation with icon
					if fix.Explanation != "" {
						fmt.Printf("    %s💡 %s%s\n", colorGray, fix.Explanation, colorReset)
					}
				}
			}

			// User choice
			var choice string
			options := []huh.Option[string]{
				huh.NewOption("Apply Fix", "fix"),
				huh.NewOption("Skip", "skip"),
				huh.NewOption("Explain Risk", "explain"),
				huh.NewOption("Quit", "quit"),
			}

			prompt := huh.NewSelect[string]().
				Title("Action:").
				Options(options...).
				Value(&choice)

			if err := prompt.Run(); err != nil {
				fmt.Println("Session interrupted.")
				return
			}

			switch choice {
			case "fix":
				// Backup file
				if err := backupMgr.BackupFile(res.File); err != nil {
					fmt.Printf("%s❌ Backup failed: %v%s\n", colorRed, err, colorReset)
					continue
				}

				// Apply fixes
				for _, fix := range fixPlan.PrimaryFixes {
					if err := ApplyFix(res.File, fix); err != nil {
						fmt.Printf("%s❌ Fix failed: %v%s\n", colorRed, err, colorReset)
						continue
					}
				}

				// Apply additional changes
				for _, multifix := range fixPlan.AdditionalChanges {
					if err := ApplyMultiFileFix(multifix); err != nil {
						fmt.Printf("%s⚠️  Additional change failed: %v%s\n", colorYellow, err, colorReset)
					}
				}

				fmt.Printf("%s✅ Fix applied!%s\n", colorGreen, colorReset)
				fixedCount++

			case "explain":
				fmt.Printf("\n%s🧠 RISK EXPLANATION:%s\n", colorBold+colorYellow, colorReset)
				fmt.Println(finding.Explain)
				if finding.Remediation != "" {
					fmt.Printf("\n%s🔧 REMEDIATION:%s\n", colorBold+colorGreen, colorReset)
					fmt.Println(finding.Remediation)
				}
				fmt.Println()
				// Re-prompt after explanation
				continue

			case "skip":
				skippedCount++
				fmt.Printf("%s⏭️  Skipped%s\n", colorGray, colorReset)

			case "quit":
				fmt.Println("\n👋 Fix session ended.")
				printFixSummary(fixedCount, skippedCount, backupMgr)
				return
			}
		}
	}

	printFixSummary(fixedCount, skippedCount, backupMgr)
}

func runAutoFix(results []AnalysisResult, backupMgr *BackupManager) {
	fmt.Printf("🤖 Auto-fix mode (confidence threshold: %.0f%%)\n\n", confidenceThreshold*100)

	fixedCount := 0
	skippedCount := 0

	for _, res := range results {
		for _, finding := range res.Findings {
			// Generate fix
			codeContext := getCodeContext(res.File, finding.Line, 3)
			fixPlan, err := GenerateFixPlanFromAI(finding, codeContext)
			if err != nil || fixPlan.OverallConfidence < confidenceThreshold || fixPlan.RequiresManual {
				skippedCount++
				continue
			}

			// Backup
			if err := backupMgr.BackupFile(res.File); err != nil {
				fmt.Printf("❌ Backup failed for %s: %v\n", res.File, err)
				continue
			}

			// Apply
			for _, fix := range fixPlan.PrimaryFixes {
				if err := ApplyFix(res.File, fix); err != nil {
					fmt.Printf("❌ Fix failed for %s:%d: %v\n", res.File, fix.Line, err)
					continue
				}
			}

			fmt.Printf("✅ Fixed %s:%d - %s\n", res.File, finding.Line, finding.Message)
			fixedCount++
		}
	}

	printFixSummary(fixedCount, skippedCount, backupMgr)
}

func printFixSummary(fixed, skipped int, backupMgr *BackupManager) {
	fmt.Println("\n" + strings.Repeat("═", 60))
	fmt.Printf("📊 FIX SUMMARY\n")
	fmt.Printf("   ✅ Fixed: %d\n", fixed)
	fmt.Printf("   ⏭️  Skipped: %d\n", skipped)
	fmt.Printf("   💾 Backup: %s\n", backupMgr.BackupDir)
	fmt.Println(strings.Repeat("═", 60))

	if fixed > 0 {
		fmt.Printf("\n%s🔄 To rollback:%s tharos fix --rollback %s\n", colorYellow, colorReset, backupMgr.Timestamp)
	}
}

func handleRollback(timestamp string) {
	backupMgr := &BackupManager{
		BackupDir: ".tharos-backup/" + timestamp,
		Timestamp: timestamp,
	}

	fmt.Printf("🔄 Rolling back to: %s\n", timestamp)
	if err := backupMgr.Rollback(); err != nil {
		fmt.Printf("❌ Rollback failed: %v\n", err)
		return
	}

	fmt.Println("✅ Rollback complete!")
}

func getCodeContext(filePath string, line int, contextLines int) string {
	content, err := ioutil.ReadFile(filePath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	start := max(0, line-contextLines-1)
	end := min(len(lines), line+contextLines)

	context := strings.Join(lines[start:end], "\n")
	return context
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderConfidenceMeter creates a visual confidence meter
func renderConfidenceMeter(confidence float64) string {
	totalBars := 10
	filledBars := int(confidence * float64(totalBars))

	var meter strings.Builder
	for i := 0; i < totalBars; i++ {
		if i < filledBars {
			if confidence >= 0.9 {
				meter.WriteString(colorGreen + "█" + colorReset)
			} else if confidence >= 0.7 {
				meter.WriteString(colorYellow + "█" + colorReset)
			} else {
				meter.WriteString(colorRed + "█" + colorReset)
			}
		} else {
			meter.WriteString(colorGray + "░" + colorReset)
		}
	}

	return meter.String()
}
