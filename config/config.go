package config

import (
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"path/filepath"
)

type Config struct {
	Port      int    `yaml:"port"`
	ServerUrl string `yaml:"server_url"`
}

var GlobalConfig *Config

func ParseConfig(filename string) *Config {

	// 获取当前可执行文件的完整路径
	executable, err := os.Executable()
	if err != nil {
		log.Fatalf("\n\nUnable to get executable path: %+v\n\n", err)
	}

	// 如使用IDE调试，请改为本地路径
	dir := filepath.Dir(executable)
	configPath := filepath.Join(dir, filename)

	data, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("\n\nNot found config file, %+v\n\n", err)
	}

	var cfg *Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		log.Fatalf("\n\nUnable to parse config file, %+v\n\n", err)
	}
	GlobalConfig = cfg
	return cfg
}
