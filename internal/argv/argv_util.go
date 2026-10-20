// argv_util.go 双引擎共用的参数文本格式化工具。
package argv

import "strconv"

// formatFloat 契约要求的 float 文本格式：
// strconv.FormatFloat(v, 'g', -1, 64)（0.7→"0.7"，不产生 "0.700000"）。
func formatFloat(v float64) string {
	return strconv.FormatFloat(v, 'g', -1, 64)
}
