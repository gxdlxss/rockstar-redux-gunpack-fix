package app

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
)

// CopyResult содержит статистику копирования директории.
type CopyResult struct {
	Copied  int32
	Skipped int32
	Errors  int32
}

// CopyDir рекурсивно копирует содержимое src в dst параллельно.
// Файлы с одинаковым размером, где источник не новее назначения, пропускаются.
func CopyDir(src, dst string) CopyResult {
	var result CopyResult
	var wg sync.WaitGroup
	copyDirRecursive(src, dst, &wg, &result)
	wg.Wait()
	return result
}

func copyDirRecursive(src, dst string, wg *sync.WaitGroup, result *CopyResult) {
	entries, err := os.ReadDir(src)
	if err != nil {
		log.Printf("Ошибка чтения директории %s: %v", src, err)
		atomic.AddInt32(&result.Errors, 1)
		return
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyDirRecursive(srcPath, dstPath, wg, result)
		} else {
			wg.Add(1)
			go func(s, d string) {
				defer wg.Done()
				copyFile(s, d, result)
			}(srcPath, dstPath)
		}
	}
}

func copyFile(src, dst string, result *CopyResult) {
	srcInfo, err := os.Stat(src)
	if err != nil {
		atomic.AddInt32(&result.Errors, 1)
		log.Printf("Ошибка stat %s: %v", src, err)
		return
	}

	// Пропускаем файл если он не изменился (размер совпадает и источник не новее цели)
	if dstInfo, err := os.Stat(dst); err == nil {
		if srcInfo.Size() == dstInfo.Size() && !srcInfo.ModTime().After(dstInfo.ModTime()) {
			atomic.AddInt32(&result.Skipped, 1)
			return
		}
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		atomic.AddInt32(&result.Errors, 1)
		log.Printf("Ошибка создания папки для %s: %v", dst, err)
		return
	}

	in, err := os.Open(src)
	if err != nil {
		atomic.AddInt32(&result.Errors, 1)
		log.Printf("Ошибка открытия %s: %v", src, err)
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		atomic.AddInt32(&result.Errors, 1)
		log.Printf("Ошибка создания %s: %v", dst, err)
		return
	}
	defer out.Close()

	if _, err = io.Copy(out, in); err != nil {
		atomic.AddInt32(&result.Errors, 1)
		log.Printf("Ошибка копирования %s -> %s: %v", src, dst, err)
		return
	}
	_ = out.Sync()
	atomic.AddInt32(&result.Copied, 1)
}
