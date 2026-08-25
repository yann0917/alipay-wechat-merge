package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// 以下两个变量由 ldflags 在编译时注入（见 .github/workflows/release.yml），
// 本地开发构建时保持默认值。
var (
	// Version 发布版本号，例如 v1.2.3
	Version = "dev"
	// Build 构建时间，格式为 UTC 时间的 RFC3339，例如 2026-08-25T10:00:00Z
	Build = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "打印版本号",
	Long:  `打印 awm 版本号`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("awm %s (build %s)\n", Version, Build)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
