package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "wordle-data",
	Short: "Create Wordle imitation-learning datasets",
	Long:  "Create synthetic Wordle game-state datasets for machine learning.",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
