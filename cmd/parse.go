package cmd

import (
	"encoding/csv"
	"fmt"
	"io/ioutil"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/manifoldco/promptui"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

type Bill struct {
	title  []string   // 标题
	aliPay [][]string // 支付宝账单
	wechat [][]string // 微信账单
}

func (b *Bill) ReadAliPay(path string) error {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	// 支付宝账单是 GBK 格式的，需要转码
	gbkFile := transform.NewReader(f, simplifiedchinese.GBK.NewDecoder())
	reader := csv.NewReader(gbkFile)
	reader.FieldsPerRecord = -1

	data, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		return err
	}
	title := data[4]
	columnLen := len(title)
	for _, v := range data {
		if len(v) == columnLen && !reflect.DeepEqual(v, title) {
			var temp []string
			// 对于资金状态为空的情况，交易关闭，不计入结算。
			if v[15] == "" {
				continue
			}
			// 对于资金状态为“资金转移”的，如果服务费为零，不计入结算。
			if v[15] == "资金转移" && v[12] == "0.00" {
				continue
			}
			// 移除单元格的空格
			for _, v1 := range v {
				temp = append(temp, strings.Trim(v1, " "))
			}
			// 将服务费算作支出，填写到金额中。
			fee, _ := strconv.ParseFloat(v[12], 64)
			if fee > 0.00 {
				tempV := v
				tempV[9] = v[12]
				tempV[10] = "支出"
				data = append(data, tempV) // FIXME: 循环的 data 是复制出来的副本
			}

			// fmt.Println(temp[9])
			// 支付宝账单[]string{"交易号", "商家订单号", "交易创建时间", "付款时间 ", "最近修改时间", "交易来源地", "类型", "交易对方", "商品名称", "金额（元）", "收/支", "交易状态 ", "服务费（元）", "成功退款（元）", "备注", "资金状态"}

			// 微信账单
			// []string{"交易时间", "交易类型", "交易对方", "商品", "收/支", "金额(元)", "支付方式", "当前状态", "交易单号", "商户单号", "备注"}
			tempV := []string{temp[2], temp[6], temp[7], temp[8], temp[10], "¥" + temp[9], "支付宝", temp[15], temp[0], temp[1], temp[14]}
			// fmt.Printf("%#v\n", temp)
			b.aliPay = append(b.aliPay, tempV)
		}
	}
	return nil
}

func (b *Bill) ReadWechatPay(path string) error {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer f.Close()

	reader := csv.NewReader(f)
	reader.FieldsPerRecord = -1

	data, err := reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		return err
	}
	title := data[16]
	// fmt.Printf("%#v\n", title)
	columnLen := len(title)
	for _, v := range data {
		if len(v) == columnLen && !reflect.DeepEqual(v, title) {
			// []string{"交易时间", "交易类型", "交易对方", "商品", "收/支", "金额(元)", "支付方式", "当前状态", "交易单号", "商户单号", "备注"}

			// 删除收支为空且备注为空或备注为服务费￥0.00的，中性交易。
			if (v[4] == "/" && v[10] == "/") || (v[4] == "/" && v[10] == "服务费¥0.00") {
				continue
			}
			var temp []string
			for _, v1 := range v {
				temp = append(temp, strings.Trim(strings.Trim(v1, "/"), " "))
			}
			// TODO: 从备注中提取出有服务费的项目，填写进入收支明细中。
			b.wechat = append(b.wechat, temp)
		}
	}
	return nil
}

// WriteMergeFile 合并账单
func (b *Bill) WriteMergeFile(path string) error {
	destFile, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println(err)
	}
	defer destFile.Close()
	_, err = destFile.WriteString("\xEF\xBB\xBF")
	if err != nil {
		return err
	} // 写入一个UTF-8 BOM

	df := csv.NewWriter(destFile)
	b.GetTitle()
	err = df.Write(b.title)
	if err != nil {
		return err
	}

	for _, row := range b.aliPay {
		err = df.Write(row)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	for _, row := range b.wechat {
		err = df.Write(row)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}

	df.Flush()
	return nil
}

// ReadMergeFile 读取合并后的账单
func (b *Bill) ReadMergeFile(path string) (data [][]string, err error) {
	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer f.Close()
	reader := csv.NewReader(f)

	reader.FieldsPerRecord = -1
	data, err = reader.ReadAll()
	if err != nil {
		fmt.Println(err)
		return
	}
	return
}

// 支付宝账单
// []string{"交易号", "商家订单号", "交易创建时间", "付款时间 ", "最近修改时间", "交易来源地", "类型", "交易对方", "商品名称", "金额（元）", "收/支", "交易状态 ", "服务费（元）", "成功退款（元）", "备注", "资金状态"}

// 微信账单
// []string{"交易时间", "交易类型", "交易对方", "商品", "收/支", "金额(元)", "支付方式", "当前状态", "交易单号", "商户单号", "备注"}

// GetTitle 转换 title
func (b *Bill) GetTitle() {
	titles := []string{"交易时间", "交易类型", "交易对方", "商品", "收/支", "金额(元)", "支付方式", "交易状态", "交易单号", "商户单号", "备注", "服务费(元)"}
	b.title = append(b.title, titles...)
}

// GetCSVPath 查找当前目录下的账单文件
func GetCSVPath() (items []string) {
	pwd, _ := os.Getwd()
	fileInfoList, _ := ioutil.ReadDir(pwd)
	for _, info := range fileInfoList {
		if !info.IsDir() && strings.Contains(info.Name(), ".csv") {
			items = append(items, info.Name())
		}
	}
	return
}

func PromptSelectCSVPath(bType string) string {
	bMap := map[string]string{
		"1": "支付宝",
		"2": "微信",
	}
	items := GetCSVPath()
	index := -1
	var result string
	var err error

	for index < 0 {
		prompt := promptui.SelectWithAdd{
			Label: "请选择【" + bMap[bType] + "】账单",
			Items: items,
		}

		index, result, err = prompt.Run()

		if index == -1 {
			items = append(items, result)
		}
	}

	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Input: %s\n", result)

	return result
}

func PromptSelectAnalysis() string {
	items := GetCSVPath()
	prompt := promptui.Select{
		Label: "请选择【合并后的】账单",
		Items: items,
	}
	_, result, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("You choose: %s\n", result)

	return result
}
