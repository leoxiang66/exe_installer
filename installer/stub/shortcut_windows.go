//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func createShortcuts(targetExe, workingDir string, meta InstallMeta) error {
	if _, err := os.Stat(targetExe); err != nil {
		return fmt.Errorf("target exe missing: %w", err)
	}

	if workingDir == "" {
		workingDir = filepath.Dir(targetExe)
	}

	iconPath := targetExe
	var errs []string

	if meta.CreateDesktopShortcut {
		if p, err := desktopDir(); err == nil {
			link := filepath.Join(p, meta.ProductName+".lnk")
			if err2 := createShortcut(link, targetExe, workingDir, iconPath); err2 != nil {
				errs = append(errs, "Desktop:"+err2.Error())
			}
		} else {
			errs = append(errs, "DesktopDir:"+err.Error())
		}
	}

	if meta.CreateStartMenuShortcut {
		if p, err := startMenuDir(meta.ProductName); err == nil {
			if err = os.MkdirAll(p, 0o755); err != nil {
				errs = append(errs, "StartMenu mkdir:"+err.Error())
			} else {
				link := filepath.Join(p, meta.ProductName+".lnk")
				if err2 := createShortcut(link, targetExe, workingDir, iconPath); err2 != nil {
					errs = append(errs, "StartMenu:"+err2.Error())
				}
			}
		} else {
			errs = append(errs, "StartMenuDir:"+err.Error())
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "; "))
	}
	return nil
}

func desktopDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Desktop"), nil
}

func startMenuDir(product string) (string, error) {
	appData := os.Getenv("AppData")
	if appData == "" {
		return "", fmt.Errorf("AppData env empty")
	}
	return filepath.Join(appData, "Microsoft", "Windows", "Start Menu", "Programs", product), nil
}

func createShortcut(linkPath, targetPath, workingDir, iconPath string) error {
	psEsc := func(s string) string {
		return strings.ReplaceAll(s, `'`, `''`)
	}

	script := fmt.Sprintf(`$ErrorActionPreference='Stop';
$W=New-Object -ComObject WScript.Shell;
$S=$W.CreateShortcut('%s');
$S.TargetPath='%s';
$S.WorkingDirectory='%s';
$S.IconLocation='%s';
$S.WindowStyle=1;
$S.Save();`,
		psEsc(linkPath),
		psEsc(targetPath),
		psEsc(workingDir),
		psEsc(iconPath),
	)

	cmd := exec.Command("powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("powershell error: %v output: %s", err, string(out))
	}
	return nil
}