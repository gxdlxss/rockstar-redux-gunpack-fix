// main.go
package main

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

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
	// Например, если это GTA5.exe (проверьте в Диспетчере задач).
	// Если в Диспетчере задач он отображается как "GTA5.exe", то пишем так.
	// Если по-другому — укажите нужное имя.
	const gtaProcessName = "GTA5.exe"

	log.Printf("Будем копировать файлы каждые несколько секунд, пока не запустится процесс %s", gtaProcessName)

	for {
		// Проверяем, запущен ли GTA
		if checkIfProcessRunning(gtaProcessName) {
			log.Printf("Процесс %s обнаружен. Останавливаем копирование и выходим.", gtaProcessName)
			break
		}

		log.Println("GTA не запущен. Выполняем копирование...")

		// Копирование для gunpack
		var wg sync.WaitGroup
		log.Printf("Копирование gunpack: %s -> %s", cfg.GunpackNew, cfg.GunpackOld)
		copyDirRecursive(cfg.GunpackNew, cfg.GunpackOld, &wg)
		wg.Wait()

		// Копирование для redux
		var wg2 sync.WaitGroup
		log.Printf("Копирование redux: %s -> %s", cfg.ReduxNew, cfg.ReduxOld)
		copyDirRecursive(cfg.ReduxNew, cfg.ReduxOld, &wg2)
		wg2.Wait()

		log.Println("Копирование завершено. Ждём несколько секунд и проверяем снова.")

		// Интервал между копированиями (например, 5 секунд)
		time.Sleep(5 * time.Second)
	}
	log.Println("Работа программы завершена. Закрываемся.")
	fmt.Println("Работа программы завершена. Подробности смотрите в app.log")
}
