package main

import (
	"io/ioutil"
	"log"
	"os/exec"
	"syscall"
)

// setAutoRun устанавливает или удаляет запись автозапуска через реестр Windows.
func setAutoRun(enable bool, exePath string) error {
	const regPath = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	if enable {
		cmd := exec.Command("reg", "add", regPath, "/v", "auto-redux-gunpack", "/t", "REG_SZ", "/d", exePath, "/f")
		if err := cmd.Run(); err != nil {
			log.Printf("Ошибка установки автозапуска: %v", err)
			return err
		}
		batContent := `reg delete HKCU\Software\Microsoft\Windows\CurrentVersion\Run /v auto-redux-gunpack /f`
		if err := ioutil.WriteFile("remove_autorun.bat", []byte(batContent), 0644); err != nil {
			log.Printf("Ошибка создания remove_autorun.bat: %v", err)
			return err
		}
		log.Println("Автозапуск включён. Файл remove_autorun.bat создан.")
	} else {
		cmd := exec.Command("reg", "delete", regPath, "/v", "auto-redux-gunpack", "/f")
		if err := cmd.Run(); err != nil {
			log.Printf("Ошибка удаления автозапуска: %v", err)
		} else {
			log.Println("Запись автозапуска удалена.")
		}
	}
	return nil
}

// hideConsole скрывает окно консоли, чтобы программа продолжала работать в фоне.
func hideConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	getConsoleWindow := kernel32.NewProc("GetConsoleWindow")
	hwnd, _, _ := getConsoleWindow.Call()
	if hwnd == 0 {
		return
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	showWindow := user32.NewProc("ShowWindow")
	const SW_HIDE = 0
	showWindow.Call(hwnd, SW_HIDE)
}
