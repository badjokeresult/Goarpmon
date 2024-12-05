package logger

import (
	"bytes"
	"compress/gzip"
	"log"
	"os"
	"strconv"
	"strings"
)

func StartLogging(logFilePath string, maxLogFileSize int64) error {
	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	fileStat, err := f.Stat()
	if err != nil {
		return err
	}

	if fileStat.Size() >= maxLogFileSize {
		if err := rotateLogFile(f); err != nil {
			return err
		}
	}

	log.SetOutput(f)
	return nil
}

func rotateLogFile(file *os.File) error {
	filenames, err := os.ReadDir("./")
	if err != nil {
		return err
	}

	fileIndexMax := 0
	for _, name := range filenames {
		fileNameParts := strings.Split(name.Name(), ".")
		fileNameIndex, err := strconv.Atoi(fileNameParts[len(fileNameParts)-2])
		if err != nil {
			return err
		}
		if fileNameIndex > fileIndexMax {
			fileIndexMax = fileNameIndex
		}
	}

	fileName := file.Name() + "." + strconv.Itoa(fileIndexMax) + ".gz"

	err = compressWithGzip(file, fileName)
	return err
}

func compressWithGzip(oldFile *os.File, newFileName string) error {
	var b bytes.Buffer
	content := []byte{}
	w := gzip.NewWriter(&b)
	if _, err := oldFile.Read(content); err != nil {
		return err
	}
	w.Write(content)
	w.Close()
	if err := os.WriteFile(newFileName, content, 0600); err != nil {
		return err
	}

	err := oldFile.Truncate(0)
	return err
}
