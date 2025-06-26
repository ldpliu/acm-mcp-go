package config

import (
	"fmt"
)

// Config 表示应用程序配置
type Config struct {
	// Kubeconfig 文件路径
	RepoPath string
	// 是否启用资源创建操作
	RepoType string
}

// NewConfig 从命令行参数创建配置
func NewConfig(repoPath string, repoType string) *Config {
	return &Config{
		RepoPath: repoPath,
		RepoType: repoType,
	}
}

func (c *Config) Validate() error {
	// 检查 kubeconfig 是否可访问
	if c.RepoPath == "" {
		return fmt.Errorf("无法访问 repopath 文件")
	}
	return nil
}
