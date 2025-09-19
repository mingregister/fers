package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	CryptoKey string `mapstructure:"crypto_key"`
	Log       string `mapstructure:"log"`
	TargetDir string `mapstructure:"target_dir"`
	OssDir    string `mapstructure:"oss_dir"`
}

func NewConfig() (*Config, error) {
	config, err := LoadFromFile("config")
	if err != nil {
		// 如果加载配置文件失败，使用默认配置
		return nil, fmt.Errorf("load config file, err: %w", err)
	}
	return config, nil
}

// LoadFromFile 使用Viper从配置文件加载配置
func LoadFromFile(configName string) (*Config, error) {
	v := viper.New()

	// 设置配置文件名（不包含扩展名）
	v.SetConfigName(configName)
	v.SetConfigType("yaml")

	// 添加配置文件搜索路径
	// 1. 当前工作目录
	v.AddConfigPath(".")

	// 2. 可执行文件所在目录
	if execPath, err := os.Executable(); err == nil {
		execDir := filepath.Dir(execPath)
		v.AddConfigPath(execDir)
	}

	// 3. 用户主目录
	if homeDir, err := os.UserHomeDir(); err == nil {
		v.AddConfigPath(homeDir)
	}

	// 读取配置文件
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config failed, %w", err)
	}

	// 将配置解析到结构体
	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("unable to decode config into struct: %w", err)
	}

	return &config, nil
}
