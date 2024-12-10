package zabbix

import (
	"encoding/json"
	"os"
	"os/exec"
)

func SendData(sender string, config string, key string, logFile string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	cmd := exec.Command(sender, "-c", config, "-k", key, "-o", string(data[:]), "-vv")
	f, err := os.OpenFile(logFile, os.O_RDWR, 0640)
	if err != nil {
		return err
	}
	defer f.Close()

	cmd.Stdout = f
	cmd.Stderr = f

	err = cmd.Run()
	return err
}
