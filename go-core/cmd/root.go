package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile         string
	outputFormat    string
	jsonOutput      bool
	aiEnabled       bool
	fixMode         bool
	verbose         bool
	policyPath      string
	policyDir       string
	interactiveMode bool
)

var rootCmd = &cobra.Command{
	Use:   "tharos",
	Short: "Tharos - Modern AI-Powered Git Hook Security Scanner",
	Long: `Tharos is a specialized git commit hook scanner that acts as an 
intelligent gatekeeper for your codebase, catching security vulnerabilities 
and leaks before they are committed.`,

	Version: "1.3.1",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ./tharos.yaml)")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output in JSON format")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "text", "output format (text, json, sarif)")
	rootCmd.PersistentFlags().BoolVar(&verbose, "verbose", false, "verbose output")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		viper.AddConfigPath(".")
		viper.SetConfigType("yaml")
		viper.SetConfigName("tharos")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if verbose {
			fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		}
	}
}
