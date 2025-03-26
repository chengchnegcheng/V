//go:build windows
// +build windows

package xray

import (
	"os/exec"
	"syscall"
)

// configureProcessAttributes 配置平台特定的进程属性
func configureProcessAttributes(cmd *exec.Cmd) {
	// 避免显示命令行窗口
	cmd.SysProcAttr = &syscall.SysProcAttr{
		HideWindow: true,
	}
}
