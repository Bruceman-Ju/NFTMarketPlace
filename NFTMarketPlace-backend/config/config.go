package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	Server struct {
		Port           int      `mapstructure:"port"`
		AllowedOrigins []string `mapstructure:"allowed_origins"`
	}
	Database struct {
		DSN string `mapstructure:"dsn"`
	}
	Redis struct {
		Addr     string `mapstructure:"addr"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
	}
	JWT struct {
		Algorithm      string `mapstructure:"algorithm"`
		Secret         string `mapstructure:"secret"`
		PrivateKeyPath string `mapstructure:"private_key_path"`
		PublicKeyPath  string `mapstructure:"public_key_path"`
		ExpireHours    int    `mapstructure:"expire_hours"`
	}
	Eth struct {
		RPCURL          string `mapstructure:"rpc_url"`
		ContractAddress string `mapstructure:"contract_address"`
		StartBlock      uint64 `mapstructure:"start_block"`
		WebSocketURL    string `mapstructure:"websocket_url"`
	}
}

var Cfg Config

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("config")

	err := viper.ReadInConfig()
	if err != nil {
		panic(err)
	}

	err = viper.Unmarshal(&Cfg)
	if err != nil {
		panic(err)
	}
}
