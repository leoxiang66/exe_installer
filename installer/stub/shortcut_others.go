//go:build !windows

package main

// 非 Windows 平台占位实现
func createShortcuts(targetExe, workingDir string, meta InstallMeta) error {
	return nil
}