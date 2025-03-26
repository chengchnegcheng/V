# Ubuntu系统编译修复指南

在Ubuntu系统上编译V项目时，可能会遇到以下错误：

```
xray/manager.go:1160:4: unknown field HideWindow in struct literal of type "syscall".SysProcAttr
```

这是因为`HideWindow`字段是Windows平台特有的，在Linux系统上不存在。

## 解决方案一：手动修改代码

1. 打开`xray/manager.go`文件
2. 找到类似以下的代码段：
   ```go
   // 在 Windows 上设置特定的进程属性
   if runtime.GOOS == "windows" {
       // 避免显示命令行窗口
       cmd.SysProcAttr = &syscall.SysProcAttr{
           HideWindow: true,
       }
   }
   ```
3. 将其修改为：
   ```go
   // 设置平台特定的进程属性
   cmd.SysProcAttr = &syscall.SysProcAttr{}
   ```

## 解决方案二：使用提供的修复脚本

1. 我们提供了一个自动修复脚本，执行以下命令：
   ```bash
   chmod +x xray/fix_syscall_linux.sh
   ./xray/fix_syscall_linux.sh
   ```

## 解决方案三：使用平台特定源文件

1. 如果你希望使用更优雅的解决方案，可以：
   ```bash
   # 对于Linux系统
   cp xray/syscall_fix.go.linux xray/syscall_fix.go
   
   # 对于Windows系统
   cp xray/syscall_fix.go.windows xray/syscall_fix.go
   ```
2. 然后继续编译项目

## 后续步骤

修复完成后，就可以正常编译并运行V项目了：

```bash
go build -o v
./v
``` 