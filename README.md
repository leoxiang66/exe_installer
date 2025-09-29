# exe_installer

go build -o stub.exe -ldflags="-H=windowsgui" -trimpath -buildvcs=false -tags windows -v -x -a -gcflags=all=-N -asmflags=all=-trimpath=. -ldflags="-s -w" ./installer/stub

## Windows 构建并嵌入管理员权限 Manifest

在 Windows 上可使用 windres 或 go:embed 方式；当前简化方式：在构建后使用 mt.exe 注入：

```powershell
# 1. 构建 stub.exe
go build -o stub.exe ./installer/stub
# 2. 注入 manifest (需要 Windows SDK mt.exe)
mt.exe -manifest installer\stub\stub.manifest -outputresource:stub.exe;#1
# 3. 重新打包最终安装器
go run ./main.go
```

### 一键构建脚本 build.ps1
示例：
```powershell
# 仅构建 stub.exe 并自动选择 manifest 嵌入方式
./build.ps1 -Mode build

# 构建并打包 (生成最终安装器)
./build.ps1 -Mode package

# 强制使用 windres (需要已安装 mingw-w64)
./build.ps1 -Mode package -ManifestMethod windres

# 使用 mt.exe (Windows SDK) 方式注入
./build.ps1 -Mode package -ManifestMethod mt

# 跳过 manifest (无管理员 UAC)
./build.ps1 -Mode build -ManifestMethod none

# 清理输出
./build.ps1 -Mode clean
```

安装完成后生成的 uninstall.exe 是由 stub 复制而来，因此同样带有管理员请求。

如果想在交叉编译(从 macOS 构建 Windows) 时直接嵌入，可添加一个 .syso 资源文件：
1. 使用 windres 生成 stub_windows.syso：
   # 方式 A: 使用 RC 脚本 (已提供 installer/stub/stub.rc)
   windres installer/stub/stub.rc -O coff -o installer/stub/stub_windows.syso
   # 方式 B: 直接引用 manifest (某些 windres 版本也支持)
2. 放到 installer/stub 目录下，再 go build 即自动链接。










注意：requireAdministrator 会触发 UAC，用户取消则安装/卸载中止。3. 再执行 go build。   rsrc -manifest installer/stub/stub.manifest -o installer/stub/stub_windows.syso   go install github.com/akavel/rsrc@latest2. 或使用第三方工具 rsrc (Go 编写) 生成 .syso：1. 安装 mingw-w64 (获得 windres)；Windows SDK 未安装或 PowerShell 提示找不到 mt.exe 时，可改用：### 没有 mt.exe 的情况注意：requireAdministrator 会触发 UAC，用户取消则安装/卸载中止。