package cmd

import (
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "Check AI provider configuration",
	Long:  `Display the status of AI provider configuration (Gemini, Groq) and show setup instructions.`,
	Run:   runSetup,
}

func init() {
	rootCmd.AddCommand(setupCmd)
}

func runSetup(cmd *cobra.Command, args []string) {
	setupCommand()
}
