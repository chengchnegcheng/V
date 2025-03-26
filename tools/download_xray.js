/**
 * Xray下载辅助工具
 * 用于手动下载Xray版本并放置到正确位置
 */

const fs = require('fs');
const path = require('path');
const https = require('https');
const { createWriteStream, existsSync, mkdirSync } = require('fs');
const { exec } = require('child_process');
const { pipeline } = require('stream');
const { promisify } = require('util');
const extract = require('extract-zip');

// 配置
const config = {
  // 项目根目录，根据实际情况修改
  rootDir: path.resolve(__dirname, '..'),
  // Xray二进制文件目录
  binDir: path.resolve(__dirname, '../xray/bin'),
  // 下载目录
  downloadDir: path.resolve(__dirname, '../xray/downloads'),
  // 支持的版本
  versions: [
    'v1.8.24', 'v1.8.23', 'v1.8.22', 'v1.8.21', 'v1.8.20',
    'v1.8.19', 'v1.8.18', 'v1.8.17', 'v1.8.16', 'v1.8.15',
    'v25.3.6', 'v25.3.3', 'v25.2.21', 'v25.2.18', 'v25.1.30'
  ],
  // 镜像URLs（按优先级排序）
  mirrors: [
    'https://mirror.ghproxy.com/https://github.com/XTLS/Xray-core/releases/download/',
    'https://github.mirrorz.org/XTLS/Xray-core/releases/download/',
    'https://github.com/XTLS/Xray-core/releases/download/'
  ]
};

// 创建必要的目录
function createDirectories() {
  if (!existsSync(config.binDir)) {
    mkdirSync(config.binDir, { recursive: true });
    console.log(`创建目录: ${config.binDir}`);
  }
  
  if (!existsSync(config.downloadDir)) {
    mkdirSync(config.downloadDir, { recursive: true });
    console.log(`创建目录: ${config.downloadDir}`);
  }
}

// 获取平台信息
function getPlatformInfo() {
  const platform = process.platform;
  const arch = process.arch;
  
  let osName = platform;
  let archName = arch;
  
  // 转换操作系统名称
  if (platform === 'win32') {
    osName = 'windows';
  } else if (platform === 'darwin') {
    osName = 'macos';
  }
  
  // 转换架构名称
  if (arch === 'x64') {
    archName = '64';
  } else if (arch === 'ia32' || arch === 'x86') {
    archName = '32';
  } else if (arch === 'arm64') {
    archName = 'arm64-v8a';
  } else if (arch === 'arm') {
    archName = 'arm32-v7a';
  }
  
  return { os: osName, arch: archName };
}

// 下载文件
async function downloadFile(url, dest) {
  const streamPipeline = promisify(pipeline);
  
  return new Promise((resolve, reject) => {
    console.log(`开始下载: ${url}`);
    console.log(`保存到: ${dest}`);
    
    const request = https.get(url, async (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // 处理重定向
        console.log(`重定向到: ${response.headers.location}`);
        try {
          await downloadFile(response.headers.location, dest);
          resolve();
        } catch (err) {
          reject(err);
        }
        return;
      }
      
      if (response.statusCode !== 200) {
        reject(new Error(`下载失败，状态码: ${response.statusCode}`));
        return;
      }
      
      try {
        await streamPipeline(response, createWriteStream(dest));
        console.log('下载完成!');
        resolve();
      } catch (err) {
        reject(err);
      }
    });
    
    request.on('error', (err) => {
      console.error(`下载错误: ${err.message}`);
      reject(err);
    });
  });
}

// 解压文件
async function extractZip(zipPath, destDir) {
  console.log(`解压 ${zipPath} 到 ${destDir}`);
  try {
    await extract(zipPath, { dir: destDir });
    console.log('解压完成!');
    
    // 如果是Windows，确保xray.exe有执行权限
    if (process.platform === 'win32') {
      const xrayExePath = path.join(destDir, 'xray.exe');
      if (existsSync(xrayExePath)) {
        console.log(`Xray可执行文件找到: ${xrayExePath}`);
      } else {
        console.error(`错误: ${xrayExePath} 不存在!`);
      }
    } else {
      // 在Linux/macOS上设置执行权限
      const xrayPath = path.join(destDir, 'xray');
      if (existsSync(xrayPath)) {
        exec(`chmod +x "${xrayPath}"`, (err) => {
          if (err) console.error(`设置执行权限失败: ${err.message}`);
          else console.log('已设置执行权限');
        });
      }
    }
  } catch (err) {
    console.error(`解压失败: ${err.message}`);
    throw err;
  }
}

// 下载Xray版本
async function downloadXrayVersion(version) {
  if (!config.versions.includes(version)) {
    console.error(`错误: 不支持的版本 ${version}`);
    console.log(`支持的版本: ${config.versions.join(', ')}`);
    return false;
  }
  
  const { os, arch } = getPlatformInfo();
  console.log(`检测到系统: ${os}-${arch}`);
  
  const fileName = `Xray-${os}-${arch}.zip`;
  const versionDir = path.join(config.binDir, version);
  const zipPath = path.join(config.downloadDir, `${version}-${fileName}`);
  
  // 创建版本目录
  if (!existsSync(versionDir)) {
    mkdirSync(versionDir, { recursive: true });
    console.log(`创建目录: ${versionDir}`);
  }
  
  // 检查目标文件是否已存在
  const execFile = process.platform === 'win32' ? 'xray.exe' : 'xray';
  const execPath = path.join(versionDir, execFile);
  
  if (existsSync(execPath)) {
    console.log(`${version} 已存在: ${execPath}`);
    return true;
  }
  
  // 尝试所有镜像
  let downloaded = false;
  for (const mirror of config.mirrors) {
    const url = `${mirror}${version}/${fileName}`;
    try {
      await downloadFile(url, zipPath);
      downloaded = true;
      break;
    } catch (err) {
      console.error(`从 ${mirror} 下载失败: ${err.message}`);
      console.log('尝试下一个镜像...');
    }
  }
  
  if (!downloaded) {
    console.error('所有镜像下载失败!');
    return false;
  }
  
  // 解压文件
  try {
    await extractZip(zipPath, versionDir);
    console.log(`${version} 安装成功到 ${versionDir}`);
    return true;
  } catch (err) {
    console.error(`解压失败: ${err.message}`);
    return false;
  }
}

// 主函数
async function main() {
  // 创建必要的目录
  createDirectories();
  
  // 获取命令行参数
  const args = process.argv.slice(2);
  const version = args[0] || 'v1.8.24';  // 默认版本
  
  console.log(`准备下载 Xray ${version}...`);
  
  try {
    const success = await downloadXrayVersion(version);
    if (success) {
      console.log(`\n✅ Xray ${version} 下载并安装成功!`);
    } else {
      console.error(`\n❌ Xray ${version} 下载或安装失败!`);
      process.exit(1);
    }
  } catch (err) {
    console.error(`\n❌ 发生错误: ${err.message}`);
    process.exit(1);
  }
}

// 运行主函数
main(); 