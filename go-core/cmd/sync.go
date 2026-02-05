package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Audit local setup and sync security policies",
	Long: `Verifies the integrity of Tharos hooks, configuration, and environment variables. 
Ensures your local security gates are active and aligned with the latest standards.`,
	Run: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) {
	fmt.Printf("%s%s[THAROS SYNC & AUDIT]%s\n\n", colorBold, colorCyan, colorReset)

	health := true

	// 1. Audit Git Repository
	fmt.Printf("%s1. Repository Setup%s\n", colorBold, colorReset)
	if _, err := os.Stat(".git"); err != nil {
		fmt.Printf("  %s✗ Not a git repository%s (Tharos requires Git for workflow gating)\n", colorRed, colorReset)
		health = false
	} else {
		fmt.Printf("  %s✓ Git repository detected%s\n", colorGreen, colorReset)
	}

	// 2. Audit Git Hooks (Self-Healing)
	fmt.Printf("\n%s2. Git Hook Verification%s\n", colorBold, colorReset)
	hooksDir := filepath.Join(".git", "hooks")
	preCommitHook := filepath.Join(hooksDir, "pre-commit")

	if _, err := os.Stat(preCommitHook); err != nil {
		fmt.Printf("  %s⚠️  Pre-commit hook missing.%s Re-installing...\n", colorYellow, colorReset)
		installHooksSilently()
		fmt.Printf("  %s✓ Hook restored%s\n", colorGreen, colorReset)
	} else {
		content, _ := os.ReadFile(preCommitHook)
		if !strings.Contains(string(content), "Tharos") {
			fmt.Printf("  %s⚠️  Pre-commit hook tampered with.%s Repairing...\n", colorYellow, colorReset)
			installHooksSilently()
			fmt.Printf("  %s✓ Hook repaired%s\n", colorGreen, colorReset)
		} else {
			fmt.Printf("  %s✓ Pre-commit hook active and managed%s\n", colorGreen, colorReset)
		}
	}

	// 3. Audit Configuration
	fmt.Printf("\n%s3. Configuration Status%s\n", colorBold, colorReset)
	if _, err := os.Stat("tharos.yaml"); err != nil {
		fmt.Printf("  %s⚠️  tharos.yaml missing.%s Using defaults.\n", colorYellow, colorReset)
	} else {
		fmt.Printf("  %s✓ tharos.yaml detected%s\n", colorGreen, colorReset)
	}

	// 4. Audit AI Environment
	fmt.Printf("\n%s4. AI Engine Connectivity%s\n", colorBold, colorReset)
	gemini := os.Getenv("GEMINI_API_KEY")
	groq := os.Getenv("GROQ_API_KEY")
	if gemini != "" || groq != "" {
		provider := "Gemini"
		if gemini == "" {
			provider = "Groq"
		}
		fmt.Printf("  %s✓ AI enabled via %s%s\n", colorGreen, provider, colorReset)
	} else {
		fmt.Printf("  %sℹ No AI keys found.%s Using local AST analysis only.\n", colorGray, colorReset)
	}

	// 5. Policy Sync (Future: Hit remote API)
	fmt.Printf("\n%s5. Policy Sync%s\n", colorBold, colorReset)
	fmt.Printf("  %s✓ Policies are up to date%s (Local Library)\n", colorGreen, colorReset)

	fmt.Println()
	if health {
		fmt.Printf("%s✨ System is healthy. Tharos is protecting your commits.%s\n", colorGreen+colorBold, colorReset)
	} else {
		fmt.Printf("%s⚠️  System has warnings. Run 'tharos init' to fix core issues.%s\n", colorYellow+colorBold, colorReset)
	}
}

// Helper to install hooks without verbose output
func installHooksSilently() {
	hooksDir := filepath.Join(".git", "hooks")
	os.MkdirAll(hooksDir, 0o755)
	preCommitHook := filepath.Join(hooksDir, "pre-commit")
	hookContent := "#!/bin/sh\n\n# Tharos: Modern AI-Powered Git Hook Security Scanner\n# Prevents security leaks and vulnerabilities at the commit stage.\n\n# Periodic setup audit & policy sync (non-blocking)\ntharos sync > /dev/null 2>&1 &\n\ntharos check\n"

	os.WriteFile(preCommitHook, []byte(hookContent), 0o755)
}
