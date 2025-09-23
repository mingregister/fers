package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFromFile_ValidConfig(t *testing.T) {
	// Create a temporary directory for test config
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a valid config file
	configContent := `
crypto_key: "test-crypto-key-123"
log: "test.log"
target_dir: "/tmp/test"
log_level: 1
storage:
  remote_type: "localhost"
  localhost:
    work_dir: "/tmp/localhost"
  oss:
    enabled: true
    endpoint: "oss-cn-hangzhou.aliyuncs.com"
    access_key_id: "test-access-key"
    access_key_secret: "test-secret"
    bucket_name: "test-bucket"
    region: "cn-hangzhou"
    workDir: "/oss/work"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	// Change to temp directory to test config loading
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	// Load config
	config, err := LoadFromFile("config")
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify config values
	if config.CryptoKey != "test-crypto-key-123" {
		t.Errorf("Expected CryptoKey 'test-crypto-key-123', got '%s'", config.CryptoKey)
	}

	if config.Log != "test.log" {
		t.Errorf("Expected Log 'test.log', got '%s'", config.Log)
	}

	if config.TargetDir != "/tmp/test" {
		t.Errorf("Expected TargetDir '/tmp/test', got '%s'", config.TargetDir)
	}

	if config.LogLevel != 1 {
		t.Errorf("Expected LogLevel 1, got %d", config.LogLevel)
	}

	if config.Storage.RemoteType != "localhost" {
		t.Errorf("Expected RemoteType 'localhost', got '%s'", config.Storage.RemoteType)
	}

	if config.Storage.Localhost.Workdir != "/tmp/localhost" {
		t.Errorf("Expected Localhost.Workdir '/tmp/localhost', got '%s'", config.Storage.Localhost.Workdir)
	}

	if !config.Storage.Oss.Enabled {
		t.Error("Expected OSS.Enabled to be true")
	}

	if config.Storage.Oss.Endpoint != "oss-cn-hangzhou.aliyuncs.com" {
		t.Errorf("Expected OSS.Endpoint 'oss-cn-hangzhou.aliyuncs.com', got '%s'", config.Storage.Oss.Endpoint)
	}

	if config.Storage.Oss.AccessKeyID != "test-access-key" {
		t.Errorf("Expected OSS.AccessKeyID 'test-access-key', got '%s'", config.Storage.Oss.AccessKeyID)
	}

	if config.Storage.Oss.AccessKeySecret != "test-secret" {
		t.Errorf("Expected OSS.AccessKeySecret 'test-secret', got '%s'", config.Storage.Oss.AccessKeySecret)
	}

	if config.Storage.Oss.BucketName != "test-bucket" {
		t.Errorf("Expected OSS.BucketName 'test-bucket', got '%s'", config.Storage.Oss.BucketName)
	}

	if config.Storage.Oss.Region != "cn-hangzhou" {
		t.Errorf("Expected OSS.Region 'cn-hangzhou', got '%s'", config.Storage.Oss.Region)
	}

	if config.Storage.Oss.WorkDir != "/oss/work" {
		t.Errorf("Expected OSS.WorkDir '/oss/work', got '%s'", config.Storage.Oss.WorkDir)
	}
}

func TestLoadFromFile_MinimalConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create a minimal config file
	configContent := `
crypto_key: "minimal-key"
target_dir: "/tmp/minimal"
storage:
  remote_type: "localhost"
  localhost:
    work_dir: "/tmp/minimal-localhost"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	config, err := LoadFromFile("config")
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify required values
	if config.CryptoKey != "minimal-key" {
		t.Errorf("Expected CryptoKey 'minimal-key', got '%s'", config.CryptoKey)
	}

	if config.TargetDir != "/tmp/minimal" {
		t.Errorf("Expected TargetDir '/tmp/minimal', got '%s'", config.TargetDir)
	}

	// Verify default values
	if config.LogLevel != 0 {
		t.Errorf("Expected default LogLevel 0, got %d", config.LogLevel)
	}

	if config.Storage.RemoteType != "localhost" {
		t.Errorf("Expected RemoteType 'localhost', got '%s'", config.Storage.RemoteType)
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	tempDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	_, err = LoadFromFile("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent config file")
	}
}

