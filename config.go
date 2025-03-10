package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
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
