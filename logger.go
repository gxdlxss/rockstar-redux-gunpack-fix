// logger.go
package main

import (
	"fmt"
	"log"
	"os"
)

func initLogger() {
	logFile, err := os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Ошибка открытия лог-файла:", err)
		os.Exit(1)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags) // Запись даты и времени
	log.Println("Запуск программы auto-redux-gunpack")
}
