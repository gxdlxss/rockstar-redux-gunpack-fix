package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// getRunningProcesses вызывает tasklist и возвращает список имён процессов.
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

// checkIfProcessRunning проверяет, запущен ли процесс с именем processName (без учёта регистра).
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

// isProcessRunningByPath проверяет, запущен ли процесс по полному пути к исполняемому файлу.
// Используется утилита wmic для поиска процесса по пути.
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
