package app

import (
	"fmt"
	"io"
	"log"
	"os"
)

var logFile *os.File

// InitLogger настраивает логирование одновременно в файл и в консоль.
func InitLogger() {
	var err error
	logFile, err = os.OpenFile("app.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Ошибка открытия лог-файла:", err)
		os.Exit(1)
	}
	log.SetOutput(io.MultiWriter(logFile, os.Stdout))
	log.SetFlags(log.LstdFlags)
	log.Println("=== Запуск auto-redux-gunpack v2.0 ===")
}

// SwitchToFileOnly переключает логгер только на файл (вызывать перед скрытием консоли).
func SwitchToFileOnly() {
	if logFile != nil {
		log.SetOutput(logFile)
	}
}
