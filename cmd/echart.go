package cmd

import (
	"strings"

	"github.com/spf13/cobra"
)

var _ = &cobra.Command{
	Use:   "chart",
	Short: "图表分析账单",
	Long:  `使用 awm chart 分析合并后的账单`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		oPath = PromptSelectAnalysis()
		var b Bill
		if strings.Contains(oPath, ".csv") {
			_, err := b.ReadMergeFile(oPath)
			if err != nil {
				return err
			}
		}
		return nil
	},
}

func init() {
	// rootCmd.AddCommand(chartCmd)
}
