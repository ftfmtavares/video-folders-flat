package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const separator = "???"

func validateParams(operation string, folder string, filename string) (bool, string) {
	valid := true
	msg := ""
	if operation == "" || folder == "" || filename == "" {
		msg = "Missing parameters"
		valid = false
	} else if operation != "flat" && operation != "restore" {
		msg = "Operation must be either flat or restore"
		valid = false
	} else if fileInfo, err := os.Stat(folder); err != nil || !fileInfo.IsDir() {
		msg = "Invalid Folder"
		valid = false
	} else if operation == "flat" {
		if fileInfo, err := os.Stat(filename); err == nil || !os.IsNotExist(err) {
			if err != nil && !os.IsNotExist(err) {
				msg = err.Error()
			} else if fileInfo.IsDir() {
				msg = "Invalid file name"
			} else {
				msg = "Target file already exists"
			}
			valid = false
		}
		if valid {
			file, err := os.Create(filename)
			if err != nil {
				msg = err.Error()
				valid = false
			}
			file.Close()
		}
	} else {
		if fileInfo, err := os.Stat(filename); err != nil || fileInfo.IsDir() {
			if err != nil && os.IsNotExist(err) {
				msg = "File does not exist"
			} else if fileInfo.IsDir() {
				msg = "Invalid file name"
			} else {
				msg = "Invalid file name"
			}
			valid = false
		}
	}

	return valid, msg
}

func flat(folder string, filename string, year bool, extension string) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	files := []string{}
	subFolders := []string{}
	err = filepath.Walk(folder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			subFolders = append(subFolders, path)
			return nil
		}

		originalName := info.Name()
		if !strings.HasSuffix(strings.ToLower(originalName), "."+extension) {
			return nil
		}

		fileDir := path[:len(path)-len(originalName)]
		newName := fmt.Sprintf("%d%02d%02d", info.ModTime().Year(), info.ModTime().Month(), info.ModTime().Day()) + "_" + originalName
		line := fileDir + separator + originalName + separator + newName
		files = append(files, line)
		_, err = f.WriteString(line + "\n")
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	for _, line := range files {
		parts := strings.Split(line, separator)
		if year {
			parts[2] = parts[2][:4] + string(os.PathSeparator) + parts[2]
			newSubFolder := folder + parts[2][:4] + string(os.PathSeparator)
			err := os.MkdirAll(newSubFolder, os.ModePerm)
			if err != nil {
				panic(err)
			}
		}
		err := os.Rename(parts[0]+parts[1], folder+parts[2])
		if err != nil {
			panic(err)
		}
	}

	for _, subFolder := range subFolders {
		if subFolder != folder {
			err = os.RemoveAll(subFolder)
			if err != nil {
				panic(err)
			}
		}
	}
}

func restore(folder string, filename string, year bool) {
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

	subFolders := []string{}
	for _, line := range files {
		parts := strings.Split(line, separator)

		err := os.MkdirAll(parts[0], os.ModePerm)
		if err != nil {
			panic(err)
		}

		if year {
			parts[2] = parts[2][:4] + string(os.PathSeparator) + parts[2]
			subFolders = append(subFolders, folder+parts[2][:4]+string(os.PathSeparator))
		}

		err = os.Rename(folder+parts[2], parts[0]+parts[1])
		if err != nil {
			panic(err)
		}
	}

	for _, subFolder := range subFolders {
		err = os.RemoveAll(subFolder)
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
	extension := flag.String("extension", "", "Video file extension")
	filename := flag.String("filename", "", "File name where video files data is stored")
	year := flag.Bool("year", false, "Flat video files by year")
	flag.Parse()

	if !strings.HasSuffix(*folder, string(os.PathSeparator)) {
		*folder = *folder + string(os.PathSeparator)
	}

	if valid, msg := validateParams(*operation, *folder, *filename); !valid {
		fmt.Println(msg)
		fmt.Println("")
		flag.PrintDefaults()
		return
	}

	if *operation == "flat" {
		flat(*folder, *filename, *year, *extension)
	} else {
		restore(*folder, *filename, *year)
	}
}
