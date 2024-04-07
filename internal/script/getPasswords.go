package script

import (
	"archive/zip"
	"bufio"
	"extractPasswordsForMelanty/internal/errorz"
	"fmt"
	"github.com/nwaples/rardecode"
	"io"
	"log"
	"os"
	"strings"
	"sync"
)

func ReadPasswords(archivePath string) ([]string, error) {
	switch {
	case strings.HasSuffix(archivePath, ".zip"):
		return readFromZip(archivePath)

	case strings.HasSuffix(archivePath, ".rar"):
		return readFromRar(archivePath)
	default:
		return []string{}, errorz.UnsupportedFileType
	}
}

func readFromZip(zipPath string) ([]string, error) {
	passwords := make(chan string)
	errors := make(chan error)

	r, err := zip.OpenReader(zipPath)
	if err != nil {
		return []string{}, err
	}
	defer r.Close()

	var wg sync.WaitGroup
	for _, f := range r.File {
		if strings.HasSuffix(f.Name, "/Passwords.txt") {
			wg.Add(1)
			go func(f *zip.File) {
				defer wg.Done()
				rc, err := f.Open()
				if err != nil {
					errors <- err
					return
				}
				defer rc.Close()

				for _, password := range parsePasswords(rc, f.Name) {
					passwords <- password
				}
			}(f)
		}
	}

	go func() {
		wg.Wait()
		close(passwords)
	}()

	var result []string
	for password := range passwords {
		result = append(result, password)
	}

	select {
	case err := <-errors:
		return []string{}, err
	default:
		return result, nil
	}
}

func readFromRar(rarPath string) ([]string, error) {
	passwords := make(chan string)
	errors := make(chan error)

	f, err := os.Open(rarPath)
	if err != nil {
		return []string{}, err
	}
	defer f.Close()

	r, err := rardecode.NewReader(f, "")
	if err != nil {
		return []string{}, err
	}

	var wg sync.WaitGroup
	for {
		header, err := r.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			errors <- err
			return []string{}, err
		}
		if strings.HasSuffix(header.Name, "/Passwords.txt") {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for _, password := range parsePasswords(r, header.Name) {
					passwords <- password
				}
			}()
		}
	}

	go func() {
		wg.Wait()
		close(passwords)
	}()

	var result []string
	for password := range passwords {
		result = append(result, password)
	}

	select {
	case err := <-errors:
		return []string{}, err
	default:
		return result, nil
	}
}

func parsePasswords(r io.Reader, name string) []string {
	scanner := bufio.NewScanner(r)
	var passwords []string
	var url, username, password string
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "URL: "):
			url = strings.TrimPrefix(line, "URL: ")
		case strings.HasPrefix(line, "Username: "):
			username = strings.TrimPrefix(line, "Username: ")
		case strings.HasPrefix(line, "Password: "):
			password = strings.TrimPrefix(line, "Password: ")
			passwords = append(passwords, fmt.Sprintf("%s:%s:%s", url, username, password))
		}
	}
	if err := scanner.Err(); err != nil {
		log.Printf("Error reading passwords: %v - %s\n", err, name)
		fmt.Printf("При возникновении этой ошибки - удалите папку в которой возникла ошибка (%s)", name)
	}
	return passwords
}
