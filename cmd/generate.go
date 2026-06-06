package cmd

import (
	"context"
	"fmt"
	"runtime"
	"time"
	"wordle/dataset"

	"github.com/spf13/cobra"
)

var generateConfig = dataset.Config{
	OutputDir:       "data",
	TopK:            dataset.FixedTopK,
	MaxDepth:        dataset.MaxDepth,
	RecordsPerDepth: 5,
	IncludeOpening:  true,
	Workers:         runtime.GOMAXPROCS(0),
	Seed:            -1,
}

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate Wordle imitation-learning dataset files",
	RunE: func(cmd *cobra.Command, args []string) error {
		if generateConfig.Seed < 0 {
			generateConfig.Seed = time.Now().UnixNano()
		}

		generateConfig.ProgressWriter = cmd.OutOrStdout()
		result, err := dataset.Generate(context.Background(), generateConfig)
		if err != nil {
			return err
		}

		for _, split := range result.Splits {
			fmt.Fprintf(
				cmd.OutOrStdout(),
				"%s: wrote %d records for %d solutions to %s and %s\n",
				split.Name,
				split.RecordCount,
				split.SolutionCount,
				split.BinaryPath,
				split.MetadataPath,
			)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)

	generateCmd.Flags().StringVar(&generateConfig.OutputDir, "output", generateConfig.OutputDir, "directory for generated dataset files")
	generateCmd.Flags().Int64Var(&generateConfig.Seed, "seed", generateConfig.Seed, "random seed; negative values use the current time")
	generateCmd.Flags().IntVar(&generateConfig.Workers, "workers", generateConfig.Workers, "number of solution-generation workers")
	generateCmd.Flags().IntVar(&generateConfig.RecordsPerDepth, "records-per-depth", generateConfig.RecordsPerDepth, "records generated for each solution at each depth")
}
