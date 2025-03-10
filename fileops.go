package main

import (
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// copyFile копирует один файл из src в dst, перезаписывая его, если он существует.
func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		log.Printf("Ошибка открытия файла %s: %v", src, err)
		return err
	}
	defer in.Close()

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
		log.Printf("Ошибка копирования %s -> %s: %v", src, dst, err)
		return err
	}
	_ = out.Sync()
	log.Printf("Файл скопирован: %s -> %s", src, dst)
	return nil
}

// copyDirRecursive рекурсивно копирует содержимое директории src в dst.
func copyDirRecursive(src, dst string, wg *sync.WaitGroup) {
	entries, err := ioutil.ReadDir(src)
	if err != nil {
		log.Printf("Ошибка чтения директории %s: %v", src, err)
		return
	}
	for _, e := range entries {
		srcPath := filepath.Join(src, e.Name())
		dstPath := filepath.Join(dst, e.Name())
		if e.IsDir() {
			copyDirRecursive(srcPath, dstPath, wg)
		} else {
			wg.Add(1)
			go func(s, d string) {
				defer wg.Done()
				_ = copyFile(s, d)
			}(srcPath, dstPath)
		}
	}
}
