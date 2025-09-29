//go:build windows

package main

import (
	"fmt"
	"path/filepath"

	"golang.org/x/sys/windows/registry"
)

// writeRegistry 写入安装与卸载信息到当前用户注册表。
// Keys:
//  1. HKCU\Software\<ProductName> : InstallDir, ExePath, Version
//  2. HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\<ProductName>
//     以便显示在“应用和功能”/“卸载程序”列表。
func writeRegistry(meta InstallMeta, installDir, exePath string) error {
	if meta.ProductName == "" {
		return fmt.Errorf("empty product name")
	}
	if installDir == "" || exePath == "" {
		return fmt.Errorf("empty paths")
	}

	basePath := `Software\\` + meta.ProductName
	if err := setValues(registry.CURRENT_USER, basePath, map[string]any{
		"InstallDir": installDir,
		"ExePath":    exePath,
		"Version":    meta.Version,
	}); err != nil {
		return fmt.Errorf("write base key: %w", err)
	}

	uninstallPath := `Software\\Microsoft\\Windows\\CurrentVersion\\Uninstall\\` + meta.ProductName
	uninstallString := exePath // 若未来有独立卸载器，可替换
	if err := setValues(registry.CURRENT_USER, uninstallPath, map[string]any{
		"DisplayName":          meta.ProductName,
		"DisplayVersion":       meta.Version,
		"InstallLocation":      installDir,
		"Publisher":            "",
		"UninstallString":      uninstallString,
		"QuietUninstallString": uninstallString + " /S", // 预留
		"DisplayIcon":          exePath + ",0",
		"NoModify":             uint32(1),
		"NoRepair":             uint32(1),
		"InstallSource":        filepath.Dir(exePath),
	}); err != nil {
		return fmt.Errorf("write uninstall key: %w", err)
	}

	return nil
}

func setValues(root registry.Key, path string, kv map[string]any) error {
	k, _, err := registry.CreateKey(root, path, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	for name, v := range kv {
		switch val := v.(type) {
		case string:
			if err := k.SetStringValue(name, val); err != nil {
				return err
			}
		case uint32:
			if err := k.SetDWordValue(name, val); err != nil {
				return err
			}
		default:
			// ignore unsupported types
		}
	}
	return nil
}
