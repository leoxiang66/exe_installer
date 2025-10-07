package main

import (
	"log"

	"exe_installer/installer"
)

func main() {
	err := installer.CreateInstaller(
		"./stub.exe",
		"./myproject.exe",
		"./lol_yuumi_setup_v082.exe",
		installer.Options{
			ProductName:             "lolyuumi",
			ExeName:                 "myproject.exe", // 若你的真实文件名不同，改这里
			CreateDesktopShortcut:   true,
			CreateStartMenuShortcut: true,
			Version:                 "0.8.2",
			ShortcutName:            "悠米助手纯净版",
		},
	)
	if err != nil {
		log.Fatal(err)
	}
}