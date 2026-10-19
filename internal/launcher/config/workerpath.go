// workerpath.go 持久化用户通过文件选择框选定的 worker 可执行文件路径。
//
// 与 34 字段的业务配置（config.json）分离：worker 路径是启动器自身设置，
// 不参与业务校验与表单，单独存档为 UserConfigDir()/NinferLauncher/worker.json。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WorkerPathRecord worker.json 的顶层结构（camelCase，与 Config tag 风格一致）。
type WorkerPathRecord struct {
	WorkerPath string `json:"workerPath"`
}

// WorkerSettingsPath 返回 worker 路径存档文件路径：
// UserConfigDir()/NinferLauncher/worker.json（与主配置同目录）。
func WorkerSettingsPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("获取用户配置目录失败: %w", err)
	}
	return filepath.Join(dir, "NinferLauncher", "worker.json"), nil
}

// LoadWorkerPath 读取已保存的 worker 路径。
//   - 存档文件不存在：返回空串，不报错（视为首次使用）。
//   - JSON 损坏：返回错误，由调用方决定回退路径。
func LoadWorkerPath() (string, error) {
	path, err := WorkerSettingsPath()
	if err != nil {
		return "", err
	}
	return loadWorkerPathFrom(path)
}

// SaveWorkerPathDefault 将 p 写入默认存档文件，返回实际写入路径。
func SaveWorkerPathDefault(p string) (string, error) {
	path, err := WorkerSettingsPath()
	if err != nil {
		return "", err
	}
	if err := saveWorkerPathTo(path, p); err != nil {
		return path, err
	}
	return path, nil
}

// loadWorkerPathFrom 从指定文件读取 worker 路径（测试可注入路径）。
// 文件不存在返回空串不报错；出现其他 IO/JSON 错误时返回错误。
func loadWorkerPathFrom(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // 缺文件 → 首次使用
		}
		return "", fmt.Errorf("读取 worker 路径存档 %s 失败: %w", path, err)
	}

	var rec WorkerPathRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return "", fmt.Errorf("解析 worker 路径存档 %s 失败: %w", path, err)
	}
	return rec.WorkerPath, nil
}

// saveWorkerPathTo 将 worker 路径写入指定文件，必要时创建父目录（测试可注入路径）。
func saveWorkerPathTo(path, p string) error {
	data, err := json.MarshalIndent(WorkerPathRecord{WorkerPath: p}, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化 worker 路径失败: %w", err)
	}

	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("创建配置目录 %s 失败: %w", dir, err)
		}
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("写入 worker 路径存档 %s 失败: %w", path, err)
	}
	return nil
}
