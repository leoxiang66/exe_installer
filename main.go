package main

import (
	"log"

	"exe_installer/installer"
)

func main() {
	err := installer.CreateInstaller(
		"./stub.exe",
		"./LOL-Yuumi-v070.exe",
		"./setup.exe",
		installer.Options{
			ProductName:             "lolyuumi",
			ExeName:                 "LOL-Yuumi-v070.exe", // 若你的真实文件名不同，改这里
			CreateDesktopShortcut:   true,
			CreateStartMenuShortcut: true,
			Version:                 "1.0.0",
			ShortcutName:            "悠米助手纯净版",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}