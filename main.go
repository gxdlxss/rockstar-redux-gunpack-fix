package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

// defaultGTAPathForUser возвращает дефолтный путь к GTA5.exe с использованием USERNAME.
func defaultGTAPathForUser() string {
	username := os.Getenv("USERNAME")
	if username == "" {
		username = "DefaultUser"
	}
	return fmt.Sprintf("C:\\Users\\%s\\AppData\\Local\\altv-majestic\\backup\\GTA5.exe", username)
}

// runCopyingLoop запускает цикл копирования.
// Если процесс GTA запущен, он ждет 5 секунд и продолжает проверку;
// если GTA не запущен, выполняется копирование директорий.
func runCopyingLoop(gtaProcessName, gtaProcessPath string, cfg *Config, username string) {
	log.Printf("Запускается цикл копирования: проверяем процесс %s (по имени или по пути)", gtaProcessName)
	for {
		if checkIfProcessRunning(gtaProcessName) || isProcessRunningByPath(gtaProcessPath) {
			log.Printf("Процесс %s обнаружен. Ждем 5 секунд перед повторной проверкой...", gtaProcessName)
			time.Sleep(5 * time.Second)
			continue
		}

		log.Println("GTA не запущен. Выполняем копирование...")

		var wg sync.WaitGroup
		log.Printf("Копирование gunpack: %s -> %s", cfg.GunpackNew, cfg.GunpackOld)
		copyDirRecursive(cfg.GunpackNew, cfg.GunpackOld, &wg)
		wg.Wait()

		var wg2 sync.WaitGroup
		log.Printf("Копирование redux: %s -> %s", cfg.ReduxNew, cfg.ReduxOld)
		copyDirRecursive(cfg.ReduxNew, cfg.ReduxOld, &wg2)
		wg2.Wait()

		defaultReduxBackupDir := fmt.Sprintf("C:\\Users\\%s\\AppData\\Local\\altv-majestic\\backup", username)
		var wg3 sync.WaitGroup
		log.Printf("Копирование redux: %s -> %s", cfg.ReduxNew, defaultReduxBackupDir)
		copyDirRecursive(cfg.ReduxNew, defaultReduxBackupDir, &wg3)
		wg3.Wait()

		log.Println("Копирование завершено. Ждем 5 секунд перед следующей проверкой...")
		time.Sleep(5 * time.Second)
	}
	// Этот лог никогда не выполнится, так как цикл бесконечный.
	// log.Println("Работа копировального цикла завершена.")
}

func main() {
	initLogger()
	configPath := "config.json"
	var cfg *Config
	var err error

	// Если конфигурация отсутствует – первичный запуск: запрос настроек
	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		log.Println("Конфигурация не найдена. Выполняется первичная настройка.")
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

		cfg.GtaExePath = prompt(fmt.Sprintf("Введите полный путь к GTA5.exe (по умолчанию: %s): ", defaultGTAPathForUser()))
		if cfg.GtaExePath == "" {
			cfg.GtaExePath = defaultGTAPathForUser()
		}

		auto := prompt("Использовать автозапуск при старте ПК? (y/n): ")
		if auto == "y" || auto == "Y" {
			cfg.AutoRun = true
			exePath, _ := os.Executable()
			if err := setAutoRun(true, exePath+" -autostart"); err != nil {
				log.Printf("Ошибка установки автозапуска: %v", err)
			}
		}

		if err := saveConfig(cfg, configPath); err != nil {
			log.Printf("Ошибка сохранения конфигурации: %v", err)
		}

		// После ввода настроек скрываем консоль и запускаем цикл копирования
		hideConsole()
		runCopyingLoop("GTA5.exe", cfg.GtaExePath, cfg, os.Getenv("USERNAME"))
		log.Println("Программа завершена (первичный запуск).")
		return
	} else {
		// Конфигурация существует – загружаем её
		cfg, err = loadConfig(configPath)
		if err != nil {
			log.Printf("Ошибка загрузки конфигурации: %v", err)
			return
		}
	}

	// Определяем режим запуска по аргументу "-autostart"
	isAutoStart := false
	if len(os.Args) > 1 && os.Args[1] == "-autostart" {
		isAutoStart = true
	}

	username := os.Getenv("USERNAME")
	if username == "" {
		username = "DefaultUser"
	}
	if cfg.GtaExePath == "" {
		cfg.GtaExePath = defaultGTAPathForUser()
	}

	if isAutoStart {
		// Автозапуск: сразу скрываем консоль и запускаем цикл копирования
		hideConsole()
		runCopyingLoop("GTA5.exe", cfg.GtaExePath, cfg, username)
		log.Println("Автозапуск: программа завершилась.")
	} else {
		// Ручной запуск: выводим сообщение, запускаем копировальный цикл в горутине,
		// затем через 10 секунд скрываем консоль, а программа продолжает работу в фоне.
		fmt.Println("Программа запущена, все изменения вносите в config.json")
		go runCopyingLoop("GTA5.exe", cfg.GtaExePath, cfg, username)
		time.Sleep(10 * time.Second)
		log.Println("Ручной запуск: скрываем консоль через 10 секунд.")
		hideConsole()
		// Блокируем основной поток, чтобы программа продолжала работу в фоне.
		select {}
	}
}
