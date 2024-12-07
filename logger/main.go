package logger

import (
	"log"
	"os"

	"github.com/natefinch/lumberjack"
)

func StartLogging(logFilePath string, maxLogFileSizeMb int) (*log.Logger, error) {
	f, err := os.OpenFile(logFilePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	log := log.New(f, "", log.Ldate|log.Ltime)
	log.SetOutput(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    maxLogFileSizeMb,
		MaxBackups: 5,
		MaxAge:     30,
	})
	return log, nil
}

func SetLoggerPath(logFilePath string, MaxFileSizeMb int) {
	log.SetOutput(&lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    MaxFileSizeMb,
		MaxBackups: 5,
		MaxAge:     30,
	})
}
