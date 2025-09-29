//go:build !windows

package main

func isUninstallMode() bool              { return false }
func createUninstaller(dir string) error { _ = dir; return nil }
func runUninstall()                      {}
