//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/sys/windows/registry"
)

// 判断当前是否为卸载模式：可执行文件名包含 "uninstall"。
func isUninstallMode() bool {
	exe, err := os.Executable()
	if err != nil {
		return false
	}
	name := strings.ToLower(filepath.Base(exe))
	return strings.Contains(name, "uninstall")
}

// createUninstaller 复制当前安装器 stub 为 uninstall.exe (最小方案：复制自身 exe)。
// 由于 stub 已内置安装逻辑，我们生成一个精简的卸载入口可执行：这里采取写一个小的批次逻辑：
// 简化实现：直接复制当前 exe 为 uninstall.exe 并在注册表中区分。实际更优方式是构建一个单独 Uninstaller 源码。
func createUninstaller(installDir string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	dst := filepath.Join(installDir, "uninstall.exe")
	if _, err := os.Stat(dst); err == nil {
		return nil // 已存在
	}
	data, err := os.ReadFile(exe)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0o755)
}

// runUninstall 卸载流程：读取注册表信息推断安装目录（或当前目录），删除快捷方式、注册表再删除目录。
func runUninstall() {
	fmt.Println("正在卸载...")
	// 这里简单：通过可执行所在目录上一级推断安装根目录。
	exe, _ := os.Executable()
	installDir := filepath.Dir(exe)

	// 注册表删除
	// 我们需要 productName：尝试从目录名推断（末级目录名）
	productName := filepath.Base(installDir)
	baseKey := `Software\\` + productName
	uninstallKey := `Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\` + productName
	_ = registry.DeleteKey(registry.CURRENT_USER, uninstallKey)
	_ = registry.DeleteKey(registry.CURRENT_USER, baseKey)

	// 删除快捷方式
	desktopLnk := filepath.Join(userDesktopDir(), productName+".lnk")
	startMenuLnk := filepath.Join(startMenuProgramsDir(), productName, productName+".lnk")
	os.Remove(desktopLnk)
	os.Remove(startMenuLnk)
	os.RemoveAll(filepath.Join(startMenuProgramsDir(), productName))

	// 删除安装目录（延迟删除：因为自身仍在目录内，简单方案：提示手动删除或使用 cmd /c start 新进程延迟）
	// 为简单起见：尝试直接删除除自身外的文件。
	entries, _ := os.ReadDir(installDir)
	for _, e := range entries {
		p := filepath.Join(installDir, e.Name())
		if strings.EqualFold(p, exe) { // 保留自身稍后批处理删除
			continue
		}
		_ = os.RemoveAll(p)
	}
	if err := scheduleSelfDelete(exe, installDir); err != nil {
		fmt.Printf("自删除计划失败（手动删除目录）：%v\n", err)
	} else {
		fmt.Println("已计划删除卸载程序与安装目录...")
	}
	fmt.Println("卸载完成。")
}

// userDesktopDir 返回当前用户桌面目录（简单拼接，不做特殊 Shell 查询）。
func userDesktopDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	return filepath.Join(home, "Desktop")
}

// startMenuProgramsDir 返回开始菜单 Programs 目录。
func startMenuProgramsDir() string {
	appData := os.Getenv("AppData")
	if appData == "" {
		return ""
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs")
}

// scheduleSelfDelete: 使用临时批处理在当前进程退出后循环尝试删除 exe 与安装目录，最后删除自身批处理。
func scheduleSelfDelete(exePath, installDir string) error {
	tempBat := filepath.Join(os.TempDir(), fmt.Sprintf("_uninst_del_%d.bat", os.Getpid()))
	// 构造批处理：等待1-2秒 -> 删除 exe -> 若仍存在则重试 -> 删除目录 -> 删除批处理
	script := fmt.Sprintf(`@echo off\r\n`+
		`set EXE="%s"\r\n`+
		`set DIR="%s"\r\n`+
		`:again\r\n`+
		`ping -n 2 127.0.0.1 >nul\r\n`+
		`del /f /q %s >nul 2>&1\r\n`+
		`if exist %s goto again\r\n`+
		`rmdir /s /q %s >nul 2>&1\r\n`+
		`del /f /q "%s" >nul 2>&1\r\n`,
		exePath, installDir, exePath, exePath, installDir, tempBat,
	)
	if err := os.WriteFile(tempBat, []byte(script), 0o644); err != nil {
		return err
	}
	// 启动批处理（新窗口/后台）。
	_, _ = os.StartProcess("cmd", []string{"cmd", "/c", "start", "", tempBat}, &os.ProcAttr{Files: []*os.File{os.Stdin, os.Stdout, os.Stderr}})
	return nil
}
