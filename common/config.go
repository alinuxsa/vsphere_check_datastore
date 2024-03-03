package common

import (
	"fmt"
	"os"

	"github.com/spf13/viper"
)

func InitConfig() {
	viper.AddConfigPath(".")
	viper.SetConfigName(".env")
	viper.SetConfigType("yaml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Println("读取配置文件出错: ", err)
		os.Exit(1)
	}
}

type VsphersConfig struct {
	Username string
	Password string
	Host     string
	Insecure bool
}
