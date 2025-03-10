// main.go
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
		// Если GTA запущена по имени или по полному пути, прекращаем работу
		if checkIfProcessRunning(gtaProcessName) || isProcessRunningByPath(gtaProcessPath) {
			log.Printf("Процесс %s обнаружен (либо по имени, либо по пути). Останавливаем копирование.", gtaProcessName)
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

		// Дополнительно копируем redux файлы в дефолтный backup (C:\Users\{username}\AppData\Local\altv-majestic\backup)
		defaultReduxBackupDir := fmt.Sprintf("C:\\Users\\%s\\AppData\\Local\\altv-majestic\\backup", username)
		var wg3 sync.WaitGroup
		log.Printf("Копирование redux: %s -> %s", cfg.ReduxNew, defaultReduxBackupDir)
		copyDirRecursive(cfg.ReduxNew, defaultReduxBackupDir, &wg3)
		wg3.Wait()

		log.Println("Копирование завершено. Ждём несколько секунд и проверяем снова.")
		time.Sleep(5 * time.Second)
	}
	log.Println("Работа копировальной части завершена.")
}

func main() {
	initLogger()
	configPath := "config.json"
	var cfg *Config
	var err error

	// Если файла конфигурации нет — первичная настройка
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

		auto := prompt("Использовать автозапуск при старте ПК? (y/n): ")
		if auto == "y" || auto == "Y" {
			cfg.AutoRun = true
			exePath, _ := os.Executable()
			if err := setAutoRun(true, exePath); err != nil {
				log.Printf("Ошибка установки автозапуска: %v", err)
			}
		}

		if err := saveConfig(cfg, configPath); err != nil {
			log.Printf("Ошибка сохранения конфигурации: %v", err)
		}
	} else {
		// Повторный запуск: просто загружаем конфигурацию
		cfg, err = loadConfig(configPath)
		if err != nil {
			log.Printf("Ошибка загрузки конфигурации: %v", err)
			return
		}
	}

	// Здесь указываем точное имя процесса, который означает "GTA запущен".
	const gtaProcessName = "GTA5.exe"

	// Получаем имя текущего пользователя для формирования дефолтных путей.
	username := os.Getenv("USERNAME")
	if username == "" {
		username = "DefaultUser"
	}
	defaultGTAPath := fmt.Sprintf("C:\\Users\\%s\\AppData\\Local\\altv-majestic\\backup\\GTA5.exe", username)
	// Запрос пути к GTA5.exe с дефолтным значением.
	gtaProcessPath := prompt(fmt.Sprintf("Введите полный путь к GTA5.exe (по умолчанию: %s): ", defaultGTAPath))
	if gtaProcessPath == "" {
		gtaProcessPath = defaultGTAPath
	}

	// Определяем режим запуска по наличию командного аргумента "-autostart"
	isAutoStart := false
	if len(os.Args) > 1 && os.Args[1] == "-autostart" {
		isAutoStart = true
	}

	if isAutoStart {
		// Автозапуск: сразу скрываем окно и запускаем копировальный цикл
		hideConsole()
		runCopyingLoop(gtaProcessName, gtaProcessPath, cfg, username)
	} else {
		// Ручной запуск: выводим сообщение, запускаем копировальный цикл,
		// затем через 10 секунд завершаем работу.
		fmt.Println("Программа запущена")
		go runCopyingLoop(gtaProcessName, gtaProcessPath, cfg, username)
		time.Sleep(10 * time.Second)
		log.Println("Ручной запуск завершён через 10 секунд.")
	}
}
