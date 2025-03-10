package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Config хранит пути к директориям, путь к GTA5.exe и флаг автозапуска.
type Config struct {
	GunpackNew string `json:"gunpack_new"`
	GunpackOld string `json:"gunpack_old"`
	ReduxNew   string `json:"redux_new"`
	ReduxOld   string `json:"redux_old"`
	GtaExePath string `json:"gta_exe_path"`
	AutoRun    bool   `json:"auto_run"`
}

// initLogger открывает (или создаёт) лог-файл app.log и настраивает логгер.
func initLogger() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Ошибка открытия лог-файла:", err)
		os.Exit(1)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags)
	log.Println("Запуск программы auto-redux-gunpack")
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

// prompt выводит приглашение и считывает ввод пользователя.
func prompt(message string) string {
	fmt.Print(message)
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		response := strings.TrimSpace(scanner.Text())
		log.Println("Ввод пользователя:", response)
		return response
	}
	return ""
}

// loadConfig читает config.json и возвращает структуру Config.
func loadConfig(path string) (*Config, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}
	log.Println("Конфигурация загружена из", path)
	return &cfg, nil
}

// saveConfig сохраняет структуру Config в config.json.
func saveConfig(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, data, 0644)
	if err != nil {
		return err
	}
	log.Println("Конфигурация сохранена в", path)
	return nil
}

// copyFile копирует один файл из src в dst (перезаписывая его, если он существует).
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		log.Printf("Ошибка открытия файла %s: %v", src, err)
		return err
	}
	defer in.Close()
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		log.Printf("Ошибка создания директорий для %s: %v", dst, err)
		return err
	}
	out, err := os.Create(dst)
	if err != nil {
		log.Printf("Ошибка создания файла %s: %v", dst, err)
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	if err != nil {
		log.Printf("Ошибка копирования %s -> %s: %v", src, dst, err)
		return err
	}
	_ = out.Sync()
	log.Printf("Файл скопирован: %s -> %s", src, dst)
	return nil
}

// copyDirRecursive рекурсивно копирует содержимое директории src в dst.
func copyDirRecursive(src, dst string, wg *sync.WaitGroup) {
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		log.Printf("Ошибка чтения директории %s: %v", src, err)
		return
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyDirRecursive(srcPath, dstPath, wg)
		} else {
			wg.Add(1)
			go func(s, d string) {
				defer wg.Done()
				_ = copyFile(s, d)
			}(srcPath, dstPath)
		}
	}
}

// setAutoRun прописывает или удаляет запись автозапуска через реестр Windows.
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

// getRunningProcesses вызывает tasklist и возвращает список имён процессов (в нижнем регистре).
func getRunningProcesses() ([]string, error) {
	cmd := exec.Command("tasklist")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	lines := strings.Split(out.String(), "\n")
	var processes []string
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) > 0 {
			processes = append(processes, strings.ToLower(fields[0]))
		}
	}
	return processes, nil
}

// checkIfProcessRunning проверяет, запущен ли процесс с именем processName.
func checkIfProcessRunning(processName string) bool {
	procs, err := getRunningProcesses()
	if err != nil {
		log.Println("Ошибка получения списка процессов:", err)
		return false
	}
	for _, p := range procs {
		if strings.EqualFold(p, processName) {
			return true
		}
	}
	return false
}

// isProcessRunningByPath проверяет, запущен ли процесс по полному пути к исполняемому файлу (через wmic).
func isProcessRunningByPath(exePath string) bool {
	escapedPath := strings.ReplaceAll(exePath, `\`, `\\`)
	cmd := exec.Command("wmic", "process", "where", fmt.Sprintf("ExecutablePath='%s'", escapedPath), "get", "ProcessId")
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Printf("Ошибка выполнения wmic: %v", err)
		return false
	}
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line != "" && !strings.EqualFold(line, "ProcessId") {
			return true
		}
	}
	return false
}

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
