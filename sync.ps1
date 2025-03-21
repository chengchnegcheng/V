# GitHub 同步脚本 - 简化版
Write-Host "正在准备同步到GitHub..." -ForegroundColor Cyan

# 配置
$GITHUB_USERNAME = "your-github-username"
$GITHUB_REPO = "v"
$COMMIT_MESSAGE = "初始提交：V 项目代码"

# 检查Git仓库
if (-not (Test-Path -Path ".git")) {
    Write-Host "初始化Git仓库..." -ForegroundColor Yellow
    git init
}

# 配置远程仓库
$remoteExists = git remote -v
if (-not ($remoteExists -match "origin")) {
    Write-Host "添加远程仓库..." -ForegroundColor Yellow
    git remote add origin "https://github.com/$GITHUB_USERNAME/$GITHUB_REPO.git"
}

# 添加文件并提交
Write-Host "添加文件并提交..." -ForegroundColor Cyan
git add .
git commit -m $COMMIT_MESSAGE

# 推送到GitHub
Write-Host "推送到GitHub..." -ForegroundColor Green
git push -u origin main

# 结果
if ($LASTEXITCODE -eq 0) {
    Write-Host "同步成功!" -ForegroundColor Green
} else {
    Write-Host "同步失败，请检查错误信息" -ForegroundColor Red
} 