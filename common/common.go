package common

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/property"
	"github.com/vmware/govmomi/vim25/mo"
)

type DatastoreInfo struct {
	Name         string
	CapacityGB   float64
	UsedSpaceGB  float64
	FreeSpaceGB  float64
	UsagePercent float64
}

func CheckDS(cmd *cobra.Command, args []string) {
	vc, _ := cmd.Flags().GetString("vsphere")
	strNow := time.Now().Format("2006_01_02_15_04_05")

	csvfile := fmt.Sprintf("%s_datastore_%s.csv", vc, strNow)
	var config VsphersConfig
	err := viper.UnmarshalKey(vc, &config)
	if err != nil {
		log.Fatalf("无法解析配置: %v", err)
	}
	u := &url.URL{
		Scheme: "https",
		Host:   config.Host,
		Path:   "/sdk",
	}
	u.User = url.UserPassword(config.Username, config.Password)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c, err := govmomi.NewClient(ctx, u, config.Insecure)
	if err != nil {
		log.Fatal("连接vsphere出错: ", err)
	}

	f := find.NewFinder(c.Client, true)
	dc, err := f.DefaultDatacenter(ctx)
	if err != nil {
		log.Fatal("设置数据中心出错: ", err)
	}
	f.SetDatacenter(dc)

	dss, err := f.DatastoreList(ctx, "*")
	if err != nil {
		log.Fatal("获取所有datastore出错: ", err)
	}

	pc := property.DefaultCollector(c.Client)

	var datastores []DatastoreInfo

	for _, ds := range dss {
		var dsMo mo.Datastore

		err = pc.RetrieveOne(ctx, ds.Reference(), []string{"summary"}, &dsMo)
		if err != nil {
			fmt.Println(err)
			continue
		}

		summary := dsMo.Summary
		// 跳过 datastore开头的本地存储
		if strings.HasPrefix(summary.Name, "datastore") {
			continue
		}

		capacity := summary.Capacity
		freeSpace := summary.FreeSpace
		usedSpace := capacity - freeSpace
		usagePercent := (float64(usedSpace) / float64(capacity)) * 100

		datastores = append(datastores, DatastoreInfo{
			Name:         summary.Name,
			CapacityGB:   float64(capacity) / 1024 / 1024 / 1024,
			UsedSpaceGB:  float64(usedSpace) / 1024 / 1024 / 1024,
			FreeSpaceGB:  float64(freeSpace) / 1024 / 1024 / 1024,
			UsagePercent: usagePercent,
		})
	}
	// 按剩余空间排序
	sort.Slice(datastores, func(i, j int) bool {
		return datastores[i].FreeSpaceGB < datastores[j].FreeSpaceGB
	})

	// 结果写入csv
	file, err := os.Create(csvfile)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// 写入表头
	writer.Write([]string{"存储名", "容量 (GB)", "已用容量 (GB)", "剩余容量 (GB)", "使用率 (%)"})

	// 写入数据
	for _, ds := range datastores {
		writer.Write([]string{
			ds.Name,
			fmt.Sprintf("%.2f", ds.CapacityGB),
			fmt.Sprintf("%.2f", ds.UsedSpaceGB),
			fmt.Sprintf("%.2f", ds.FreeSpaceGB),
			fmt.Sprintf("%.2f", ds.UsagePercent),
		})
	}
}
