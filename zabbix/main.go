package zabbix

import (
	"encoding/json"

	"goarpmon/logger"

	"os/exec"
)

func SendData(logFilePath string, maxLogSizeMb int, sender string, config string, key string, value interface{}) error {
	logger.SetLoggerPath(logFilePath, maxLogSizeMb)

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	cmd := exec.Command(sender, "-c", config, "-k", key, "-o", string(data[:]), "-vv")

	err = cmd.Run()
	return err
}
