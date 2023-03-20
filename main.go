package main

import (
	"fmt"

	"github.com/spf13/viper"
)

const ID = "ad73ca38-f4db-461c-996e-eff3091d22aa"

type Env struct {
	DBDriver      string `mapstructure:"DB_DRIVER"`
	DBSource      string `mapstructure:"DB_SOURCE"`
	ServerAddress string `mapstructure:"SERVER_ADDRESS"`
}

func init() {
	viper.SetConfigFile(`.env`)
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println(viper.GetString(`DB_DRIVER`))
}
