package config

type Config struct {
	CryptoKey string
	Log       string

	TargetDir string

	OssDir string
}

func NewConfig() *Config {
	// TODO: 从配置文件加载
	return &Config{
		CryptoKey: "admin123@alls",
		Log:       "D:/000qjl/code/fers/tmp/fers.log",
		TargetDir: "D:/000qjl/code/fers/tmp/testworddir",
		OssDir:    "D:/000qjl/code/fers/tmp/oss_mock",
	}
}
