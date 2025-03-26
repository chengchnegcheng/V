//go:build !windows
// +build !windows

package xray

import (
	"os/exec"
	"syscall"
)

// configureProcessAttributes 配置平台特定的进程属性
func configureProcessAttributes(cmd *exec.Cmd) {
	// 在类Unix系统上使用默认的SysProcAttr设置
	cmd.SysProcAttr = &syscall.SysProcAttr{}
}
