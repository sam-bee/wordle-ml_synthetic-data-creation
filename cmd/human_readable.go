package cmd

import (
	"fmt"
	"wordle/dataset"

	"github.com/spf13/cobra"
)

var humanReadableOutputDir = dataset.DefaultHumanReadableOutputDir

var humanReadableCmd = &cobra.Command{
	Use:   "human-readable DATASET.bin",
	Short: "Convert a binary dataset file to human-readable JSON",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		outputPath, err := dataset.WriteHumanReadableFile(args[0], humanReadableOutputDir)
		if err != nil {
			return err
		}

		fmt.Fprintf(cmd.OutOrStdout(), "wrote %s\n", outputPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(humanReadableCmd)

	humanReadableCmd.Flags().StringVar(&humanReadableOutputDir, "output", humanReadableOutputDir, "directory for human-readable JSON output")
}
