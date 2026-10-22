package config

import (
	"fmt"
	"os"
)

// saveJSONAtomic 先写临时文件再 rename：进程中途被杀时原文件保持完整，
// 不会留下半写损坏的 JSON（下次 Load 不会再报 invalid JSON）。
// rename 失败时尽力删除遗留临时文件，清理失败仅附注到错误上。
func saveJSONAtomic(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	if err := os.Rename(tmp, path); err != nil {
		if rmErr := os.Remove(tmp); rmErr != nil {
			return fmt.Errorf("%w（临时文件清理亦失败: %v）", err, rmErr)
		}
		return err
	}
	return nil
}
