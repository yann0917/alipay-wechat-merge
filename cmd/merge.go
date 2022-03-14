package cmd

import (
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	aPath string
	wPath string
	oPath string
)

var mergeCmd = &cobra.Command{
	Use:   "merge",
	Short: "合并账单",
	Long:  `使用 awm merge 合并支付宝 微信账单`,
	Args:  cobra.OnlyValidArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		var b Bill
		aPath = PromptSelectCSVPath("1")
		if strings.Contains(aPath, ".csv") {
			if err := b.ReadAliPay(aPath); err != nil {
				return err
			}
		}
		wPath = PromptSelectCSVPath("2")
		if strings.Contains(wPath, ".csv") {
			if err := b.ReadWechatPay(wPath); err != nil {
				return err
			}
		}
		unix := strconv.FormatInt(time.Now().Unix(), 10)
		if err := b.WriteMergeFile("output_" + unix + ".csv"); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(mergeCmd)
}
