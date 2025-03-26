/**
 * Xray下载辅助工具
 * 用于手动下载Xray版本并放置到正确位置
 * 增强版：与Go自动下载器兼容
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
    'https://ghproxy.com/https://github.com/XTLS/Xray-core/releases/download/',
    'https://gh-proxy.com/https://github.com/XTLS/Xray-core/releases/download/',
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

// 下载文件 - 增强版，支持多种HTTP客户端
async function downloadFile(url, dest) {
  // 确保下载目录存在
  const downloadDir = path.dirname(dest);
  if (!existsSync(downloadDir)) {
    mkdirSync(downloadDir, { recursive: true });
  }
  
  console.log(`开始下载: ${url}`);
  console.log(`保存到: ${dest}`);
  
  // 尝试不同的下载方式
  try {
    await downloadWithHttps(url, dest);
    return;
  } catch (err) {
    console.error(`HTTPS下载失败: ${err.message}`);
    console.log('尝试使用系统命令下载...');
    
    try {
      await downloadWithSystem(url, dest);
      return;
    } catch (err) {
      console.error(`系统命令下载失败: ${err.message}`);
      throw new Error(`所有下载方法均失败: ${err.message}`);
    }
  }
}

// 使用Node.js内置HTTPS模块下载
async function downloadWithHttps(url, dest) {
  const streamPipeline = promisify(pipeline);
  
  return new Promise((resolve, reject) => {
    const request = https.get(url, {
      headers: {
        'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36'
      },
      timeout: 30000 // 30秒超时
    }, async (response) => {
      if (response.statusCode === 302 || response.statusCode === 301) {
        // 处理重定向
        console.log(`重定向到: ${response.headers.location}`);
        try {
          await downloadWithHttps(response.headers.location, dest);
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
    
    request.on('timeout', () => {
      request.destroy();
      reject(new Error('下载超时'));
    });
  });
}

// 使用系统命令下载
async function downloadWithSystem(url, dest) {
  return new Promise((resolve, reject) => {
    let command = '';
    if (process.platform === 'win32') {
      // Windows - 使用PowerShell
      command = `powershell -Command "Invoke-WebRequest -Uri '${url}' -OutFile '${dest}' -UseBasicParsing"`;
    } else {
      // Linux/Mac - 使用curl
      command = `curl -L -o "${dest}" --connect-timeout 30 --retry 3 "${url}"`;
    }
    
    console.log(`执行命令: ${command}`);
    exec(command, (error, stdout, stderr) => {
      if (error) {
        console.error(`命令执行失败: ${error.message}`);
        console.error(`标准错误: ${stderr}`);
        reject(error);
        return;
      }
      
      // 验证文件是否下载成功
      try {
        const stats = fs.statSync(dest);
        if (stats.size === 0) {
          reject(new Error('下载文件大小为0字节'));
          return;
        }
        console.log(`文件下载成功，大小: ${stats.size} 字节`);
        resolve();
      } catch (err) {
        reject(new Error(`验证文件失败: ${err.message}`));
      }
    });
  });
}

// 解压文件 - 增强版，更好的错误处理
async function extractZip(zipPath, destDir) {
  console.log(`解压 ${zipPath} 到 ${destDir}`);
  
  // 确保目标目录存在
  if (!existsSync(destDir)) {
    mkdirSync(destDir, { recursive: true });
  }
  
  // 首先尝试使用Node.js的extract-zip
  try {
    await extract(zipPath, { dir: destDir });
    console.log('使用Node.js解压成功!');
    
    // 设置可执行权限（如果需要）
    await setExecutablePermission(destDir);
    return;
  } catch (err) {
    console.error(`Node.js解压失败: ${err.message}`);
    console.log('尝试使用系统命令解压...');
  }
  
  // 尝试使用系统命令
  try {
    await extractWithSystem(zipPath, destDir);
    console.log('使用系统命令解压成功!');
    
    // 设置可执行权限（如果需要）
    await setExecutablePermission(destDir);
    return;
  } catch (err) {
    console.error(`系统命令解压失败: ${err.message}`);
    throw new Error(`所有解压方法均失败: ${err.message}`);
  }
}

// 使用系统命令解压
async function extractWithSystem(zipPath, destDir) {
  return new Promise((resolve, reject) => {
    let command = '';
    if (process.platform === 'win32') {
      // Windows - 使用PowerShell
      command = `powershell -Command "Expand-Archive -Path '${zipPath}' -DestinationPath '${destDir}' -Force"`;
    } else {
      // Linux/Mac - 使用unzip
      command = `unzip -o "${zipPath}" -d "${destDir}"`;
    }
    
    console.log(`执行命令: ${command}`);
    exec(command, (error, stdout, stderr) => {
      if (error) {
        console.error(`命令执行失败: ${error.message}`);
        console.error(`标准错误: ${stderr}`);
        reject(error);
        return;
      }
      
      console.log(`命令输出: ${stdout}`);
      resolve();
    });
  });
}

// 设置可执行权限
async function setExecutablePermission(destDir) {
  if (process.platform === 'win32') {
    // Windows不需要设置可执行权限
    return;
  }
  
  // Linux/Mac需要设置权限
  const xrayPath = path.join(destDir, 'xray');
  if (existsSync(xrayPath)) {
    return new Promise((resolve, reject) => {
      exec(`chmod +x "${xrayPath}"`, (err) => {
        if (err) {
          console.error(`设置执行权限失败: ${err.message}`);
          reject(err);
          return;
        }
        console.log('已设置可执行权限');
        resolve();
      });
    });
  }
}

// 下载Xray版本 - 增强版
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

  console.log(`开始下载 Xray 版本 ${version}...`);
  
  // 尝试所有镜像
  let downloaded = false;
  let lastError = null;
  
  for (const mirror of config.mirrors) {
    const url = `${mirror}${version}/${fileName}`;
    try {
      console.log(`尝试镜像: ${mirror}`);
      await downloadFile(url, zipPath);
      downloaded = true;
      break;
    } catch (err) {
      console.error(`从 ${mirror} 下载失败: ${err.message}`);
      lastError = err;
      console.log('尝试下一个镜像...');
    }
  }
  
  if (!downloaded) {
    console.error('所有镜像下载失败!');
    if (lastError) {
      console.error(`最后一个错误: ${lastError.message}`);
    }
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