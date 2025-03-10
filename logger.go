package main

import (
	"fmt"
	"log"
	"os"
)

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
