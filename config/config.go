/**
 * @Author: YMBoom
 * @Description:
 * @File:  config
 * @Version: 1.0.0
 * @Date: 2023/01/12 9:44
 */
package config

import (
	"bytes"
	_ "embed"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	config = new(Config)
	//go:embed config.yaml
	cfg []byte
)

type Config struct {
	ExtraData []struct {
		Random    string `json:"random"`
		ExtraData string `json:"extraData"`
	} `json:"extraData"`
	Extend string `json:"extend"`

	Cookies       []string `json:"cookies"`
	UserAgent     string   `json:"userAgent"`
	Referer       string   `json:"referer"`
	ContentType   string   `json:"contentType"`
	Domain        string   `json:"domain"`
	FunctionId    string   `json:"functionId"`
	Appid         string   `json:"appid"`
	Uuid          string   `json:"uuid"`
	Client        string   `json:"client"`
	MonitorSource string   `json:"monitorSource"`
	At            string   `json:"at"`
	Early         int      `json:"early"`
}

func ReadConfig() {
	r := bytes.NewReader(cfg)

	viper.SetConfigType("yaml")

	if err := viper.ReadConfig(r); err != nil {
		panic(err)
	}

	if err := viper.Unmarshal(config); err != nil {
		panic(err)
	}

	viper.SetConfigName("config")
	viper.AddConfigPath("./")
	viper.OnConfigChange(func(e fsnotify.Event) {
		if err := viper.Unmarshal(config); err != nil {
			panic(err)
		}
	})
}

func Get() *Config {
	return config
}
