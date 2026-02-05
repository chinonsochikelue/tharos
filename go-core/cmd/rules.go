package cmd

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "List all active security rules and policies",
	Long:  `Display a unified list of built-in AST checks, custom configuration rules, and externally loaded policies.`,
	Run:   runRules,
}

func init() {
	rootCmd.AddCommand(rulesCmd)

	rulesCmd.Flags().StringVar(&policyPath, "policy", "", "path to external policy file (YAML)")
	rulesCmd.Flags().StringVar(&policyDir, "policy-dir", "policies", "directory for policy files")
}

func runRules(cmd *cobra.Command, args []string) {
	// Load external policies if flags are set
	loadExternalPolicies(policyPath, policyDir)

	rules := GetAllActiveRules()

	if len(rules) == 0 {
		fmt.Println("No active security rules found.")
		return
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("36")).
		Padding(0, 1)

	fmt.Println(headerStyle.Render("\n📜 THAROS ACTIVE SECURITY POLICIES"))
	fmt.Println()

	re := table.New().
		Border(lipgloss.NormalBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		Headers("ID", "SOURCE", "SEVERITY", "DESCRIPTION").
		Rows(formatRuleRows(rules)...)

	fmt.Println(re.Render())
	fmt.Println()
}

func formatRuleRows(rules []UnifiedRule) [][]string {
	var rows [][]string
	for _, r := range rules {
		sevCol := getSeverityColor(r.Severity)
		rows = append(rows, []string{
			r.ID,
			r.Source,
			sevCol + strings.ToUpper(r.Severity) + colorReset,
			r.Description,
		})
	}
	return rows
}
