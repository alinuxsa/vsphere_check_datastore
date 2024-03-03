package cmd

import (
	"github.com/alinuxsa/vsphere_check_datastore/common"
	"github.com/spf13/cobra"
)

var (
	VC string
)
var (
	rootCmd = &cobra.Command{
		Use:   "checkds",
		Short: "检查vsphere中datastore空间使用情况",
		Long:  `检查vsphere中datastore空间使用情况,生成csv文件并按使用率排序`,
		Run:   common.CheckDS,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&VC, "vsphere", "v", "", "配置文件中的vsphere名称")
	rootCmd.MarkPersistentFlagRequired("vsphere")
	cobra.OnInitialize(common.InitConfig)
}
