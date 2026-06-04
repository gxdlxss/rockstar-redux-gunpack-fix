package app

import (
	"log"
	"os/exec"
	"syscall"
)

// SetAutoRun добавляет или удаляет запись автозапуска в реестре Windows.
func SetAutoRun(enable bool, exePath string) error {
	const regPath = `HKCU\Software\Microsoft\Windows\CurrentVersion\Run`
	if enable {
		cmd := exec.Command("reg", "add", regPath,
			"/v", "auto-redux-gunpack",
			"/t", "REG_SZ",
			"/d", exePath,
			"/f")
		if err := cmd.Run(); err != nil {
			log.Printf("Ошибка установки автозапуска: %v", err)
			return err
		}
		log.Println("Автозапуск включён.")
	} else {
		cmd := exec.Command("reg", "delete", regPath,
			"/v", "auto-redux-gunpack", "/f")
		if err := cmd.Run(); err != nil {
			log.Printf("Ошибка удаления автозапуска: %v", err)
		} else {
			log.Println("Автозапуск отключён.")
		}
	}
	return nil
}

// HideConsole скрывает окно консоли так, что программа продолжает работу в фоне.
func HideConsole() {
	kernel32 := syscall.NewLazyDLL("kernel32.dll")
	hwnd, _, _ := kernel32.NewProc("GetConsoleWindow").Call()
	if hwnd == 0 {
		return
	}
	user32 := syscall.NewLazyDLL("user32.dll")
	user32.NewProc("ShowWindow").Call(hwnd, 0) // SW_HIDE = 0
}

// ShowNotification показывает balloon-уведомление в трее Windows через PowerShell.
func ShowNotification(title, message string) {
	script := `Add-Type -AssemblyName System.Windows.Forms; ` +
		`$n = New-Object System.Windows.Forms.NotifyIcon; ` +
		`$n.Icon = [System.Drawing.SystemIcons]::Information; ` +
		`$n.Visible = $true; ` +
		`$n.ShowBalloonTip(4000, '` + title + `', '` + message + `', [System.Windows.Forms.ToolTipIcon]::Info); ` +
		`Start-Sleep -Milliseconds 4500; $n.Dispose()`
	cmd := exec.Command("powershell", "-WindowStyle", "Hidden", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	_ = cmd.Start()
}
