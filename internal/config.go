package app

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
)

// Config хранит настройки программы.
type Config struct {
	GunpackNew string `json:"gunpack_new"`
	GunpackOld string `json:"gunpack_old"`
	ReduxNew   string `json:"redux_new"`
	ReduxOld   string `json:"redux_old"`
	GtaExePath string `json:"gta_exe_path"`
	AutoRun    bool   `json:"auto_run"`
}

func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err = json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	log.Println("Конфигурация загружена из", path)
	return &cfg, nil
}

func SaveConfig(cfg *Config, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	if err = os.WriteFile(path, data, 0644); err != nil {
		return err
	}
	log.Println("Конфигурация сохранена в", path)
	return nil
}

// Prompt выводит подсказку и считывает ввод. Если ввод пустой — возвращает defaultVal.
func Prompt(label, defaultVal string) string {
	if defaultVal != "" {
		fmt.Printf("  %s [%s]: ", label, defaultVal)
	} else {
		fmt.Printf("  %s: ", label)
	}
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		val := strings.TrimSpace(scanner.Text())
		if val == "" {
			return defaultVal
		}
		return val
	}
	return defaultVal
}
