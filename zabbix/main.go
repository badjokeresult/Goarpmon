package zabbix

import (
	"encoding/json"
	"os"
	"os/exec"
)

func sendData(sender string, config string, key string, value interface{}) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	cmd := exec.Command(sender, "-c", config, "-k", key, "-o", string(data[:]), "-vv")

	currDir := os.Getenv("ARPMONDIR")
	f, err := os.OpenFile(currDir+"/"+"zabbix_sender.log", os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()
	cmd.Stdout = f
	cmd.Stderr = f
	cmd.Dir = currDir

	err = cmd.Run()
	return err
}
