package main

import (
	"extractPasswordsForMelanty/internal/script"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

func main() {
	log.Println("PasswordsExtractor by Onlypage")
	dirPath := "./logs"
	files, err := os.ReadDir(dirPath)
	if err != nil {
		os.Mkdir(dirPath, os.ModePerm)
		log.Println("Создана папка /logs, поместите в нее архивы с логами")
		return
	}

	outputFile, err := os.OpenFile("output.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Println(err)
		return
	}
	defer outputFile.Close()

	start := time.Now()
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		archivePath := filepath.Join(dirPath, file.Name())
		passwords, err := script.ReadPasswords(archivePath)
		if err != nil {
			log.Printf("%s - %s", file.Name(), err.Error())
			if err = os.Remove("logs/" + file.Name()); err != nil {
				log.Println(err.Error())
			}
			continue
		}
		if err = os.Remove("logs/" + file.Name()); err != nil {
			log.Println(err.Error())
		}

		for _, password := range passwords {
			if _, err := outputFile.WriteString(password + "\n"); err != nil {
				log.Println(err)
				return
			}
		}
	}
	elapsed := time.Since(start)

	fmt.Printf("Выполнение кода заняло %s\n", elapsed)
	log.Println("скрипт завершил свою работу")

	var input string
	fmt.Println("Нажмите Enter, чтобы завершить работу")
	fmt.Scanln(&input)
}
