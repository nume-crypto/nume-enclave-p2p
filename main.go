package main

import (
	"fmt"

	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile(`.env`)
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	fmt.Println(viper.GetString(`DB_DRIVER`))
}
