package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/go-git/go-git/v5"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize Tharos hooks in the current repository",
	Long:  `Sets up a pre-commit hook in the .git directory to run Tharos analysis automatically before every commit.`,
	Run:   runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) {
	fmt.Println("🛡️ Initializing Tharos...")

	dir, err := os.Getwd()
	if err != nil {
		fmt.Printf("❌ Failed to get current directory: %v\n", err)
		os.Exit(1)
	}

	// Verify it's a git repo
	_, err = git.PlainOpen(dir)
	if err != nil {
		fmt.Printf("❌ Not a git repository: %v\n", err)
		os.Exit(1)
	}

	hooksDir := filepath.Join(dir, ".git", "hooks")
	if _, err := os.Stat(hooksDir); os.IsNotExist(err) {
		err = os.MkdirAll(hooksDir, 0o755)
		if err != nil {
			fmt.Printf("❌ Failed to create hooks directory: %v\n", err)
			os.Exit(1)
		}
	}

	preCommitHook := filepath.Join(hooksDir, "pre-commit")

	// Hook content
	hookContent := "#!/bin/sh\n\n# Tharos: Modern AI-Powered Git Hook Security Scanner\n# Prevents security leaks and vulnerabilities at the commit stage.\n\n# Periodic setup audit & policy sync (non-blocking)\ntharos sync > /dev/null 2>&1 &\n\ntharos check\n"

	err = os.WriteFile(preCommitHook, []byte(hookContent), 0o755)
	if err != nil {
		fmt.Printf("❌ Failed to write pre-commit hook: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✅ Tharos hooks installed successfully!")
}
