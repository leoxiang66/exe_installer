//go:build !windows

package main

// writeRegistry 在非 Windows 平台为无操作，以保持编译通过。
func writeRegistry(meta InstallMeta, installDir, exePath string) error {
	_ = meta
	_ = installDir
	_ = exePath
	return nil
}
