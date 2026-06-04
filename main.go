package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	app "github.com/gxdlxss/rockstar-redux-gunpack-fix/internal"
)

const version = "2.1.0"

func defaultGTAPath() string {
	username := os.Getenv("USERNAME")
	if username == "" {
		username = "User"
	}
	return fmt.Sprintf(`C:\Users\%s\AppData\Local\altv-majestic\backup\GTA5.exe`, username)
}

func defaultBackupDir() string {
	username := os.Getenv("USERNAME")
	if username == "" {
		username = "User"
	}
	return fmt.Sprintf(`C:\Users\%s\AppData\Local\altv-majestic\backup`, username)
}

func setup(configPath string) *app.Config {
	fmt.Println()
	fmt.Println("  +------------------------------------------+")
	fmt.Println("  |  Redux & Gunpack Fix  --  Первая настройка  |")
	fmt.Println("  +------------------------------------------+")
	fmt.Println()
	fmt.Println("  Укажите пути к папкам. Enter = использовать значение по умолчанию.")
	fmt.Println()

	cfg := &app.Config{}
	cfg.GunpackNew = app.Prompt("gunpack-new (папка с новыми файлами gunpack)", "")
	cfg.GunpackOld = app.Prompt("gunpack-old (куда копировать gunpack)", "")
	cfg.ReduxNew = app.Prompt("redux-new   (папка с новыми файлами redux)", "")
	cfg.ReduxOld = app.Prompt("redux-old   (куда копировать redux)", "")
	cfg.GtaExePath = app.Prompt("Путь к GTA5.exe", defaultGTAPath())

	fmt.Println()
	auto := app.Prompt("Добавить в автозапуск Windows? (y/n)", "n")
	if auto == "y" || auto == "Y" {
		cfg.AutoRun = true
		exePath, _ := os.Executable()
		if err := app.SetAutoRun(true, exePath+" -autostart"); err != nil {
			fmt.Println("  [!] Ошибка автозапуска:", err)
		} else {
			fmt.Println("  [+] Автозапуск включён.")
		}
	}

	if err := app.SaveConfig(cfg, configPath); err != nil {
		log.Fatalf("Ошибка сохранения конфигурации: %v", err)
	}

	fmt.Println()
	fmt.Println("  [OK] Настройка завершена! Программа уходит в фон...")
	fmt.Println()
	return cfg
}

func runLoop(cfg *app.Config) {
	backupDir := defaultBackupDir()
	log.Println("Ожидание запуска GTA5...")

	type named struct {
		name string
		r    app.CopyResult
	}

	wasRunning := false

	for {
		isRunning := app.IsGTARunning(cfg.GtaExePath)

		// Переход: не запущена → запущена — именно тот момент для копирования
		if !wasRunning && isRunning {
			log.Println("GTA запускается — копируем файлы...")
			start := time.Now()

			ch := make(chan named, 3)
			go func() {
				r := app.CopyDir(cfg.GunpackNew, cfg.GunpackOld)
				ch <- named{"Gunpack", r}
			}()
			go func() {
				r := app.CopyDir(cfg.ReduxNew, cfg.ReduxOld)
				ch <- named{"Redux -> old", r}
			}()
			go func() {
				r := app.CopyDir(cfg.ReduxNew, backupDir)
				ch <- named{"Redux -> backup", r}
			}()

			var totalCopied int32
			for i := 0; i < 3; i++ {
				res := <-ch
				totalCopied += res.r.Copied
				log.Printf("  %-20s скопировано: %d  пропущено: %d  ошибок: %d",
					res.name, res.r.Copied, res.r.Skipped, res.r.Errors)
			}

			elapsed := time.Since(start).Round(time.Millisecond)
			log.Printf("Готово за %v. Ждём завершения GTA5...", elapsed)

			if totalCopied > 0 {
				app.ShowNotification("Redux & Gunpack Fix",
					fmt.Sprintf("Файлы подставлены за %v", elapsed))
			}
		}

		if !isRunning && wasRunning {
			log.Println("GTA закрыта. Ожидание следующего запуска...")
		}

		wasRunning = isRunning

		if isRunning {
			time.Sleep(2 * time.Second) // GTA работает — редкий опрос
		} else {
			time.Sleep(300 * time.Millisecond) // Ждём запуска — частый опрос
		}
	}
}

func main() {
	flagAutostart := flag.Bool("autostart", false, "Запуск в фоне (используется автозапуском Windows)")
	flagConfig := flag.Bool("config", false, "Пересоздать конфигурацию")
	flagVersion := flag.Bool("version", false, "Показать версию программы")
	flag.Parse()

	if *flagVersion {
		fmt.Printf("auto-redux-gunpack v%s\n", version)
		return
	}

	app.InitLogger()
	configPath := "config.json"

	if *flagConfig {
		_ = os.Remove(configPath)
		log.Println("Конфигурация сброшена по флагу --config")
	}

	var cfg *app.Config
	var err error

	if _, err = os.Stat(configPath); os.IsNotExist(err) {
		cfg = setup(configPath)
		app.SwitchToFileOnly()
		app.HideConsole()
		runLoop(cfg)
		return
	}

	cfg, err = app.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("Ошибка загрузки конфигурации: %v", err)
	}
	if cfg.GtaExePath == "" {
		cfg.GtaExePath = defaultGTAPath()
	}

	if *flagAutostart {
		app.SwitchToFileOnly()
		app.HideConsole()
		runLoop(cfg)
		return
	}

	// Ручной запуск: показываем статус 5 секунд, затем уходим в фон
	fmt.Printf("\n  auto-redux-gunpack v%s запущена\n", version)
	fmt.Println("  Файлы будут подставлены в момент запуска GTA5.")
	fmt.Println("  Для перенастройки запустите с флагом: --config")
	fmt.Println("  Окно закроется через 5 секунд...\n")

	go runLoop(cfg)
	time.Sleep(5 * time.Second)
	app.SwitchToFileOnly()
	app.HideConsole()
	select {}
}
