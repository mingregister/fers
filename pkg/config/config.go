package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

type Config struct {
	CryptoKey string  `mapstructure:"crypto_key"`
	Log       string  `mapstructure:"log"`
	TargetDir string  `mapstructure:"target_dir"`
	Storage   Storage `mapstructure:"storage"`
	LogLevel  int     `mapstructure:"log_level"`
	Pprof     Pprof   `mapstructure:"pprof"`
}

type Pprof struct {
	Enabled bool   `mapstructure:"enabled"`
	Port    int    `mapstructure:"port"`
	Host    string `mapstructure:"host"`
}

type Storage struct {
	RemoteType string    `mapstructure:"remote_type"`
	Localhost  Localhost `mapstructure:"localhost"`
	Oss        OSS       `mapstructure:"oss"`
}

type Localhost struct {
	Workdir string `mapstructure:"work_dir"`
}

// OSS contains the configuration for OSS client
type OSS struct {
	Enabled         bool   `mapstructure:"enabled"`
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"access_key_id"`
	AccessKeySecret string `mapstructure:"access_key_secret"`
	BucketName      string `mapstructure:"bucket_name"`
	Region          string `mapstructure:"region"`
	WorkDir         string `mapstructure:"workDir"`
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

	v.SetDefault("log_level", 0) // 默认日志级别为INFO

	// 设置 pprof 默认配置
	v.SetDefault("pprof.enabled", false)
	v.SetDefault("pprof.port", 6060)
	v.SetDefault("pprof.host", "localhost")

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
		v.AddConfigPath(fmt.Sprintf("%s/.fers", homeDir))
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
