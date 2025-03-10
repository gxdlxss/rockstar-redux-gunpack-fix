// processcheck.go
package main

import (
	"bytes"
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
			processName := fields[0]
			processes = append(processes, processName)
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
