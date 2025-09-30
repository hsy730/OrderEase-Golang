package config

import (
	"fmt"
	"os"
	"strconv"

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
		LogLevel  int    `yaml:"logLevel"`
	} `yaml:"database"`

	JWT struct {
		Secret     string `yaml:"secret"`
		Expiration int    `yaml:"expiration"`
	} `yaml:"jwt"`
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

	// 从环境变量覆盖配置
	// 服务器配置
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			AppConfig.Server.Port = p
		}
	}

	if host := os.Getenv("SERVER_HOST"); host != "" {
		AppConfig.Server.Host = host
	}

	// 数据库配置
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		AppConfig.Database.Host = dbHost
	}

	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		if p, err := strconv.Atoi(dbPort); err == nil {
			AppConfig.Database.Port = p
		}
	}

	if dbUsername := os.Getenv("DB_USERNAME"); dbUsername != "" {
		AppConfig.Database.Username = dbUsername
	}

	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		AppConfig.Database.Password = dbPassword
	}

	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		AppConfig.Database.DBName = dbName
	}

	// JWT配置
	if jwtSecret := os.Getenv("JWT_SECRET"); jwtSecret != "" {
		AppConfig.JWT.Secret = jwtSecret
	}

	if jwtExp := os.Getenv("JWT_EXPIRATION"); jwtExp != "" {
		if exp, err := strconv.Atoi(jwtExp); err == nil {
			AppConfig.JWT.Expiration = exp
		}
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
