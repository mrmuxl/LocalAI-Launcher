package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// SaveConfig 将 cfg 序列化为 JSON 并写入 path，必要时创建父目录。
// 任何失败都返回错误，由调用方（gui 层）反馈给用户，不 panic。
func SaveConfig(cfg Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建配置目录 %s 失败: %w", dir, err)
		}
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写入配置文件 %s 失败: %w", path, err)
	}
	return nil
}
