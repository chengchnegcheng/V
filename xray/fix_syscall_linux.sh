#!/bin/bash
# 这个脚本用于修复Linux上的syscall.SysProcAttr.HideWindow错误

# 检查manager.go是否包含HideWindow
if grep -q "HideWindow" xray/manager.go; then
  echo "应用补丁以修复syscall.SysProcAttr.HideWindow错误..."
  sed -i 's/cmd.SysProcAttr = \&syscall.SysProcAttr{/cmd.SysProcAttr = \&syscall.SysProcAttr{/' xray/manager.go
  sed -i '/HideWindow: true,/d' xray/manager.go
  echo "补丁应用成功"
else
  echo "无需应用补丁"
fi 