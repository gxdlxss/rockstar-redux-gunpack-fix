package app

import (
	"log"
	"strings"
	"syscall"
	"unsafe"
)

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procCreateToolhelp32Snapshot = kernel32.NewProc("CreateToolhelp32Snapshot")
	procProcess32FirstW          = kernel32.NewProc("Process32FirstW")
	procProcess32NextW           = kernel32.NewProc("Process32NextW")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
)

const (
	th32csSnapProcess  = 0x00000002
	invalidHandleValue = ^uintptr(0)
)

// processEntry32 — структура PROCESSENTRY32W из Win32 API.
type processEntry32 struct {
	dwSize              uint32
	cntUsage            uint32
	th32ProcessID       uint32
	th32DefaultHeapID   uintptr
	th32ModuleID        uint32
	cntThreads          uint32
	th32ParentProcessID uint32
	pcPriClassBase      int32
	dwFlags             uint32
	szExeFile           [260]uint16
}

// IsGTARunning возвращает true если GTA5.exe запущена.
// Использует Win32 CreateToolhelp32Snapshot — без запуска внешних процессов, мгновенно.
func IsGTARunning(_ string) bool {
	snapshot, _, _ := procCreateToolhelp32Snapshot.Call(th32csSnapProcess, 0)
	if snapshot == invalidHandleValue {
		log.Println("Ошибка снимка процессов")
		return false
	}
	defer procCloseHandle.Call(snapshot)

	var entry processEntry32
	entry.dwSize = uint32(unsafe.Sizeof(entry))

	ret, _, _ := procProcess32FirstW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	for ret != 0 {
		if strings.EqualFold(syscall.UTF16ToString(entry.szExeFile[:]), "GTA5.exe") {
			return true
		}
		ret, _, _ = procProcess32NextW.Call(snapshot, uintptr(unsafe.Pointer(&entry)))
	}
	return false
}
