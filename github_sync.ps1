# GitHub 同步脚本
# 这个脚本帮助将 V 项目同步到 GitHub 仓库

# 配置 - 请修改以下变量为你的 GitHub 仓库信息
$GITHUB_USERNAME = "your-github-username"
$GITHUB_REPO = "v"  # 仓库名称
$COMMIT_MESSAGE = "初始提交：V 项目代码"

# 检查是否已经初始化 Git 仓库
Write-Host "检查 Git 仓库状态..." -ForegroundColor Cyan
$gitDirExists = Test-Path -Path ".git" -PathType Container

if (-not $gitDirExists) {
    Write-Host "初始化 Git 仓库..." -ForegroundColor Yellow
    git init
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Git 初始化失败，请检查 Git 是否已正确安装" -ForegroundColor Red
        exit 1
    }
}

# 检查远程仓库是否已配置
$remoteConfigured = $false
$remoteOutput = git remote -v 2>$null
if ($remoteOutput -match "origin") {
    $remoteConfigured = $true
}

if (-not $remoteConfigured) {
    Write-Host "配置远程仓库..." -ForegroundColor Yellow
    git remote add origin "https://github.com/$GITHUB_USERNAME/$GITHUB_REPO.git"
    
    if ($LASTEXITCODE -ne 0) {
        Write-Host "远程仓库配置失败" -ForegroundColor Red
        exit 1
    }
}

# 添加所有文件到 Git
Write-Host "将文件添加到 Git..." -ForegroundColor Cyan
git add .

# 提交更改
Write-Host "提交更改..." -ForegroundColor Cyan
git commit -m "$COMMIT_MESSAGE"

# 推送到 GitHub
Write-Host "推送到 GitHub..." -ForegroundColor Green
git push -u origin main

if ($LASTEXITCODE -eq 0) {
    Write-Host "成功同步到 GitHub 仓库: https://github.com/$GITHUB_USERNAME/$GITHUB_REPO" -ForegroundColor Green
    Write-Host "项目已成功同步!" -ForegroundColor Green
} else {
    Write-Host "推送到 GitHub 失败，请检查错误信息" -ForegroundColor Red
    Write-Host "可能需要先在 GitHub 上创建仓库，或检查认证信息是否正确" -ForegroundColor Yellow
    
    # 提供解决方案建议
    Write-Host "`n解决方案建议:" -ForegroundColor Cyan
    Write-Host "1. 确保已在 GitHub 上创建名为 $GITHUB_REPO 的仓库" -ForegroundColor White
    Write-Host "2. 如果遇到认证问题，请尝试使用个人访问令牌 (PAT):" -ForegroundColor White
    Write-Host "   git remote set-url origin https://$GITHUB_USERNAME:YOUR_PAT@github.com/$GITHUB_USERNAME/$GITHUB_REPO.git" -ForegroundColor White
    Write-Host "3. 也可以尝试使用 GitHub CLI 进行身份验证:" -ForegroundColor White
    Write-Host "   gh auth login" -ForegroundColor White
    Write-Host "4. 如果存在冲突，可能需要先执行 git pull 获取远程更改" -ForegroundColor White
}

Write-Host "`n如需修改远程仓库地址，请使用以下命令:" -ForegroundColor Cyan
Write-Host "git remote set-url origin https://github.com/YOUR_USERNAME/YOUR_REPO.git" -ForegroundColor White 