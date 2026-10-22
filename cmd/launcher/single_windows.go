//go:build windows

package main

import (
	"errors"
	"syscall"
	"unsafe"
)

// errAlreadyRunning 表示已存在同名互斥锁（本进程是第二个实例）。
var errAlreadyRunning = errors.New("launcher: 已有实例在运行")

// mutexName 全局命名互斥锁名（计划 T6）。
const mutexName = "Global\\NinferLauncher"

// errAlreadyExists 是 CreateMutexW 遇到同名互斥锁时的系统错误码（ERROR_ALREADY_EXISTS）。
const errAlreadyExists = 183

// mutexHandle 单实例互斥锁句柄（进程存活期间保持打开，退出时由系统自动释放）。
var mutexHandle uintptr

// acquireSingleInstanceLock 创建单实例命名互斥锁；已存在返回 errAlreadyRunning。
func acquireSingleInstanceLock() error {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	createMutex := kernel32.NewProc("CreateMutexW")
	name, err := syscall.UTF16PtrFromString(mutexName)
	if err != nil {
		return err
	}
	r1, _, lastErr := createMutex.Call(0, 1, uintptr(unsafe.Pointer(name)))
	if lastErr == syscall.Errno(errAlreadyExists) {
		// Windows API 陷阱：同名互斥体已存在时 CreateMutexW 返回已有互斥体的
		// 有效句柄（r1 != 0），必须按 lastErr 判定而非 r1，否则第二个实例不会被拦截。
		return errAlreadyRunning
	}
	if r1 == 0 {
		return lastErr
	}
	mutexHandle = r1
	return nil
}

// showSecondInstanceDialog 弹出中文提示对话框（user32 MessageBoxW，MB_ICONINFORMATION）。
func showSecondInstanceDialog() {
	user32 := syscall.NewLazyDLL("user32.dll")
	msgBox := user32.NewProc("MessageBoxW")
	title, _ := syscall.UTF16PtrFromString("NInfer 启动器")
	msg, _ := syscall.UTF16PtrFromString("NInfer 启动器已在运行，请勿重复启动。")
	msgBox.Call(0, uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(title)), 0x00000040)
}
