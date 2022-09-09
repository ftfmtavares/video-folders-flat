package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func validateParams(operation string, folder string, filename string) (bool, string) {
	valid := true
	msg := ""
	if operation == "" || folder == "" || filename == "" {
		msg = "Missing parameters"
		valid = false
	} else if operation != "flat" && operation != "restore" {
		msg = "Operation must be either flat or restore"
		valid = false
	} else if !fs.ValidPath(folder) {
		msg = "Invalid folder"
		valid = false
	} else if fileInfo, err := os.Stat(folder); err != nil || !fileInfo.IsDir() {
		msg = "Invalid Folder"
		valid = false
	} else if !fs.ValidPath(filename) {
		msg = "Invalid file name"
		valid = false
	} else if operation == "flat" {
		if fileInfo, err := os.Stat(filename); err == nil || !strings.Contains(err.Error(), "The system cannot find the file specified") {
			if err != nil && !strings.Contains(err.Error(), "The system cannot find the file specified") {
				msg = err.Error()
			} else if fileInfo.IsDir() {
				msg = "Invalid file name"
			} else {
				msg = "Target file already exists"
			}
			valid = false
		}
	} else {
		if fileInfo, err := os.Stat(filename); err != nil || fileInfo.IsDir() {
			if err != nil && strings.Contains(err.Error(), "The system cannot find the file specified") {
				msg = "File does not exist"
			} else {
				msg = "Invalid file name"
			}
			valid = false
		}
	}

	return valid, msg
}

func flat(folder string, filename string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	files := []string{}
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		originalName := info.Name()
		fileDir := path[:len(path)-len(originalName)]
		newName := "" + fmt.Sprintf("%d%02d%02d", info.ModTime().Year(), info.ModTime().Month(), info.ModTime().Day()) + "_" + originalName
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(originalName), ".mp4") {
			line := fileDir + "???" + originalName + "???" + newName
			files = append(files, line)
			_, err := f.WriteString(line + "\n")
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, line := range files {
		parts := strings.Split(line, "???")
		err := os.Rename(parts[0]+parts[1], folder+parts[2])
		if err != nil {
			panic(err)
		}
	}

	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() && path != folder {
			err = os.RemoveAll(path)
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		panic(err)
	}
}

func restore(folder string, filename string) {
	f, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	files := []string{}
	for scanner.Scan() {
		files = append(files, scanner.Text())
	}
	f.Close()

	for _, line := range files {
		parts := strings.Split(line, "???")

		err := os.MkdirAll(parts[0], os.ModePerm)
		if err != nil {
			panic(err)
		}

		err = os.Rename(folder+parts[2], parts[0]+parts[1])
		if err != nil {
			panic(err)
		}
	}

	err = os.Remove(filename)
	if err != nil {
		panic(err)
	}
}

func main() {
	fmt.Println("Flat and Restore Video Folders Tool")
	fmt.Println("")

	operation := flag.String("operation", "", "flat or restore")
	folder := flag.String("folder", "", "Folder where the video files are")
	filename := flag.String("filename", "", "File name where video files data is stored")
	flag.Parse()

	if valid, msg := validateParams(*operation, *folder, *filename); !valid {
		fmt.Println(msg)
		fmt.Println("")
		flag.PrintDefaults()
		return
	}

	if !strings.HasSuffix(*folder, "\\") {
		*folder = *folder + "\\"
	}

	if *operation == "flat" {
		flat(*folder, *filename)
	} else {
		restore(*folder, *filename)
	}
}
