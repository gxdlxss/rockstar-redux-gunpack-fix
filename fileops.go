// fileops.go
package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		log.Printf("Ошибка открытия файла %s: %v", src, err)
		return err
	}
	defer in.Close()

	// Создаём все необходимые директории
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
		log.Printf("Ошибка копирования из %s в %s: %v", src, dst, err)
		return err
	}

	if err = out.Sync(); err != nil {
		log.Printf("Ошибка синхронизации файла %s: %v", dst, err)
		return err
	}
	log.Printf("Файл скопирован: %s -> %s", src, dst)
	return nil
}

func copyDirRecursive(src, dst string, wg *sync.WaitGroup) {
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		log.Printf("Ошибка чтения директории %s: %v", src, err)
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// Рекурсивное копирование подпапки
			copyDirRecursive(srcPath, dstPath, wg)
		} else {
			wg.Add(1)
			go func(s, d string) {
				defer wg.Done()
				if err := copyFile(s, d); err != nil {
					log.Printf("Ошибка копирования файла %s в %s: %v", s, d, err)
				}
			}(srcPath, dstPath)
		}
	}
}
