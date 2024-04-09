package configs

import (
	"git.omidgolestani.ir/clinic/crm.api/pkg/infrastructure/log"
	"github.com/spf13/viper"
)

func init() {
}

func ReadConfig() (out Config) {
	viper.SetConfigFile("./.env") // name of config file (without extension)
	viper.SetConfigType("yaml")
	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&out)
	if err != nil {
		log.Error.Fatal(err)
	}
	return
}

type ServerMode = string

const (
	ServerModeDevelopment ServerMode = "development"
	ServerModeProduction  ServerMode = "production"
	ServerModeTest        ServerMode = "test"
)

type Config struct {
	ServerMode      ServerMode `mapstructure:"mode"`
	Port            int        `mapstructure:"port"`
	Host            string     `mapstructure:"host"`
	DataStorage     string     `mapstructure:"data-storage"`
	DataStorageName string     `mapstructure:"data-storage-name"`
	FileStorageName string     `mapstructure:"file-storage-name"`
	FileStorage     string     `mapstructure:"file-storage"`
	Cors            []string   `mapstructure:"cors"`
}
