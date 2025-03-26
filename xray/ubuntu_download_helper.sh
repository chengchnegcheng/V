#!/bin/bash
# Ubuntu系统上手动下载和安装Xray的辅助脚本

set -e

VERSION="v1.8.23"  # 你可以修改为所需版本
DOWNLOAD_URL="https://github.com/XTLS/Xray-core/releases/download/${VERSION}/Xray-linux-64.zip"
INSTALL_DIR="xray/bin/${VERSION}"

echo "=== Ubuntu系统Xray下载辅助脚本 ==="
echo "版本: ${VERSION}"
echo "下载URL: ${DOWNLOAD_URL}"
echo "安装目录: ${INSTALL_DIR}"

# 创建必要的目录
mkdir -p "${INSTALL_DIR}"
mkdir -p xray/downloads

# 下载Xray
echo "正在下载Xray..."
curl -L -o "xray/downloads/xray-${VERSION}.zip" "${DOWNLOAD_URL}"

# 验证下载
if [ ! -f "xray/downloads/xray-${VERSION}.zip" ]; then
    echo "下载失败! 文件不存在"
    exit 1
fi

filesize=$(stat -c%s "xray/downloads/xray-${VERSION}.zip")
echo "下载完成，文件大小: ${filesize} 字节"

if [ "${filesize}" -eq 0 ]; then
    echo "错误: 下载的文件为空!"
    exit 1
fi

# 解压文件
echo "正在解压..."
unzip -o "xray/downloads/xray-${VERSION}.zip" -d "${INSTALL_DIR}"

# 设置权限
chmod +x "${INSTALL_DIR}/xray"

# 检查是否成功
if [ -f "${INSTALL_DIR}/xray" ]; then
    echo "Xray已成功安装到: ${INSTALL_DIR}/xray"
    echo "文件大小: $(stat -c%s "${INSTALL_DIR}/xray") 字节"
    echo "文件类型: $(file "${INSTALL_DIR}/xray")"
else
    echo "错误: 安装失败，未找到xray可执行文件"
    exit 1
fi

echo "你现在可以运行应用程序了: go run main.go"
echo "安装完成!" 