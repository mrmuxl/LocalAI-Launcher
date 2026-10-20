// workerpath.go 持久化用户通过文件选择框选定的 worker 可执行文件路径。
//
// 与 32 字段的业务配置（config.json）分离：worker 路径是启动器自身设置，
// 不参与业务校验与表单，单独存档为启动器 exe 同级目录下的 worker.json。
// 全部配置文件（config.json/llama.json/worker.json）均便携式存放于
// 启动器 exe 同级目录。
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// exeDir 返回启动器可执行文件所在目录（便携式配置根）。
// os.Executable 失败时回退当前工作目录，再失败回退 "."。
func exeDir() string {
	exe, err := os.Executable()
	if err != nil {
		if wd, wdErr := os.Getwd(); wdErr == nil {
			return wd
		}
		return "."
	}
	return filepath.Dir(exe)
}

// WorkerPathRecord worker.json 的顶层结构（camelCase，与 Config tag 风格一致）。
// 双引擎各存一份 worker 路径；旧存档只有 workerPath 字段，读取时
// LlamaWorkerPath 为空串（视为未指定），向后兼容。
type WorkerPathRecord struct {
	WorkerPath      string `json:"workerPath"`
	LlamaWorkerPath string `json:"llamaWorkerPath,omitempty"`
}

// WorkerSettingsPath 返回 worker 路径存档文件路径：
// 启动器 exe 同级目录下的 worker.json。
func WorkerSettingsPath() string {
	return filepath.Join(exeDir(), "worker.json")
}

// ConfigPathFor 返回启动器 exe 同级目录下的指定配置文件名：
// 配置持久化文件与启动器同目录（便携式，整个目录拷走即整套配置）。
// filename 如 "config.json"（ninfer）或 "llama.json"（llama.cpp）。
func ConfigPathFor(filename string) string {
	return filepath.Join(exeDir(), filename)
}

// LoadWorkerPath 读取已保存的 worker 路径。
//   - 存档文件不存在：返回空串，不报错（视为首次使用）。
//   - JSON 损坏：返回错误，由调用方决定回退路径。
func LoadWorkerPath() (string, error) {
	rec, err := loadWorkerRecord()
	if err != nil {
		return "", err
	}
	return rec.WorkerPath, nil
}

// LoadLlamaWorkerPath 读取已保存的 llama.cpp worker 路径（语义同 LoadWorkerPath）。
func LoadLlamaWorkerPath() (string, error) {
	rec, err := loadWorkerRecord()
	if err != nil {
		return "", err
	}
	return rec.LlamaWorkerPath, nil
}

// loadWorkerRecord 读取完整存档记录（测试与两个 Load 入口共用）。
func loadWorkerRecord() (WorkerPathRecord, error) {
	return loadWorkerRecordFrom(WorkerSettingsPath())
}

// SaveWorkerPathDefault 将 p 写入默认存档文件的 WorkerPath 字段：
// 读出现有记录、只更新该字段后整体写回，保留 LlamaWorkerPath。
func SaveWorkerPathDefault(p string) (string, error) {
	return saveWorkerPathField(func(rec *WorkerPathRecord) { rec.WorkerPath = p })
}

// SaveLlamaWorkerPath 将 llama.cpp worker 路径写入默认存档文件的
// LlamaWorkerPath 字段：读出现有记录、只更新该字段后整体写回，
// 保留 ninfer 的 WorkerPath。
func SaveLlamaWorkerPath(p string) (string, error) {
	return saveWorkerPathField(func(rec *WorkerPathRecord) { rec.LlamaWorkerPath = p })
}

// saveWorkerPathField 读出存档、应用字段更新函数后整体写回。
func saveWorkerPathField(apply func(rec *WorkerPathRecord)) (string, error) {
	path := WorkerSettingsPath()
	rec, err := loadWorkerRecordFrom(path)
	if err != nil {
		return path, err
	}
	apply(&rec)
	if err := saveWorkerRecordTo(path, rec); err != nil {
		return path, err
	}
	return path, nil
}

// loadWorkerRecordFrom 从指定文件读取完整存档记录（测试可注入路径）。
// 文件不存在返回零值记录不报错；出现其他 IO/JSON 错误时返回错误。
func loadWorkerRecordFrom(path string) (WorkerPathRecord, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return WorkerPathRecord{}, nil // 缺文件 → 首次使用
		}
		return WorkerPathRecord{}, fmt.Errorf("读取 worker 路径存档 %s 失败: %w", path, err)
	}

	var rec WorkerPathRecord
	if err := json.Unmarshal(data, &rec); err != nil {
		return WorkerPathRecord{}, fmt.Errorf("解析 worker 路径存档 %s 失败: %w", path, err)
	}
	return rec, nil
}

// saveWorkerRecordTo 将完整存档记录写入指定文件，必要时创建父目录（测试可注入路径）。
func saveWorkerRecordTo(path string, rec WorkerPathRecord) error {
	data, err := json.MarshalIndent(rec, "", "  ")
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