func TestLoadFromFile_InvalidYAML(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create an invalid YAML file
	invalidYAML := `
crypto_key: "test-key"
invalid_yaml: [
  - item1
  - item2
  missing_closing_bracket
`

	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	_, err = LoadFromFile("config")
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestLoadFromFile_OSSConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config with OSS settings
	configContent := `
crypto_key: "oss-test-key"
target_dir: "/tmp/oss-test"
storage:
  remote_type: "oss"
  oss:
    enabled: true
    endpoint: "oss-cn-beijing.aliyuncs.com"
    access_key_id: "LTAI4G..."
    access_key_secret: "secret123"
    bucket_name: "my-bucket"
    region: "cn-beijing"
    workDir: "/data"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	config, err := LoadFromFile("config")
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	if config.Storage.RemoteType != "oss" {
		t.Errorf("Expected RemoteType 'oss', got '%s'", config.Storage.RemoteType)
	}

	if config.Storage.Oss.Endpoint != "oss-cn-beijing.aliyuncs.com" {
		t.Errorf("Expected OSS endpoint 'oss-cn-beijing.aliyuncs.com', got '%s'", config.Storage.Oss.Endpoint)
	}

	if config.Storage.Oss.BucketName != "my-bucket" {
		t.Errorf("Expected OSS bucket 'my-bucket', got '%s'", config.Storage.Oss.BucketName)
	}
}

func TestNewConfig_NoConfigFile(t *testing.T) {
	// Test in a directory without config file
	tempDir := t.TempDir()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	_, err = NewConfig()
	if err == nil {
		t.Error("Expected error when no config file exists")
	}
}

func TestNewConfig_WithValidConfig(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	configContent := `
crypto_key: "new-config-test"
target_dir: "/tmp/new-config"
storage:
  remote_type: "localhost"
  localhost:
    work_dir: "/tmp/new-config-localhost"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	config, err := NewConfig()
	if err != nil {
		t.Fatalf("NewConfig failed: %v", err)
	}

	if config.CryptoKey != "new-config-test" {
		t.Errorf("Expected CryptoKey 'new-config-test', got '%s'", config.CryptoKey)
	}
}

func TestConfig_DefaultValues(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Create config without log_level to test default
	configContent := `
crypto_key: "default-test"
target_dir: "/tmp/default"
storage:
  remote_type: "localhost"
  localhost:
    work_dir: "/tmp/default-localhost"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	config, err := LoadFromFile("config")
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Test default log level
	if config.LogLevel != 0 {
		t.Errorf("Expected default LogLevel 0, got %d", config.LogLevel)
	}
}

func TestConfig_StructTags(t *testing.T) {
	// Test that struct tags are properly defined
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "config.yaml")

	// Use different field names to test mapstructure tags
	configContent := `
crypto_key: "tag-test"
log: "tag-test.log"
target_dir: "/tmp/tag-test"
log_level: 2
storage:
  remote_type: "oss"
  localhost:
    work_dir: "/tmp/tag-localhost"
  oss:
    enabled: false
    endpoint: "tag-endpoint"
    access_key_id: "tag-key-id"
    access_key_secret: "tag-secret"
    bucket_name: "tag-bucket"
    region: "tag-region"
    workDir: "/tag/work"
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to create test config file: %v", err)
	}

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(originalDir)

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	config, err := LoadFromFile("config")
	if err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	// Verify all fields are properly mapped
	if config.CryptoKey != "tag-test" {
		t.Errorf("crypto_key mapping failed: expected 'tag-test', got '%s'", config.CryptoKey)
	}

	if config.Log != "tag-test.log" {
		t.Errorf("log mapping failed: expected 'tag-test.log', got '%s'", config.Log)
	}

	if config.TargetDir != "/tmp/tag-test" {
		t.Errorf("target_dir mapping failed: expected '/tmp/tag-test', got '%s'", config.TargetDir)
	}

	if config.LogLevel != 2 {
		t.Errorf("log_level mapping failed: expected 2, got %d", config.LogLevel)
	}

	if config.Storage.RemoteType != "oss" {
		t.Errorf("remote_type mapping failed: expected 'oss', got '%s'", config.Storage.RemoteType)
	}

	if config.Storage.Localhost.Workdir != "/tmp/tag-localhost" {
		t.Errorf("work_dir mapping failed: expected '/tmp/tag-localhost', got '%s'", config.Storage.Localhost.Workdir)
	}

	if config.Storage.Oss.Enabled {
		t.Error("enabled mapping failed: expected false, got true")
	}

	if config.Storage.Oss.WorkDir != "/tag/work" {
		t.Errorf("workDir mapping failed: expected '/tag/work', got '%s'", config.Storage.Oss.WorkDir)
	}
}
