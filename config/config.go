package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Server struct {
		Port           int      `yaml:"port"`
		BasePath       string   `yaml:"basePath"`
		Host           string   `yaml:"host"`
		Domain         string   `yaml:"domain"`
		AllowedOrigins []string `yaml:"allowedOrigins"`
	} `yaml:"server"`

	Database struct {
		Driver    string `yaml:"driver"`
		Host      string `yaml:"host"`
		Port      int    `yaml:"port"`
		Username  string `yaml:"username"`
		Password  string `yaml:"password"`
		DBName    string `yaml:"dbname"`
		Charset   string `yaml:"charset"`
		ParseTime bool   `yaml:"parseTime"`
		Loc       string `yaml:"loc"`
	} `yaml:"database"`
}

var AppConfig Config

func LoadConfig(configPath string) error {
	configFile, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件错误: %v", err)
	}

	err = yaml.Unmarshal(configFile, &AppConfig)
	if err != nil {
		return fmt.Errorf("解析配置文件错误: %v", err)
	}

	return nil
}

func GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=%v&loc=%s",
		AppConfig.Database.Username,
		AppConfig.Database.Password,
		AppConfig.Database.Host,
		AppConfig.Database.Port,
		AppConfig.Database.DBName,
		AppConfig.Database.Charset,
		AppConfig.Database.ParseTime,
		AppConfig.Database.Loc,
	)
}
