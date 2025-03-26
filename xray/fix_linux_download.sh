#!/bin/bash
# Xray Linux版本下载修复脚本

set -e

VERSION="v1.8.23"  # 目标版本，可以根据需要修改
DOWNLOAD_URL="https://mirror.ghproxy.com/https://github.com/XTLS/Xray-core/releases/download/${VERSION}/Xray-linux-64.zip"
DOWNLOAD_DIR="xray/downloads"
INSTALL_DIR="xray/bin/${VERSION}"

echo "=== Xray Linux 版本下载修复脚本 ==="
echo "版本: ${VERSION}"
echo "下载URL: ${DOWNLOAD_URL}"
echo "下载目录: ${DOWNLOAD_DIR}"
echo "安装目录: ${INSTALL_DIR}"

# 创建必要的目录
mkdir -p "${DOWNLOAD_DIR}"
mkdir -p "${INSTALL_DIR}"

# 下载 Xray Linux 版本
echo "正在下载 Xray Linux 版本..."
curl -L -o "${DOWNLOAD_DIR}/Xray-linux-64.zip" "${DOWNLOAD_URL}"

# 验证下载
if [ ! -f "${DOWNLOAD_DIR}/Xray-linux-64.zip" ]; then
    echo "下载失败! 文件不存在"
    exit 1
fi

filesize=$(stat -c%s "${DOWNLOAD_DIR}/Xray-linux-64.zip" 2>/dev/null || stat -f%z "${DOWNLOAD_DIR}/Xray-linux-64.zip")
echo "下载完成，文件大小: ${filesize} 字节"

if [ "${filesize}" -eq 0 ]; then
    echo "错误: 下载的文件为空!"
    exit 1
fi

# 复制下载的文件到app.asar目录以便Node.js工具可以访问
echo "复制下载的文件到app.asar目录..."
cp "${DOWNLOAD_DIR}/Xray-linux-64.zip" ./

# 使用Node.js工具解压
echo "正在使用Node.js工具解压..."
if [ -d "tools" ] && [ -f "tools/download_xray.js" ]; then
    cd tools
    if [ ! -d "node_modules" ] || [ ! -f "node_modules/extract-zip/index.js" ]; then
        echo "安装依赖..."
        npm install
    fi
    echo "运行Node.js Xray下载工具..."
    node download_xray.js "${VERSION}"
    cd ..
else
    echo "错误: 找不到tools/download_xray.js"
    echo "使用系统命令尝试解压..."
    unzip -o "${DOWNLOAD_DIR}/Xray-linux-64.zip" -d "${INSTALL_DIR}" || {
        echo "系统unzip命令失败，请安装Node.js和extract-zip模块"
        exit 1
    }
fi

# 检查解压后的文件
if [ -f "${INSTALL_DIR}/xray" ]; then
    chmod +x "${INSTALL_DIR}/xray"
    echo "✅ 安装成功! Xray可执行文件位于: ${INSTALL_DIR}/xray"
    echo "文件大小: $(stat -c%s "${INSTALL_DIR}/xray" 2>/dev/null || stat -f%z "${INSTALL_DIR}/xray") 字节"
else
    echo "❌ 安装失败! 未找到xray可执行文件"
    exit 1
fi

echo "完成!" 