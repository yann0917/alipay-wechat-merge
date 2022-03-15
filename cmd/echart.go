package cmd

import (
	"os"
	"sort"
	"strings"

	"github.com/go-echarts/go-echarts/v2/charts"
	"github.com/go-echarts/go-echarts/v2/opts"
	"github.com/go-echarts/go-echarts/v2/types"
	"github.com/spf13/cobra"
)

var chartCmd = &cobra.Command{
	Use:   "chart",
	Short: "图表分析账单",
	Long:  `使用 awm chart 分析合并后的账单`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		oPath = PromptSelectAnalysis()
		var b Bill
		if strings.Contains(oPath, ".csv") {
			list, err := b.ReadMergeFile(oPath)
			if err != nil {
				return err
			}
			Bar(list)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(chartCmd)
}

// generate random data for bar chart
func generateBarItems(list []*Account, pType int) []opts.BarData {
	items := make([]opts.BarData, 0)
	iTotal, oTotal := 0.0, 0.0
	for _, account := range list {
		if account.IO == "收入" {
			iTotal += account.Money * 100
		}
		if account.IO == "支出" {
			oTotal += account.Money * 100
		}
	}
	if pType == 1 {
		items = append(items, opts.BarData{Name: "累计收入", Value: iTotal / 100})
	}
	if pType == 2 {
		items = append(items, opts.BarData{Name: "累计支出", Value: oTotal / 100})
	}
	return items
}

func generateLineItems(list []*Account, pType int) []opts.LineData {
	sort.Sort(Accounts(list))
	items := make([]opts.LineData, 0)
	goodsName := ""
	for _, account := range list {
		if account.TransFrom != "" {
			goodsName = account.TransFrom + ":" + account.GoodsName
		} else {
			goodsName = account.TransType
		}
		switch pType {
		case 1:
			if account.IO == "收入" {
				items = append(items, opts.LineData{Name: goodsName, Value: account.Money})
			} else {
				items = append(items, opts.LineData{Name: "", Value: 0})
			}
		case 2:
			if account.IO == "支出" {
				items = append(items, opts.LineData{Name: goodsName, Value: account.Money})
			} else {
				items = append(items, opts.LineData{Name: "", Value: 0})
			}
		}
	}

	return items
}

func Bar(list []*Account) {
	// create a new bar instance
	bar := charts.NewBar()
	// set some global options like Title/Legend/ToolTip or anything else
	bar.SetGlobalOptions(
		charts.WithInitializationOpts(opts.Initialization{
			PageTitle: "账单分析",
			Theme:     types.ThemeMacarons}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "收支分析",
			Subtitle: "账单中收支项总和",
		}),
	)

	var xAxis []string
	sort.Sort(Accounts(list))
	for _, account := range list {
		xAxis = append(xAxis, account.TransAt.Time.String())
	}

	bar.SetXAxis([]string{"累计"}).
		AddSeries("收入", generateBarItems(list, 1)).
		AddSeries("支出", generateBarItems(list, 2))

	line := charts.NewLine()
	var lineX []string
	for _, account := range list {
		lineX = append(lineX, account.TransAt.Format(layout))
	}

	line.SetGlobalOptions(charts.WithInitializationOpts(opts.Initialization{
		PageTitle: "账单分析",
		Theme:     types.ThemeMacarons}),
		charts.WithLegendOpts(opts.Legend{
			Show: true,
		}),
		charts.WithDataZoomOpts(opts.DataZoom{
			Type: "slider",
		}),
		charts.WithTooltipOpts(opts.Tooltip{
			Show: true,
		}),
		charts.WithTitleOpts(opts.Title{
			Title:    "收支详情分析",
			Subtitle: "账单中收支项详情 by 交易时间",
		}),
	)

	line.SetXAxis(lineX).
		AddSeries("收入", generateLineItems(list, 1)).
		AddSeries("支出", generateLineItems(list, 2)).
		SetSeriesOptions(charts.WithLineChartOpts(opts.LineChart{Smooth: true}))

	// Where the magic happens
	f, _ := os.Create("charts.html")
	bar.Render(f)
	line.Render(f)
}
