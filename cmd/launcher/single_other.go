//go:build !windows

package main

import (
	"errors"
	"fmt"
	"os"
)

// errAlreadyRunning 与 Windows 版同义（非 Windows 无命名互斥锁，恒不触发）。
var errAlreadyRunning = errors.New("launcher: 已有实例在运行")

// acquireSingleInstanceLock 非 Windows 平台 no-op：计划仅要求 Windows 单实例锁。
func acquireSingleInstanceLock() error { return nil }

// showSecondInstanceDialog 非 Windows 平台写 stderr。
func showSecondInstanceDialog() {
	fmt.Fprintln(os.Stderr, "NInfer 启动器已在运行，请勿重复启动。")
}
