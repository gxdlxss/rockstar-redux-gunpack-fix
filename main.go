package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

func runCopyingLoop(gtaProcessName, gtaProcessPath string, cfg *Config, username string) {
	log.Printf("Будем копировать файлы каждые несколько секунд, пока не запустится процесс %s (по имени или по пути)", gtaProcessName)
	for {
		// Если GTA запущена (по имени или по полному пути), завершаем цикл
		if checkIfProcessRunning(gtaProcessName) || isProcessRunningByPath(gtaProcessPath) {
			log.Printf("Процесс %s обнаружен. Останавливаем копирование.", gtaProcessName)
			break
		}

		log.Println("GTA не запущен. Выполняем копирование...")

		// Копирование для gunpack
		var wg sync.WaitGroup
		log.Printf("Копирование gunpack: %s -> %s", cfg.GunpackNew, cfg.GunpackOld)
		copyDirRecursive(cfg.GunpackNew, cfg.GunpackOld, &wg)
		wg.Wait()

		// Копирование для redux по пути из конфигурации
		var wg2 sync.WaitGroup
		log.Printf("Копирование redux: %s -> %s", cfg.ReduxNew, cfg.ReduxOld)
		copyDirRecursive(cfg.ReduxNew, cfg.ReduxOld, &wg2)
		wg2.Wait()

		// Дополнительно копируем redux файлы в дефолтный backup:
		defaultReduxBackupDir := fmt.Sprintf("C:\\Users\\%s\\AppData\\Local\\altv-majestic\\backup", username)
		var wg3 sync.WaitGroup
		log.Printf("Копирование redux: %s -> %s", cfg.ReduxNew, defaultReduxBackupDir)
		copyDirRecursive(cfg.ReduxNew, defaultReduxBackupDir, &wg3)
		wg3.Wait()

		log.Println("Копирование завершено. Ждем 5 секунд...")
		time.Sleep(5 * time.Second)
	}
	log.Println("Работа копировального цикла завершена.")
}

func main() {
	initLogger()
	configPath := "config.json"
	var cfg *Config
	var err error

	// Если конфиг отсутствует – первичный запуск (настройка)
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		log.Println("Файл конфигурации не найден. Первичный запуск и настройка конфигурации.")
		cfg = &Config{}

		fmt.Println("Первичная настройка директорий:")
		fmt.Println("  gunpack-new: директория с новыми файлами для gunpack, копируются в gunpack-old")
		fmt.Println("  gunpack-old: директория, куда копируются файлы gunpack")
		fmt.Println("  redux-new:   директория с новыми файлами для redux, копируются в redux-old")
		fmt.Println("  redux-old:   директория, куда копируются файлы redux")

		cfg.GunpackNew = prompt("Введите путь к директории gunpack-new: ")
		cfg.GunpackOld = prompt("Введите путь к директории gunpack-old: ")
		cfg.ReduxNew = prompt("Введите путь к директории redux-new: ")
		cfg.ReduxOld = prompt("Введите путь к директории redux-old: ")

		// Запрашиваем путь к GTA5.exe; если пользователь нажмет Enter, будет использовано дефолтное значение
		cfg.GtaExePath = prompt(fmt.Sprintf("Введите полный путь к GTA5.exe (по умолчанию: %s): ", defaultGTAPathForUser()))
		if cfg.GtaExePath == "" {
			cfg.GtaExePath = defaultGTAPathForUser()
		}

		auto := prompt("Использовать автозапуск при старте ПК? (y/n): ")
		if auto == "y" || auto == "Y" {
			cfg.AutoRun = true
			exePath, _ := os.Executable()
			// Прописываем автозапуск с параметром -autostart
			if err := setAutoRun(true, exePath+" -autostart"); err != nil {
				log.Printf("Ошибка установки автозапуска: %v", err)
			}
		}

		if err := saveConfig(cfg, configPath); err != nil {
			log.Printf("Ошибка сохранения конфигурации: %v", err)
		}

		// После настройки запускаем копировальный цикл
		runCopyingLoop("GTA5.exe", cfg.GtaExePath, cfg, os.Getenv("USERNAME"))
		log.Println("Программа завершена (первичный запуск).")
		return
	} else {
		// Конфиг уже существует – просто загружаем его
		cfg, err = loadConfig(configPath)
		if err != nil {
			log.Printf("Ошибка загрузки конфигурации: %v", err)
			return
		}
	}

	// Определяем режим запуска по наличию аргумента "-autostart"
	isAutoStart := false
	if len(os.Args) > 1 && os.Args[1] == "-autostart" {
		isAutoStart = true
	}

	// Если в конфиге не указан путь к GTA5.exe, используем дефолтное значение
	username := os.Getenv("USERNAME")
	if username == "" {
		username = "DefaultUser"
	}
	defaultGTAPath := defaultGTAPathForUser()
	if cfg.GtaExePath == "" {
		cfg.GtaExePath = defaultGTAPath
	}

	if isAutoStart {
		// Автозапуск: скрываем консоль сразу и запускаем копировальный цикл
		hideConsole()
		runCopyingLoop("GTA5.exe", cfg.GtaExePath, cfg, username)
		log.Println("Автозапуск: программа завершилась.")
	} else {
		// Ручной запуск: выводим сообщение и запускаем копировальный цикл в горутине,
		// затем через 10 секунд скрываем консоль, а процесс остается работающим.
		fmt.Println("Программа запущена, все изменения вносите в config.json")
		go runCopyingLoop("GTA5.exe", cfg.GtaExePath, cfg, username)
		time.Sleep(10 * time.Second)
		log.Println("Ручной запуск: скрываем консоль через 10 секунд.")
		hideConsole()
		// Блокируем главный поток, чтобы программа не завершалась
		select {}
	}
}

func defaultGTAPathForUser() string {
	username := os.Getenv("USERNAME")
	if username == "" {
		username = "DefaultUser"
	}
	return fmt.Sprintf("C:\\Users\\%s\\AppData\\Local\\altv-majestic\\backup\\GTA5.exe", username)
}
