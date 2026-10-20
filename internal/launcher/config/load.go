package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ConfigPathFor 返回推理程序所在目录下的 config.json：
// 配置持久化文件跟随 ninfer.exe（便携式，换程序目录即换一套配置）。
// workerPath 为相对路径（如默认值 ninfer.exe）时，随运行时工作目录解析。
func ConfigPathFor(workerPath string) string {
	return filepath.Join(filepath.Dir(workerPath), "config.json")
}

// LoadConfig 从 path 读取 JSON 配置。
//   - 文件不存在：返回 DefaultConfig()，不报错。
//   - JSON 损坏：返回错误，且不覆盖原文件（本函数只读，不写）。
//   - 文件存在但字段缺失：以默认值为底，文件中出现的字段覆盖对应默认值。
func LoadConfig(path string) (Config, error) {
	cfg := DefaultConfig()

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil // 缺文件→默认值
		}
		return Config{}, fmt.Errorf("读取配置文件 %s 失败: %w", path, err)
	}

	if err := json.Unmarshal(data, &cfg); err != nil {
		// 坏 JSON→返回错误；ReadFile 不修改文件，原文件保持不变
		return Config{}, fmt.Errorf("解析配置文件 %s 失败: %w", path, err)
	}

	return cfg, nil
}
